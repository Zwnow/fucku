package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	config "fucku/internal/config"
	database "fucku/internal/database"
	mailer "fucku/internal/mailer"
	token "fucku/internal/tokens"
	users "fucku/internal/users"
	"fucku/pkg"

	"github.com/joho/godotenv"
)

// We setup the database prior to running the app.
func setupApp(db *database.Database) error {
	// If it already exists, postgres will error accordingly, which we ignore.
	err := database.SetupDatabase(db)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	return nil
}

// Create context and call the actual apps entry point
func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// The actual entry point used to setup the app, register routes and serve the webserver
func run(ctx context.Context, w io.Writer) error {
	// Create slog logger and set it as default
	// First argument is the name of the log file
	logger := pkg.NewLogger("app.log", slog.LevelInfo)
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Load environment
	err := godotenv.Load()
	if err != nil {
		return err
	}

	// Create a database with a connection pool
	db, err := database.NewDatabase(os.Getenv("DB_URL"))
	if err != nil {
		logger.Error("failed to initialize DB", "error", err)
		return err
	}

	// Sets up the database & tables
	err = setupApp(db)
	if err != nil {
		logger.Error("app setup failed", "error", err)
		return err
	}

	// Creates a token service
	tokenService := token.NewTokenService(logger, db)
	appConfig := config.NewAppConfig(logger, db)

	mailer := mailer.NewMailer(logger, appConfig)

	/** WORKERS **/
	go token.StartTokenCleanup(db, logger)
	logger.Info("started token cleanup service")
	go appConfig.StartConfigWorker()
	logger.Info("started config service")

	/** ROUTES & SERVER  **/
	mux := http.NewServeMux()

	// User Routes
	// mux.Handle("POST /register", testMiddleware(te()))
	mux.Handle("POST /register", Chain(
		users.RegisterUser(db, logger, tokenService, mailer),
		RecoveryMiddleware(logger)))

	mux.Handle("POST /login", Chain(
		users.LoginUser(db, logger, tokenService),
		RecoveryMiddleware(logger),
	))

	mux.Handle("POST /logout", Chain(
		users.LogoutUser(db, logger, tokenService),
		IsAuthenticatedMiddleware(db, logger),
		CSRFMiddleware(db, logger),
		RecoveryMiddleware(logger),
	))

	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

	// Run in a go routine to cleanly shutdown in case of failure
	go func() {
		log.Println("Server is running on http://localhost:3000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	// Close connection pool
	if db.DBPool != nil {
		db.DBPool.Close()
	}

	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error("server shutdown error", "error", err)
		return err
	}

	logger.Info("server exited properly")

	return nil
}

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// Middleware that prevents the server from panicing due to errors
func RecoveryMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Server panic", "error", err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func CSRFMiddleware(db *database.Database, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			csrfToken, err := r.Cookie("csrf_token")
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			csrfHeader := r.Header.Get("X-CSRF-Token")
			if csrfHeader == "" || csrfHeader != csrfToken.Value {
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			var expiry time.Time
			row := db.DBPool.QueryRow(ctx,
				`SELECT expires_at FROM tokens WHERE token = $1`, csrfToken.Value)

			if err = row.Scan(&expiry); err != nil {
				logger.Error("failed to parse expiry from db csrf token", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if expiry.Before(time.Now().UTC()) {
				http.Error(w, "CSRF token expired", http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func IsAuthenticatedMiddleware(db *database.Database, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			row := db.DBPool.QueryRow(ctx, `SELECT user_id FROM tokens WHERE token_type = 'session' AND token = $1 AND expires_at > $2`, cookie.Value, time.Now())

			var userId string
			if err := row.Scan(&userId); err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			var u users.User
			row = db.DBPool.QueryRow(ctx, `SELECT id, username, email, verified, created_at, updated_at FROM users WHERE id = $1`, userId)
			if err = row.Scan(&u.Id, &u.Username, &u.Email, &u.Verified, &u.CreatedAt, &u.UpdatedAt); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				logger.Error("failed to parse userdata into struct", "error", err)
				return
			}

			const userKey = users.UserContextKey("user")
			ctx = context.WithValue(r.Context(), userKey, u)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
