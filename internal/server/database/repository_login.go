package database

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

type (
	// The LoginRepository type is responsible for managing the persistence of user passwords. This should
	// be instantiated against a user's individual database.
	LoginRepository struct {
		db *badger.DB
	}

	// The Login type represents a username & password combination as stored in a user's individual database.
	Login struct {
		// The password's unique identifier.
		ID uuid.UUID
		// The username associated with the password.
		Username string
		// The password.
		Password string
		// The domains where this username and password combination can be used.
		Domains []string
	}
)

func (p Login) key() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("login/")
	buf.Write(p.ID[:])

	return buf.Bytes()
}

// NewLoginRepository returns a new instance of the LoginRepository type that will persist login data using the provided
// badger.DB database.
func NewLoginRepository(db *badger.DB) *LoginRepository {
	return &LoginRepository{db: db}
}

// Create a new login record.
func (r *LoginRepository) Create(login Login) error {
	data, err := json.Marshal(login)
	if err != nil {
		return fmt.Errorf("failed to marshal password %q: %w", login.ID, err)
	}

	return update(r.db, func(txn *badger.Txn) error {
		return txn.Set(login.key(), data)
	})
}

// List all login records.
func (r *LoginRepository) List() ([]Login, error) {
	return view(r.db, func(txn *badger.Txn) ([]Login, error) {
		logins := make([]Login, 0)

		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("login/")

		iter := txn.NewIterator(opts)
		defer iter.Close()

		for iter.Rewind(); iter.Valid(); iter.Next() {
			err := iter.Item().Value(func(value []byte) error {
				var password Login
				if err := json.Unmarshal(value, &password); err != nil {
					return err
				}

				logins = append(logins, password)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}

		return logins, nil
	})
}
