package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/davidsbond/passwords/internal/server/database"
)

type (
	PasswordService struct {
		passwords RepositoryProvider[PasswordRepository]
	}

	PasswordRepository interface {
		Create(database.Password) error
		List() ([]database.Password, error)
	}

	Password struct {
		UserID   uuid.UUID
		Username string
		Password string
		Domains  []string
	}
)

func NewPasswordService(passwords RepositoryProvider[PasswordRepository]) *PasswordService {
	return &PasswordService{
		passwords: passwords,
	}
}

func (svc *PasswordService) Create(password Password) error {
	repo, err := svc.passwords.For(password.UserID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to get database for user: %w", err)
	}

	record := database.Password{
		ID:       uuid.New(),
		Username: password.Username,
		Password: password.Password,
		Domains:  password.Domains,
	}

	err = repo.Create(record)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to create password record: %w", err)
	default:
		return nil
	}
}

func (svc *PasswordService) List(userID uuid.UUID) ([]Password, error) {
	repo, err := svc.passwords.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return nil, ErrReauthenticate
	case err != nil:
		return nil, fmt.Errorf("failed to get database for user: %w", err)
	}

	results, err := repo.List()
	switch {
	case errors.Is(err, database.ErrClosed):
		return nil, ErrReauthenticate
	case err != nil:
		return nil, fmt.Errorf("failed to list password records: %w", err)
	}

	passwords := make([]Password, len(results))
	for i, result := range results {
		passwords[i] = Password{
			UserID:   userID,
			Username: result.Username,
			Password: result.Password,
			Domains:  result.Domains,
		}
	}

	return passwords, nil
}
