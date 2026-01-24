// Package database provides types and functions for managing the persistence layer of the password manager.
package database

import (
	"errors"

	"github.com/dgraph-io/badger/v4"
)

// Open a new badger database at the specified path.
func Open(path string) (*badger.DB, error) {
	return badger.Open(badger.DefaultOptions(path).WithLoggingLevel(badger.ERROR))
}

func update(db *badger.DB, fn func(txn *badger.Txn) error) error {
	err := db.Update(fn)
	switch {
	case errors.Is(err, badger.ErrDBClosed):
		return ErrClosed
	case err != nil:
		return err
	default:
		return nil
	}
}

func view[T any](db *badger.DB, fn func(txn *badger.Txn) (T, error)) (T, error) {
	var (
		out T
		err error
	)

	err = db.View(func(txn *badger.Txn) error {
		out, err = fn(txn)
		return err
	})
	switch {
	case errors.Is(err, badger.ErrDBClosed):
		return out, ErrClosed
	case err != nil:
		return out, err
	default:
		return out, nil
	}
}
