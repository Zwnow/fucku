package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
)

type UnregisteredUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Valid    bool
	Reasons  []string
}

func NewUnregisteredUser() UnregisteredUser {
	return UnregisteredUser{
		Username: "",
		Password: "",
		Email:    "",
		Valid:    true,
		Reasons:  make([]string, 0),
	}
}

func (uu *UnregisteredUser) validateUsername() {
	length := len(uu.Username)
	regex := `^[A-Za-z0-9]+$`
	re := regexp.MustCompile(regex)

	if length < 4 || length > 12 {
		uu.Valid = false
		uu.Reasons = append(uu.Reasons, "invalid username length")
	}

	if !re.MatchString(uu.Password) {
		uu.Valid = false
		uu.Reasons = append(uu.Reasons, "invalid characters in password")
	}
}

func (uu *UnregisteredUser) validatePassword() {
	length := len(uu.Password)

	if length < 8 || length > 72 {
		uu.Valid = false
		uu.Reasons = append(uu.Reasons, "invalid password length")
	}

	hasUpper := regexp.MustCompile("[A-Z]")
	hasLower := regexp.MustCompile("[a-z]")
	hasDigit := regexp.MustCompile("[0-9]")

	if !hasUpper.MatchString(uu.Password) || !hasLower.MatchString(uu.Password) || !hasDigit.MatchString(uu.Password) {
		uu.Valid = false
		uu.Reasons = append(uu.Reasons, "password does not meet requirements")
	}
}

func (uu *UnregisteredUser) validateEmail() {
	if !strings.Contains(uu.Email, "@") {
		uu.Valid = false
		uu.Reasons = append(uu.Reasons, "email is invalid")
	}

	if len(uu.Email) < 4 {
		uu.Valid = false
		uu.Reasons = append(uu.Reasons, "email is too short")
	}
}

func (uu *UnregisteredUser) hashPassword() {
	hash, err := argon2id.CreateHash(uu.Password, argon2id.DefaultParams)
	if err != nil {
		uu.Reasons = append(uu.Reasons, err.Error())
		uu.Valid = false
	}

	uu.Password = hash
}

type User struct {
	Id        string
	Email     string
	Username  string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func RegisterUser(db *Database, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uu := NewUnregisteredUser()

		// Parse body
		err := DecodeJSONBody(w, r, &uu)
		if err != nil {
			var mr *malformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.msg, mr.status)
				return
			} else {
				logger.Error("error while decoding json body in register user", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		// Validate fields
		uu.validateUsername()
		uu.validatePassword()
		uu.validateEmail()

		if !uu.Valid {
			logger.Warn("registration input validation failed", "reasons", uu.Reasons, "email", uu.Email)
			http.Error(w, strings.Join(uu.Reasons, "\n"), http.StatusBadRequest)
			return
		}

		// Hash password
		uu.hashPassword()
		if !uu.Valid {
			logger.Warn("failed to hash password", "email", uu.Email)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Insert user
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err = db.DBPool.Exec(ctx,
			`INSERT INTO users (email, username, password) VALUES ($1, $2, $3)`,
			uu.Email, uu.Username, uu.Password)
		if err != nil {
			logger.Error("error while inserting user", "error", err, "email", uu.Email, "username", uu.Username)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(200)
		fmt.Fprintln(w, "User registered successfully")
	})
}
