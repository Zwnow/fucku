package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"fucku/internal"

	"github.com/joho/godotenv"
)

// We setup the database prior to running the app.
func setupApp(db *internal.Database) error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	// If it already exists, postgres will error accordingly, which we ignore.
	err = internal.SetupDatabase(db)
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
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	db, err := internal.NewDatabase(os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("Database setup failed: %v", err)
	}

	err = setupApp(db)
	if err != nil {
		log.Fatalf("App setup failed: %v", err)
	}

	mux := http.NewServeMux()

	// User Routes
	// mux.Handle("POST /register", testMiddleware(te()))
	mux.Handle("POST /register", internal.RegisterUser(db))

	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

	go func() {
		log.Println("Server is running on http://localhost:3000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received")
	// Close connection pool
	if db.DBPool != nil {
		db.DBPool.Close()
	}

	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server exited properly")

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
