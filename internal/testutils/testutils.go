package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/config"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testdb"
	"github.com/stretchr/testify/require"
)

type serviceTestHelper struct {
	T       *testing.T
	Ctx     context.Context
	Db      *sql.DB
	Queries *database.Queries
}

func NewTestHelper(t *testing.T) *serviceTestHelper {
	ctx := context.Background()
	migrationsDir := GetMigrationsDir(t)
	db, err := testdb.SetupPostgres(ctx, migrationsDir)
	require.NoError(t, err)

	queries := database.New(db)

	return &serviceTestHelper{
		T:       t,
		Ctx:     ctx,
		Db:      db,
		Queries: queries,
	}
}

func (h *serviceTestHelper) WithTx(fn func(*database.Queries) error) {
	WithTx(h.T, h.Ctx, h.Db, fn)
}

func (h *serviceTestHelper) Cleanup() {
	if h.Db != nil {
		h.Db.Close()
	}
}

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

func rollbackTx(t *testing.T, tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		t.Logf("rollback error: %v", err)
	}
}

func WithTx(t *testing.T, ctx context.Context, db *sql.DB, fn func(*database.Queries) error) {
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	defer rollbackTx(t, tx)

	q := database.New(tx)

	err = fn(q)
	require.NoError(t, err)
}

func GetTestApiConfig() *config.ApiConfig {
	return &config.ApiConfig{
		FileserverHits:     atomic.Int32{},
		JWTSecret:          "test-jwt-secret",
		Platform:           "test-platform",
		PolkaWebhookSecret: "test-polka-webhook-secret",
		Port:               3054,
	}
}
