package internal

import (
	"context"
	database "fucku/internal/database"
	"log/slog"
	"crypto/rand"
	"math/big"
	"time"
)
const tokenCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type TokenService struct {
	Logger *slog.Logger
	DB *database.Database
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

func (ts *TokenService) NewVerificationToken(userId string) (*Token, error) {
	uniqueToken, err := ts.NewUniqueToken(8)
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
		time.Now().Add(time.Hour * 12))

	
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

func (ts *TokenService) NewUniqueToken(length int) (string, error) {
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
