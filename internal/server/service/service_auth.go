package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/davidsbond/passwords/internal/server/database"
	"github.com/davidsbond/passwords/internal/server/token"
)

type (
	AuthService struct {
		accounts  AccountRepository
		databases DatabaseManager
		tokens    TokenGenerator
	}

	TokenGenerator interface {
		Generate(uuid.UUID) (token.Token, error)
	}
)

var (
	ErrInvalidPassword = errors.New("invalid password")
)

func NewAuthService(accounts AccountRepository, databases DatabaseManager, tokens TokenGenerator) *AuthService {
	return &AuthService{
		accounts:  accounts,
		databases: databases,
		tokens:    tokens,
	}
}

func (svc *AuthService) Login(email, password string) (token.Token, error) {
	account, err := svc.accounts.FindByEmail(email)
	switch {
	case errors.Is(err, database.ErrAccountNotFound):
		return token.Token{}, ErrAccountNotFound
	case err != nil:
		return token.Token{}, fmt.Errorf("failed to find account %q: %w", email, err)
	}

	if err = bcrypt.CompareHashAndPassword(account.PasswordHash, []byte(password)); err != nil {
		return token.Token{}, ErrInvalidPassword
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

func (svc *AuthService) Logout(userID uuid.UUID) error {
	if err := svc.databases.Lock(userID); err != nil {
		return fmt.Errorf("failed to lock database for user: %w", err)
	}

	return nil
}
