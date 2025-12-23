package testdb

import (
	"context"
	"database/sql"
	"testing"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/stretchr/testify/require"
)

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
