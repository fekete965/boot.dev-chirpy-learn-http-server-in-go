package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/stretchr/testify/require"
)

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to find cwd: %v", err)
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func GetMigrationsDir(t *testing.T) string {
	cwd, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}

	return filepath.Join(cwd, "sql", "schema")
}

func WithTx(t *testing.T, ctx context.Context, db *sql.DB, fn func(*database.Queries) error) {
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			t.Logf("rollback error: %v", err)
		}
	}()

	q := database.New(tx)

	err = fn(q)
	require.NoError(t, err)
}
