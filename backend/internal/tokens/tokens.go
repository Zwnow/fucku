package internal

import (
	"time"
)

type Token struct {
	Id        string    `json:"id"`
	UserId    string    `json:"user_id"`
	TokenType string    `json:"token_type"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
