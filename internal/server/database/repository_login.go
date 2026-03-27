package database

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

type (
	// The LoginRepository type is responsible for managing the persistence of user logins. This should
	// be instantiated against a user's individual database.
	LoginRepository struct {
		db *badger.DB
	}

	// The Login type represents a username & password combination as stored in a user's individual database.
	Login struct {
		// The login's unique identifier.
		ID uuid.UUID
		// The username associated with the login.
		Username string
		// The login.
		Password string
		// The domains where this username and login combination can be used.
		Domains []string
		// When the login was created.
		CreatedAt time.Time
	}
)

var (
	// ErrLoginNotFound is the error given when performing an operation on a login record that does not exist.
	ErrLoginNotFound = errors.New("login not found")
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
		return fmt.Errorf("failed to marshal login %q: %w", login.ID, err)
	}

	return update(r.db, func(txn *badger.Txn) error {
		return txn.Set(login.key(), data)
	})
}

// List all login records.
func (r *LoginRepository) List() ([]Login, error) {
	logins := make([]Login, 0)
	err := iterate(r.db, "login/", func(login Login) error {
		logins = append(logins, login)
		return nil
	})

	slices.SortFunc(logins, func(a, b Login) int {
		return cmp.Compare(a.Username, b.Username)
	})

	return logins, err
}

// Delete a login record, returns ErrLoginNotFound if the login record does not exist.
func (r *LoginRepository) Delete(id uuid.UUID) error {
	return update(r.db, func(txn *badger.Txn) error {
		key := Login{ID: id}.key()

		if _, err := txn.Get(key); errors.Is(err, badger.ErrKeyNotFound) {
			return ErrLoginNotFound
		}

		return txn.Delete(key)
	})
}

// Get a login record by its id, returns ErrLoginNotFound if the login record does not exist.
func (r *LoginRepository) Get(id uuid.UUID) (Login, error) {
	return view(r.db, func(txn *badger.Txn) (Login, error) {
		login := Login{
			ID: id,
		}

		item, err := txn.Get(login.key())
		switch {
		case errors.Is(err, badger.ErrKeyNotFound):
			return Login{}, ErrLoginNotFound
		case err != nil:
			return Login{}, err
		}

		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &login)
		})

		return login, err
	})
}
