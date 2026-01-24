package database

import (
	"errors"
	"fmt"

	"github.com/davidsbond/x/lifetime"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

type (
	// The RepositoryProvider type is used to conveniently instantiate repositories for individual user databases. Each
	// instance of the RepositoryProvider type is used to instantiate a single repository type via a RepositoryFunc
	// implementation.
	RepositoryProvider[T any] struct {
		state State
		fn    RepositoryFunc[T]
	}

	// The RepositoryFunc type is a function that turns a badger.DB instance to an instance of the parameterized type
	// T.
	RepositoryFunc[T any] func(db *badger.DB) T
)

var (
	// ErrClosed is the error given when calling RepositoryProvider.For with a user identifier that has no open
	// database.
	ErrClosed = errors.New("closed")
)

// NewRepositoryProvider returns a new instance of the RepositoryProvider type that will provide instances of the
// parameterized type T for individual users. Databases are obtained from the provided State implementation.
func NewRepositoryProvider[T any](state State, fn RepositoryFunc[T]) *RepositoryProvider[T] {
	return &RepositoryProvider[T]{
		state: state,
		fn:    fn,
	}
}

// For returns an instance of the parameterized type T using the RepositoryFunc specified when calling New to
// create the RepositoryProvider. The state is checked for an existing, open database associated with the provided
// user identifier. If no database exists, or it has expired, this method returns ErrClosed. Callers must check for
// the ErrClosed error and correctly inform upstream that reauthentication is required.
func (m *RepositoryProvider[T]) For(id uuid.UUID) (T, error) {
	var zero T

	lt, ok := m.state.Get(id)
	if !ok || lt.Expired() {
		return zero, ErrClosed
	}

	db, err := lt.Value()
	switch {
	case errors.Is(err, lifetime.ErrExpired):
		return zero, ErrClosed
	case err != nil:
		return zero, fmt.Errorf("failed to get database for user %q: %w", id, err)
	case db.IsClosed():
		return zero, ErrClosed
	}

	return m.fn(db), nil
}
