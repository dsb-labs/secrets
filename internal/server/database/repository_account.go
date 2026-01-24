package database

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

type (
	AccountRepository struct {
		db *badger.DB
	}

	Account struct {
		ID           uuid.UUID
		Email        string
		PasswordHash []byte
	}
)

func (a Account) key() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("account/")
	buf.Write(a.ID[:])

	return buf.Bytes()
}

var (
	ErrAccountExists   = errors.New("account exists")
	ErrAccountNotFound = errors.New("account not found")
)

func NewAccountRepository(db *badger.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(account Account) error {
	return update(r.db, func(txn *badger.Txn) error {
		exists, err := accountExists(txn, account.Email)
		switch {
		case err != nil:
			return err
		case exists:
			return ErrAccountExists
		default:
			return saveAccount(txn, account)
		}
	})
}

func (r *AccountRepository) FindByEmail(email string) (Account, error) {
	return view(r.db, func(txn *badger.Txn) (Account, error) {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("account/")

		iter := txn.NewIterator(opts)
		defer iter.Close()

		var account Account
		for iter.Rewind(); iter.Valid(); iter.Next() {
			err := iter.Item().Value(func(value []byte) error {
				if err := json.Unmarshal(value, &account); err != nil {
					return err
				}

				if account.Email == email {
					iter.Close()
					return nil
				}

				return nil
			})
			if err != nil {
				return Account{}, err
			}
		}

		if account.ID == uuid.Nil {
			return Account{}, ErrAccountNotFound
		}

		return account, nil
	})
}

func accountExists(txn *badger.Txn, email string) (bool, error) {
	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte("account/")

	iter := txn.NewIterator(opts)
	defer iter.Close()

	var exists bool
	for iter.Rewind(); iter.Valid(); iter.Next() {
		err := iter.Item().Value(func(value []byte) error {
			var existing Account
			if err := json.Unmarshal(value, &existing); err != nil {
				return err
			}

			if existing.Email == email {
				exists = true
				iter.Close()
				return nil
			}

			return nil
		})
		if err != nil {
			return false, err
		}
	}

	return exists, nil
}

func saveAccount(txn *badger.Txn, account Account) error {
	data, err := json.Marshal(account)
	if err != nil {
		return err
	}

	return txn.Set(account.key(), data)
}
