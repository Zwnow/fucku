package internal_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"fucku/internal"
)

func TestMain(m *testing.M) {
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "fucku_dev")
	os.Setenv("DB_PORT", "5432")

	ctx := context.Background()
	internal.InitPool(ctx)
	code := m.Run()
	os.Exit(code)
}

func TestRegisterUser(t *testing.T) {
	body := `{"username":"testuser","email":"test@example.com", "password":"1Secret1"}`
	req := httptest.NewRequest("POST", "http://localhost:3000/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	internal.RegisterUser().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	var u internal.User
	row := internal.DBPool.QueryRow(ctx, `SELECT id, username, password, email, created_at, updated_at FROM users WHERE users.username = 'testuser'`)
	defer cancel()

	if err := row.Scan(&u.Id, &u.Username, &u.Password, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
		t.Fatalf("Error while parsing user: %v", err)
	}

	t.Logf("%+v", u)

	_, err := internal.DBPool.Exec(ctx, `DELETE FROM users WHERE users.username = 'testuser'`)
	if err != nil {
		t.Fatalf("Error cleaning up: %v", err)
	}
}
