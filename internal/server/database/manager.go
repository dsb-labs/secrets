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
		Get(id uuid.UUID) (*lifetime.Lifetime[*badger.DB], bool)
		Put(id uuid.UUID, db *lifetime.Lifetime[*badger.DB])
		Range() iter.Seq2[uuid.UUID, *lifetime.Lifetime[*badger.DB]]
		Remove(id uuid.UUID)
	}
)

var (
	// ErrInvalidKey is the error given when attempting to decrypt an individual account database with an invalid
	// key.
	ErrInvalidKey = errors.New("invalid key")
	// ErrDatabaseExists is the error given when attempting to create an individual account database where one already
	// exists.
	ErrDatabaseExists = errors.New("database exists")
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

// Create a database for a given account identifier. This database is closed upon return of this method and is not
// placed into the state.
func (m *Manager) Create(id uuid.UUID, key []byte) error {
	path := filepath.Join(m.dir, id.String())

	if _, err := os.Stat(path); err == nil {
		return ErrDatabaseExists
	}

	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("failed to create user directory %q: %w", path, err)
	}

	opts := badger.DefaultOptions(path).
		WithEncryptionKey(key).
		WithLoggingLevel(badger.ERROR).
		WithIndexCacheSize(100 << 20).
		WithNumVersionsToKeep(1)

	db, err := badger.Open(opts)
	switch {
	case errors.Is(err, badger.ErrEncryptionKeyMismatch):
		return ErrDatabaseExists
	case err != nil:
		return fmt.Errorf("failed to open database: %w", err)
	}

	return db.Close()
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

// RotateKey updates the master encryption key used to encrypt individual user databases. Returns ErrInvalidKey if the
// old encryption key is invalid.
func (m *Manager) RotateKey(id uuid.UUID, oldKey, newKey []byte) error {
	// The path may not exist yet if the user has never logged in before. So we can just do nothing here.
	path := filepath.Join(m.dir, id.String())
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	opt := badger.KeyRegistryOptions{
		Dir:                           path,
		ReadOnly:                      true,
		EncryptionKey:                 oldKey,
		EncryptionKeyRotationDuration: 10 * 24 * time.Hour,
	}

	registry, err := badger.OpenKeyRegistry(opt)
	switch {
	case errors.Is(err, badger.ErrEncryptionKeyMismatch):
		return ErrInvalidKey
	case err != nil:
		return err
	}

	// The badger database must be offline to perform a key rotation. This will effectively force a logout for all
	// devices the user is authenticated on when they change their password. We perform this check after opening
	// the key registry to ensure the old key is correct before we log the user out.
	lt, ok := m.state.Get(id)
	if ok && !lt.Expired() {
		lt.Expire()
	}

	defer registry.Close()

	opt.EncryptionKey = newKey
	if err = badger.WriteKeyRegistry(registry, opt); err != nil {
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
