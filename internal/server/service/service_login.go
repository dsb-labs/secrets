package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/davidsbond/passwords/internal/server/database"
)

type (
	LoginService struct {
		logins RepositoryProvider[LoginRepository]
	}

	LoginRepository interface {
		Create(database.Login) error
		List() ([]database.Login, error)
	}

	Login struct {
		UserID   uuid.UUID
		Username string
		Password string
		Domains  []string
	}
)

func NewLoginService(passwords RepositoryProvider[LoginRepository]) *LoginService {
	return &LoginService{
		logins: passwords,
	}
}

func (svc *LoginService) Create(login Login) error {
	repo, err := svc.logins.For(login.UserID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to get database for user: %w", err)
	}

	record := database.Login{
		ID:       uuid.New(),
		Username: login.Username,
		Password: login.Password,
		Domains:  login.Domains,
	}

	err = repo.Create(record)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to create login record: %w", err)
	default:
		return nil
	}
}

func (svc *LoginService) List(userID uuid.UUID) ([]Login, error) {
	repo, err := svc.logins.For(userID)
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
		return nil, fmt.Errorf("failed to list login records: %w", err)
	}

	logins := make([]Login, len(results))
	for i, result := range results {
		logins[i] = Login{
			UserID:   userID,
			Username: result.Username,
			Password: result.Password,
			Domains:  result.Domains,
		}
	}

	return logins, nil
}
