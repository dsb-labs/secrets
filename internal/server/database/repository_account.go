package database

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

type (
	// The AccountRepository type is responsible for managing the persistence of individual user accounts. This should
	// be instantiated against the master database, as that is where metadata for accounts is stored. Actual account
	// data, such as passwords etc should be stored within their respective, encrypted user databases.
	AccountRepository struct {
		db *badger.DB
	}

	// The Account type represents a user account as stored in the master database.
	Account struct {
		// The user's unique identifier.
		ID uuid.UUID
		// The user's email address.
		Email string
		// The user's hashed password.
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
	// ErrAccountExists is the error given when performing an operation for an account that conflicts with an existing
	// account record.
	ErrAccountExists = errors.New("account exists")
	// ErrAccountNotFound is the error given when querying an account that does not exist.
	ErrAccountNotFound = errors.New("account not found")
)

// NewAccountRepository returns a new instance of the AccountRepository type that will persist account records using
// the provided badger.DB database.
func NewAccountRepository(db *badger.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Create a new account record. Returns ErrAccountExists if an account already exists with the same email address.
func (r *AccountRepository) Create(account Account) error {
	if _, err := r.FindByEmail(account.Email); err == nil {
		return ErrAccountExists
	}

	data, err := json.Marshal(account)
	if err != nil {
		return err
	}

	return update(r.db, func(txn *badger.Txn) error {
		return txn.Set(account.key(), data)
	})
}

// FindByEmail attempts to return the account record associated with the given email address. Returns ErrAccountNotFound
// if the specified account does not exist.
func (r *AccountRepository) FindByEmail(email string) (Account, error) {
	var account Account
	err := iterate(r.db, "account/", func(acc Account) error {
		if acc.Email == email {
			account = acc
			return errStop
		}

		return nil
	})
	if err != nil {
		return Account{}, err
	}

	if account.ID == uuid.Nil {
		return Account{}, ErrAccountNotFound
	}

	return account, nil
}
