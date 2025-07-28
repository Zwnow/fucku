package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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

func RegisterUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uu := NewUnregisteredUser()

		// Parse body
		err := decodeJSONBody(w, r, &uu)
		if err != nil {
			var mr *malformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.msg, mr.status)
			} else {
				log.Print(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}

		// Validate fields
		uu.validateUsername()
		uu.validatePassword()
		uu.validateEmail()

		if !uu.Valid {
			http.Error(w, strings.Join(uu.Reasons, "\n"), http.StatusBadRequest)
		}

		// Hash password
		uu.hashPassword()
		if !uu.Valid {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

		// Insert user
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err = DBPool.Exec(ctx, `INSERT INTO users (email, username, password) VALUES ($1, $2, $3)`, uu.Email, uu.Username, uu.Password)
		if err != nil {
			log.Printf("%v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		defer cancel()

		w.WriteHeader(200)
	})
}

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048676)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly formed JSON"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &maxBytesError):
			msg := fmt.Sprintf("Request body must not be larger than %d bytes", maxBytesError.Limit)
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}
