package repository

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testRepository *Repository

func TestMain(m *testing.M) {
	ctx := context.Background()
	dbName := helpers.MustGetEnvVar("POSTGRES_DB")
	dbUser := helpers.MustGetEnvVar("POSTGRES_USER")
	dbPassword := helpers.MustGetEnvVar("POSTGRES_PASSWORD")

	postgresCtr, err := postgres.Run(
		ctx,
		"postgres:18-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
		postgres.WithOrderedInitScripts(
			filepath.Join("..", "..", "..", "db", "migrations", "000001_initial.up.sql"),
			filepath.Join("test_config", "test_data.sql"),
		),
		postgres.WithSQLDriver("pgx"),
	)

	cleanup := func() {
		if err := testcontainers.TerminateContainer(postgresCtr); err != nil {
			slog.Error("failed to terminate postgres container", "err", err)
		}
	}

	if err != nil {
		slog.Error("failed to start postgres container", "err", err)
		cleanup()

		return
	}

	connString, err := postgresCtr.ConnectionString(ctx)
	if err != nil {
		slog.Error("failed to get db connection string", "err", err)
		cleanup()

		return
	}

	testRepository, err = New(connString)
	if err != nil {
		slog.Error("failed to initialize repository", "err", err)
		cleanup()

		return
	}

	testRepository.PingDB(context.Background())

	statusCode := m.Run()

	cleanup()

	os.Exit(statusCode)
}
