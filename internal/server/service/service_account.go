package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/davidsbond/keeper/internal/server/database"
)

type (
	// The AccountService type is responsible for managing individual user accounts and their databases.
	AccountService struct {
		accounts  AccountRepository
		databases DatabaseManager
	}

	// The AccountRepository interface describes types that persist account data.
	AccountRepository interface {
		// Create should create a new account record, returning database.ErrAccountExists if an account already
		// exists with the same email address.
		Create(database.Account) error
		// FindByEmail should return the account record associated with the provided email address, returning
		// database.ErrAccountNotFound if an account record cannot be found.
		FindByEmail(string) (database.Account, error)
	}

	// The Account type represents an individual user account.
	Account struct {
		// The user's email address.
		Email string
		// The user's password.
		Password string
		// The user's display name.
		DisplayName string
	}
)

var (
	// ErrAccountExists is the error given when trying to create a new account with an email address that matches
	// an existing account.
	ErrAccountExists = errors.New("account exists")
	// ErrAccountNotFound is the error given when trying to perform an operation on an account that does not exist.
	ErrAccountNotFound = errors.New("account not found")
)

// NewAccountService returns a new instance of the AccountService type that will manage account data via the provided
// AccountRepository implementation and each account's database via the DatabaseManager implementation.
func NewAccountService(accounts AccountRepository, databases DatabaseManager) *AccountService {
	return &AccountService{
		accounts:  accounts,
		databases: databases,
	}
}

// Create a new account, returning its recovery key. Returns ErrAccountExists if an account already exists using the
// same email address. The returned recovery key is derived from the user's unencrypted password and unique identifier.
// It cannot be derived without both and is intended to be used for disaster recovery scenarios where manual decryption
// of the user's database is required.
func (svc *AccountService) Create(account Account) ([]byte, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	record := database.Account{
		ID:           uuid.New(),
		Email:        account.Email,
		PasswordHash: passwordHash,
		DisplayName:  account.DisplayName,
	}

	err = svc.accounts.Create(record)
	switch {
	case errors.Is(err, database.ErrAccountExists):
		return nil, ErrAccountExists
	case err != nil:
		return nil, fmt.Errorf("failed to create account %q: %w", account.Email, err)
	}

	// Since UUIDs are just 16 byte arrays we can use them in place of generating a salt. They'll be unique per
	// account.
	restoreKey := deriveKey(account.Password, record.ID[:])

	return restoreKey, nil
}
