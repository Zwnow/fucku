package internal

import (
	"context"
	"log/slog"
	"time"

	database "fucku/internal/database"

	"github.com/jackc/pgx/v5"
)

type AppConfig struct {
	DB            *database.Database
	Logger        *slog.Logger
	MailingActive bool
	LastFetched   time.Time
}

func NewAppConfig(logger *slog.Logger, db *database.Database) *AppConfig {
	return &AppConfig{
		DB:     db,
		Logger: logger,
	}
}

func (ac *AppConfig) StartConfigWorker() {
	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			row := ac.DB.DBPool.QueryRow(ctx, `SELECT mailing_active FROM config WHERE id = 1;`)
			if err := row.Scan(&ac.MailingActive); err != nil {
				if err == pgx.ErrNoRows {
					ac.DB.DBPool.Exec(ctx, `INSERT INTO config (mailing_active) VALUES (true)`)
				} else {
					ac.Logger.Error("failed to read config from database", "error", err)
				}
			}

			cancel()
			time.Sleep(5 * time.Second)
		}
	}()
}
