package database_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/davidsbond/x/lifetime"
	"github.com/davidsbond/x/syncmap"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/internal/server/database"
)

func TestManager(t *testing.T) {
	t.Parallel()

	state := syncmap.New[uuid.UUID, *lifetime.Lifetime[*badger.DB]]()
	manager := database.NewManager(t.TempDir(), state, time.Hour)

	t.Cleanup(func() {
		assert.NoError(t, manager.Close())
	})

	key := bytes.Repeat([]byte{0}, 32)

	t.Run("opens new database", func(t *testing.T) {
		err := manager.Unlock(uuid.NameSpaceDNS, key)
		require.NoError(t, err)

		err = manager.Unlock(uuid.NameSpaceURL, key)
		require.NoError(t, err)

		lt, ok := state.Get(uuid.NameSpaceDNS)
		require.True(t, ok)
		require.False(t, lt.Expired())
	})

	t.Run("locks existing database", func(t *testing.T) {
		require.NoError(t, manager.Lock(uuid.NameSpaceDNS))
		lt, ok := state.Get(uuid.NameSpaceDNS)
		require.True(t, ok)
		require.True(t, lt.Expired())
	})

	t.Run("opens existing database with wrong key", func(t *testing.T) {
		badKey := bytes.Repeat([]byte{1}, 32)
		err := manager.Unlock(uuid.NameSpaceDNS, badKey)
		require.Error(t, err)

		lt, ok := state.Get(uuid.NameSpaceDNS)
		require.True(t, ok)
		require.True(t, lt.Expired())
	})

	t.Run("opens existing database with correct key", func(t *testing.T) {
		require.NoError(t, manager.Lock(uuid.NameSpaceDNS))

		err := manager.Unlock(uuid.NameSpaceDNS, key)
		require.NoError(t, err)

		lt, ok := state.Get(uuid.NameSpaceDNS)
		require.True(t, ok)
		require.False(t, lt.Expired())
	})

	t.Run("updates existing database's key", func(t *testing.T) {
		newKey := bytes.Repeat([]byte{1}, 32)
		err := manager.RotateKey(uuid.NameSpaceDNS, key, newKey)
		require.NoError(t, err)

		err = manager.Unlock(uuid.NameSpaceDNS, newKey)
		require.NoError(t, err)
	})

	t.Run("deletes existing database", func(t *testing.T) {
		require.NoError(t, manager.Delete(uuid.NameSpaceDNS))

		lt, ok := state.Get(uuid.NameSpaceDNS)
		require.False(t, ok)
		require.Nil(t, lt)
	})
}
