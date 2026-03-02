package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/davidsbond/keeper/internal/server/database"
	"github.com/davidsbond/keeper/internal/server/token"
)

type (
	// The AuthService type is responsible for managing user authentication.
	AuthService struct {
		accounts  AccountRepository
		databases DatabaseManager
		tokens    TokenGenerator
	}

	// The TokenGenerator interface describes types that can generate authentication tokens.
	TokenGenerator interface {
		// Generate should return a new token.Token for the given user identifier.
		Generate(uuid.UUID) (token.Token, error)
	}
)

var (
	// ErrInvalidPassword is the error returned when a user attempts to authenticate with an incorrect password.
	ErrInvalidPassword = errors.New("invalid password")
)

// NewAuthService returns a new instance of the AuthService type that queries account data via the given AccountRepository
// implementation, manages individual user databases via the given DatabaseManager implementation and generates
// authentication tokens using the given TokenGenerator implementation.
func NewAuthService(accounts AccountRepository, databases DatabaseManager, tokens TokenGenerator) *AuthService {
	return &AuthService{
		accounts:  accounts,
		databases: databases,
		tokens:    tokens,
	}
}

// Login attempts to generate a new token.Token for the user matching the given email and password combination. Returns
// ErrAccountNotFound if the given email address does not match an existing account, or ErrInvalidPassword if the
// provided password does not match that of the specified user.
func (svc *AuthService) Login(email, password string) (token.Token, error) {
	account, err := svc.accounts.FindByEmail(email)
	switch {
	case errors.Is(err, database.ErrAccountNotFound):
		return token.Token{}, ErrAccountNotFound
	case err != nil:
		return token.Token{}, fmt.Errorf("failed to find account %q: %w", email, err)
	}

	err = bcrypt.CompareHashAndPassword(account.PasswordHash, []byte(password))
	switch {
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return token.Token{}, ErrInvalidPassword
	case err != nil:
		return token.Token{}, fmt.Errorf("failed to compare password: %w", err)
	}

	key := deriveKey(password, account.ID[:])

	// When the user logs in, we need to either open and decrypt their personal database or extend the
	// lifetime before we auto lock it back down.
	if err = svc.databases.Unlock(account.ID, key); err != nil {
		return token.Token{}, fmt.Errorf("failed to open database for account %q: %w", account.Email, err)
	}

	tkn, err := svc.tokens.Generate(account.ID)
	if err != nil {
		return token.Token{}, fmt.Errorf("failed to generate token for account %q: %w", account.Email, err)
	}

	return tkn, nil
}

// Logout locks the specified user's database.
func (svc *AuthService) Logout(userID uuid.UUID) error {
	if err := svc.databases.Lock(userID); err != nil {
		return fmt.Errorf("failed to lock database for user: %w", err)
	}

	return nil
}
