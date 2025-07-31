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

	token "fucku/internal/tokens"
	database "fucku/internal/database"
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
	logger := pkg.NewLogger("app.log", slog.LevelInfo)
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		return err
	}

	db, err := database.NewDatabase(os.Getenv("DB_URL"))
	if err != nil {
		logger.Error("failed to initialize DB", "error", err)
		return err
	}

	err = setupApp(db)
	if err != nil {
		logger.Error("app setup failed", "error", err)
		return err
	}

	tokenService := token.TokenService{
		Logger: logger,
		DB: db,
	}

	mux := http.NewServeMux()

	// User Routes
	// mux.Handle("POST /register", testMiddleware(te()))
	mux.Handle("POST /register", users.RegisterUser(db, logger, tokenService))

	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

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

func testMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v", r)

		next.ServeHTTP(w, r)
	})
}

func te() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v", r)
	})
}
