package internal

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"log"
	"log/slog"
	"math/big"
	"os"
	"time"

	database "fucku/internal/database"
)

const tokenCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type TokenService struct {
	Logger *slog.Logger
	DB     *database.Database
}

type Token struct {
	Id        string    `json:"id"`
	UserId    string    `json:"user_id"`
	TokenType string    `json:"token_type"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewTokenService(logger *slog.Logger, db *database.Database) *TokenService {
	return &TokenService{
		DB:     db,
		Logger: logger,
	}
}

func (ts *TokenService) NewVerificationToken(userId string) (*Token, error) {
	uniqueToken, err := ts.newUniqueToken(8)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	row := ts.DB.DBPool.QueryRow(ctx, `
		INSERT INTO tokens (user_id, token_type, token, expires_at)
		VALUES ($1, $2, $3, $4) RETURNING id, user_id, token_type, token, expires_at, created_at, updated_at`,
		userId,
		"email_verification",
		uniqueToken,
		time.Now().Add(time.Hour*12))

	var token Token
	if err := row.Scan(
		&token.Id,
		&token.UserId,
		&token.TokenType,
		&token.Token,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &token, nil
}

func (ts *TokenService) NewSessionToken(userId string) (*Token, error) {
	// Check if old token exists and revoke it
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := ts.DB.DBPool.Exec(ctx, `DELETE FROM tokens WHERE token_type = 'session' AND user_id = $1`, userId)
	if err != nil {
		return nil, err
	}

	uniqueToken, err := ts.newUniqueToken(32)
	if err != nil {
		return nil, err
	}

	row := ts.DB.DBPool.QueryRow(ctx, `
		INSERT INTO tokens (user_id, token_type, token, expires_at)
		VALUES ($1, $2, $3, $4) RETURNING id, user_id, token_type, token, expires_at, created_at, updated_at`,
		userId,
		"session",
		uniqueToken,
		time.Now().Add(time.Hour*24))

	var token Token
	if err := row.Scan(
		&token.Id,
		&token.UserId,
		&token.TokenType,
		&token.Token,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.UpdatedAt,
	); err != nil {
		return nil, err
	}

	log.Printf("Created session token: %+v", token)

	return &token, nil
}

func (ts *TokenService) NewCSRFToken(userId string) (*Token, error) {
	// Check if old token exists and revoke it
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := ts.DB.DBPool.Exec(ctx, `DELETE FROM tokens WHERE token_type = 'csrf' AND user_id = $1`, userId)
	if err != nil {
		return nil, err
	}

	t := hmac.New(sha256.New, []byte(os.Getenv("CSRF_SECRET")))

	row := ts.DB.DBPool.QueryRow(ctx, `
		INSERT INTO tokens (user_id, token_type, token, expires_at)
		VALUES ($1, $2, $3, $4) RETURNING id, user_id, token_type, token, expires_at, created_at, updated_at`,
		userId,
		"csrf",
		t,
		time.Now().Add(time.Hour*24))

	var token Token
	if err := row.Scan(
		&token.Id,
		&token.UserId,
		&token.TokenType,
		&token.Token,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &token, nil
}

func (ts *TokenService) newUniqueToken(length int) (string, error) {
	token := make([]byte, length)
	for i := range token {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(tokenCharset))))
		if err != nil {
			return "", err
		}
		token[i] = tokenCharset[num.Int64()]
	}
	return string(token), nil
}

func StartTokenCleanup(db *database.Database, logger *slog.Logger) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		cleanupExpiredTokens(db, logger)
		<-ticker.C
	}
}

func cleanupExpiredTokens(db *database.Database, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := db.DBPool.Exec(ctx, "DELETE FROM tokens WHERE expires_at < $1", time.Now())
	if err != nil {
		logger.Error("error cleaning up tokens", "error", err)
		return
	}

	logger.Info("Cleaned up expired tokens")
}
