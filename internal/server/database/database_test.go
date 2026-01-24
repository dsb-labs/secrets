package database_test

import (
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"
)

func testDB(t *testing.T) *badger.DB {
	t.Helper()

	opts := badger.DefaultOptions("").
		WithInMemory(true).
		WithLoggingLevel(badger.ERROR)

	db, err := badger.Open(opts)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}
