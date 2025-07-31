package internal_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	token "fucku/internal/tokens"
	database "fucku/internal/database"
	users "fucku/internal/users"
	"fucku/pkg"
)

var (
	db     *database.Database
	logger *slog.Logger
	ts token.TokenService
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
		DB: db,
		Logger: logger,
	}

	code := m.Run()

	db.DBPool.Close()
	os.Exit(code)
}

func TestRegisterUserSuccess(t *testing.T) {
	body := `{"username":"testuser","email":"test@example.com", "password":"1Secret1"}`
	req := httptest.NewRequest("POST", "http://localhost:3000/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	users.RegisterUser(db, logger, ts).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var u users.User
	row := db.DBPool.QueryRow(ctx,
		`SELECT id, username, password, email, created_at, updated_at FROM users WHERE users.username = 'testuser'`)

	if err := row.Scan(&u.Id, &u.Username, &u.Password, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
		t.Fatalf("Error while parsing user: %v", err)
	}

	_, err := db.DBPool.Exec(ctx, `DELETE FROM users WHERE users.username = 'testuser'`)
	if err != nil {
		t.Fatalf("Error cleaning up: %v", err)
	}
}

func TestRegisterUserPasswordTooShort(t *testing.T) {
	body := `{"username":"testuser","email":"test@example.com", "password":"Secret1"}`
	req := httptest.NewRequest("POST", "http://localhost:3000/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	users.RegisterUser(db, logger, ts).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRegisterUserIllegalCharacter(t *testing.T) {
	body := `{"username":"testuser","email":"test@example.com", "password":"Secret1 "}`
	req := httptest.NewRequest("POST", "http://localhost:3000/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	users.RegisterUser(db, logger, ts).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
