package main_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"fucku/testhelpers"

	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	testhelpers.SetupPostgres(ctx)

	code := m.Run()
	os.Exit(code)
}

func TestDatabaseConnection(t *testing.T) {
	connStr := "user=postgres password=postgres port=5432 dbname=fucku_test sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()
}
