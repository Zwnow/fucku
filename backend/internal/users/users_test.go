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

	config "fucku/internal/config"
	database "fucku/internal/database"
	mail "fucku/internal/mailer"
	token "fucku/internal/tokens"
	users "fucku/internal/users"
	"fucku/pkg"
)

var (
	db     *database.Database
	logger *slog.Logger
	ts     *token.TokenService
	mailer *mail.Mailer
	conf   *config.AppConfig
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

	ts = token.NewTokenService(logger, db)
	conf = config.NewAppConfig(logger, db)
	mailer = mail.NewMailer(logger, conf)

	code := m.Run()

	db.DBPool.Close()
	os.Exit(code)
}

func TestRegisterUserSuccess(t *testing.T) {
	body := `{"username":"testuser","email":"test@example.com", "password":"1Secret1"}`
	req := httptest.NewRequest("POST", "http://localhost:3000/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	users.RegisterUser(db, logger, ts, mailer).ServeHTTP(w, req)

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

	users.RegisterUser(db, logger, ts, mailer).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRegisterUserIllegalCharacter(t *testing.T) {
	body := `{"username":"testuser","email":"test@example.com", "password":"Secret1 "}`
	req := httptest.NewRequest("POST", "http://localhost:3000/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	users.RegisterUser(db, logger, ts, mailer).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestLoginUserSuccess(t *testing.T) {
	// Register user
	body := `{"username":"logintest","email":"logintest@example.com", "password":"1Secret1"}`
	registerReq := httptest.NewRequest("POST", "http://localhost:3000/login", bytes.NewBufferString(body))
	registerReq.Header.Set("Content-Type", "application/json")
	registerWriter := httptest.NewRecorder()
	users.RegisterUser(db, logger, ts, mailer).ServeHTTP(registerWriter, registerReq)

	if registerWriter.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", registerWriter.Code)
	}

	// Login
	req := httptest.NewRequest("POST", "http://localhost:3000/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	users.LoginUser(db, logger, ts).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := db.DBPool.Exec(ctx, `DELETE FROM users WHERE email = 'logintest@example.com'`)
	if err != nil {
		t.Log("Failed to cleanup logintest@example.com user")
	}

	t.Logf("%+v", w.Result().Header)
}
