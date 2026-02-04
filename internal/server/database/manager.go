package database

import (
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"time"

	"github.com/davidsbond/x/lifetime"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

type (
	// The Manager type is responsible for managing individual user's encrypted databases.
	Manager struct {
		dir        string
		state      State
		expiration time.Duration
	}

	// The State interface describes types that store references to individual badgerdb instances wrapped within
	// a lifetime.
	State interface {
		Get(uuid.UUID) (*lifetime.Lifetime[*badger.DB], bool)
		Put(uuid.UUID, *lifetime.Lifetime[*badger.DB])
		Range() iter.Seq2[uuid.UUID, *lifetime.Lifetime[*badger.DB]]
		Remove(uuid.UUID)
	}
)

// NewManager returns a new instance of the Manager type that will manage individual user databases within the specified
// directory. Each database is wrapped within a lifetime causing it to be automatically closed after a specified
// expiration time.
func NewManager(dir string, state State, expiration time.Duration) *Manager {
	return &Manager{
		dir:        dir,
		state:      state,
		expiration: expiration,
	}
}

// Unlock opens a user's encrypted database using their encryption key. Each database remains open for the amount
// of time specified when calling New to create the Manager. If this method is called for a user whose database is
// currently open, the expiration is then reset to its original value.
func (m *Manager) Unlock(id uuid.UUID, key []byte) error {
	// If the specified user's database is already open (likely from a login on another device), we'll just reset
	// the timer on the lifetime.
	lt, ok := m.state.Get(id)
	if ok && !lt.Expired() {
		return lt.Reset(m.expiration)
	}

	// Otherwise, we ensure the path to the user's directory exists.
	path := filepath.Join(m.dir, id.String())
	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("failed to create user directory %q: %w", path, err)
	}

	// And set their database up with the specified encryption key.
	opts := badger.DefaultOptions(path).
		WithEncryptionKey(key).
		WithLoggingLevel(badger.ERROR).
		WithIndexCacheSize(100 << 20).
		WithNumVersionsToKeep(1)

	db, err := badger.Open(opts)
	if err != nil {
		return fmt.Errorf("failed to open database at %q: %w", path, err)
	}

	lt = lifetime.New(db, m.expiration)
	m.state.Put(id, lt)
	return nil
}

// Lock immediately closes a user's encrypted database if it is open.
func (m *Manager) Lock(id uuid.UUID) error {
	lt, ok := m.state.Get(id)
	if !ok || lt.Expired() {
		return nil
	}

	lt.Expire()
	return nil
}

// Delete a user's encrypted database.
func (m *Manager) Delete(id uuid.UUID) error {
	if err := m.Lock(id); err != nil {
		return err
	}

	m.state.Remove(id)
	path := filepath.Join(m.dir, id.String())
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return nil
}

// Close all open user databases.
func (m *Manager) Close() error {
	errs := make([]error, 0)
	for _, lt := range m.state.Range() {
		if lt.Expired() {
			continue
		}

		db, err := lt.Value()
		switch {
		case errors.Is(err, lifetime.ErrExpired):
			continue
		case err != nil:
			errs = append(errs, err)
			continue
		}

		if err = db.Close(); err != nil {
			errs = append(errs, err)
		}

		lt.Expire()
	}

	return errors.Join(errs...)
}
