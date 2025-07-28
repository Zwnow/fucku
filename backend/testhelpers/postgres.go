package testhelpers

import (
	"context"
	"log"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupPostgres(ctx context.Context) {
	dbName := "fucku_test"
	dbUser := "postgres"
	dbPassword := "postgres"

	container, err := postgres.Run(ctx,
		"postgres:latest",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	testcontainers.WithAdditionalWaitStrategy(
		wait.ForLog("database system is ready to accept connections"),
		wait.ForListeningPort("5432/tcp"),
	)
}
