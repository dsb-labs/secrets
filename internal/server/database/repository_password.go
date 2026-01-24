package database

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

type (
	PasswordRepository struct {
		db *badger.DB
	}

	Password struct {
		ID       uuid.UUID
		Username string
		Password string
		Domains  []string
	}
)

func (p Password) key() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("password/")
	buf.Write(p.ID[:])

	return buf.Bytes()
}

func NewPasswordRepository(db *badger.DB) *PasswordRepository {
	return &PasswordRepository{db: db}
}

func (r *PasswordRepository) Create(password Password) error {
	data, err := json.Marshal(password)
	if err != nil {
		return fmt.Errorf("failed to marshal password %q: %w", password.ID, err)
	}

	return update(r.db, func(txn *badger.Txn) error {
		return txn.Set(password.key(), data)
	})
}

func (r *PasswordRepository) List() ([]Password, error) {
	return view(r.db, func(txn *badger.Txn) ([]Password, error) {
		passwords := make([]Password, 0)

		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("password/")

		iter := txn.NewIterator(opts)
		defer iter.Close()

		for iter.Rewind(); iter.Valid(); iter.Next() {
			err := iter.Item().Value(func(value []byte) error {
				var password Password
				if err := json.Unmarshal(value, &password); err != nil {
					return err
				}

				passwords = append(passwords, password)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}

		return passwords, nil
	})
}
