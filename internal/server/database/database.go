// Package database provides types and functions for managing the persistence layer of the password manager.
package database

import (
	"encoding/json"
	"errors"

	"github.com/dgraph-io/badger/v4"
)

// Open a new badger database at the specified path.
func Open(path string) (*badger.DB, error) {
	options := badger.DefaultOptions(path).
		WithLoggingLevel(badger.ERROR).
		WithNumVersionsToKeep(1)

	return badger.Open(options)
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

var (
	errStop = errors.New("stop")
)

func iterate[T any](db *badger.DB, prefix string, fn func(T) error) error {
	return db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefix)

		iter := txn.NewIterator(opts)
		defer iter.Close()

		for iter.Rewind(); iter.Valid(); iter.Next() {
			err := iter.Item().Value(func(value []byte) error {
				var item T
				if err := json.Unmarshal(value, &item); err != nil {
					return err
				}

				err := fn(item)
				switch {
				case errors.Is(err, errStop):
					iter.Close()
					return nil
				default:
					return err
				}
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}
