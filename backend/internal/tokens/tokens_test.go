package internal_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	database "fucku/internal/database"
	token "fucku/internal/tokens"
	"fucku/pkg"
	"github.com/jackc/pgx/v5"
)

var (
	db     *database.Database
	logger *slog.Logger
	ts     token.TokenService
)

func TestMain(m *testing.M) {
	os.Setenv("DB_URL", "postgresql://postgres:postgres@localhost:5432/fucku_dev")

	logger = pkg.NewLogger("tests.log", slog.LevelInfo)

	var err error
	db, err = database.NewDatabase(os.Getenv("DB_URL"))
	if err != nil {
		logger.Error("failed to set up database", "error", err)
		return
	}

	ts = token.TokenService{
		DB:     db,
		Logger: logger,
	}

	code := m.Run()

	db.DBPool.Close()
	os.Exit(code)
}

func TestSessionToken(t *testing.T) {
	token, err := ts.NewSessionToken("4c8fc246-38a3-4605-8b1e-f42544e008b6")
	if err != nil {
		t.Fatalf("error during TestSessionToken: %v", err)
	}

	if len(token.Token) != 32 {
		t.Errorf("expected token length 32, got: %d", len(token.Token))
	}

	if token.UserId != "4c8fc246-38a3-4605-8b1e-f42544e008b6" {
		t.Errorf("expected token user id 4c8fc246-38a3-4605-8b1e-f42544e008b6, got: %s", token.UserId)
	}

	if err = cleanupToken(token); err != nil {
		t.Fatalf("failed to cleanup token: %v", err)
	}
}

func TestSessionRevoke(t *testing.T) {
	tokenOne, err := ts.NewSessionToken("4c8fc246-38a3-4605-8b1e-f42544e008b6")
	if err != nil {
		t.Fatalf("error during TestSessionToken (token one): %v", err)
	}

	tokenTwo, err := ts.NewSessionToken("4c8fc246-38a3-4605-8b1e-f42544e008b6")
	if err != nil {
		t.Fatalf("error during TestSessionToken (token two): %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var tokenOneId string
	row := db.DBPool.QueryRow(ctx, `SELECT id FROM tokens WHERE id = $1`, tokenOne.Id)
	if err := row.Scan(&tokenOneId); err != nil && err != pgx.ErrNoRows {
		t.Fatalf("unexpected error (token one): %v", err)
	}

	if len(tokenOneId) != 0 {
		t.Errorf("expected tokenOneId length 0, got: %d", len(tokenOneId))
	}

	var tokenTwoId string
	row = db.DBPool.QueryRow(ctx, `SELECT id FROM tokens WHERE id = $1`, tokenTwo.Id)
	if err := row.Scan(&tokenTwoId); err != nil {
		t.Fatalf("unexpected error (token two): %v", err)
	}

	cleanupToken(tokenOne)
	cleanupToken(tokenTwo)
}

func cleanupToken(token *token.Token) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := db.DBPool.Exec(ctx, `DELETE FROM tokens WHERE id = $1`, token.Id)
	if err != nil {
		return err
	}

	return nil
}
