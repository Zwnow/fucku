package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	config "fucku/internal/config"
)

type Mailer struct {
	AppConfig *config.AppConfig
	Logger    *slog.Logger
	key       string
	secret    string
}

func NewMailer(logger *slog.Logger, ac *config.AppConfig) *Mailer {
	key := os.Getenv("MAILJET_KEY")
	secret := os.Getenv("MAILJET_SECRET")

	return &Mailer{
		AppConfig: ac,
		Logger:    logger,
		key:       key,
		secret:    secret,
	}
}

func (m *Mailer) SendRegistrationMail(username, email, token string) {
	if !m.AppConfig.MailingActive {
		return
	}

	appName := os.Getenv("APP_NAME")

	body := map[string]any{
		"Messages": []map[string]any{
			{
				"From": map[string]any{
					"Email": "svenotimm@gmail.com",
					"Name":  fmt.Sprintf("%s - Email Verification", appName),
				},
				"To": []map[string]any{
					{
						"Email": email,
						"Name":  username,
					},
				},
				"Subject":  "Your Email-Verification Code",
				"HTMLPart": fmt.Sprintf("<h3>Hey %s, welcome to %s!</h3><br />Your verification code is: %s!", username, appName, token),
				"TextPart": fmt.Sprintf("Hey %s, welcome to %s! Your verification code is: %s!", username, appName, token),
			},
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		m.Logger.Error("failed to marshal email JSON", "error", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.mailjet.com/v3.1/send", bytes.NewBuffer(bodyBytes))
	if err != nil {
		m.Logger.Error("failed to create email request", "error", err, "email", email)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(m.key, m.secret)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		m.Logger.Error("failed to send email", "error", err, "email", email)
		return
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		m.Logger.Error("failed to read response body", "error", err)
		return
	}

	m.Logger.Info("response body", "info", resBody)
}
