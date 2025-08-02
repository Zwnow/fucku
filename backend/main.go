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

	database "fucku/internal/database"
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
	tokenService := token.TokenService{
		Logger: logger,
		DB:     db,
	}

	/** WORKERS **/
	go token.StartTokenCleanup(db, logger)
	logger.Info("started token cleanup service")

	/** ROUTES & SERVER  **/
	mux := http.NewServeMux()

	// User Routes
	// mux.Handle("POST /register", testMiddleware(te()))
	mux.Handle("POST /register", Chain(
		users.RegisterUser(db, logger, tokenService),
		logger,
		RecoveryMiddleware(logger)))

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

func Chain(h http.Handler, logger *slog.Logger, middlewares ...Middleware) http.Handler {
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
