package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	db       *sql.DB
	initOnce sync.Once
)

func SetupPostgres(ctx context.Context, migrationsDir string) (*sql.DB, error) {
	var setupError error

	initOnce.Do(func() {
		pg, err := postgres.Run(
			ctx,
			"postgres:16-alpine",
			postgres.WithDatabase("testdb"),
			postgres.WithUsername("test"),
			postgres.WithPassword("text"),
			postgres.BasicWaitStrategies(),
		)
		if err != nil {
			setupError = fmt.Errorf("failed to run postgres container: %v", err)
			return
		}

		connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			setupError = fmt.Errorf("failed to get connection string: %v", err)
			return
		}

		db, err = sql.Open("postgres", connStr)
		if err != nil {
			setupError = fmt.Errorf("failed to open database: %v", err)
			return
		}

		err = goose.SetDialect("postgres")
		if err != nil {
			setupError = fmt.Errorf("failed to set dialect: %v", err)
			return
		}

		err = goose.Up(db, migrationsDir)
		if err != nil {
			setupError = fmt.Errorf("failed to run migrations: %v", err)
			return
		}
	})

	return db, setupError
}
