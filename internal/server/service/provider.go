package service

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"

	"github.com/davidsbond/passwords/internal/server/database"
)

type (
	// The RepositoryProvider interface describes types that return instances of the T type parameter scoped for a
	// particular identifier. This is intended to be a convenient way to access individual user databases while wrapping
	// them in a repository implementation of some kind without leaking the underlying implementation detail.
	RepositoryProvider[T any] interface {
		// For should return an instance of T associated with the given identifier. T would typically be some kind of
		// repository implementation backed by some persistent state. For should return database.ErrClosed if the
		// specified identifier is not associated with an open database. This should be used as a signal that the
		// user must reauthenticate to "unlock" their encrypted password data, typically achieved by bubbling
		// ErrReauthenticate upstream.
		For(uuid.UUID) (T, error)
	}
)

// LoginRepositoryProvider is a RepositoryProvider implementation that returns a LoginRepository implementation
// backed by a badger database.
func LoginRepositoryProvider(db *badger.DB) LoginRepository {
	return database.NewLoginRepository(db)
}

// NoteRepositoryProvider is a RepositoryProvider implementation that returns a NoteRepository implementation
// backed by a badger database.
func NoteRepositoryProvider(db *badger.DB) NoteRepository {
	return database.NewNoteRepository(db)
}
