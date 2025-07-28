package config_test

import (
	"context"
	"os"
	"testing"

	"fucku/testhelpers"
	"github.com/jackc/pgx"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	testhelpers.SetupPostgres(ctx)

	code := m.Run()
	os.Exit(code)
}

func TestDatabaseConnection(t *testing.T) {
	config := pgx.ConnConfig{
		User:     "postgres",
		Password: "postgres",
		Database: "postgres",
		Port:     5432,
		Host:     "localhost",
	}
	conn, err := pgx.Connect(config)
	if err != nil {
		t.Fatalf("Failed to open database connection: %v", err)
	}
	defer conn.Close()
}
