package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/davidsbond/passwords/internal/server/database"
)

type (
	// The LoginService type responsible for managing individual user login records.
	LoginService struct {
		logins RepositoryProvider[LoginRepository]
	}

	// The LoginRepository interface describes types that persist login records.
	LoginRepository interface {
		// Create should store a new login record.
		Create(database.Login) error
		// List should return all login records.
		List() ([]database.Login, error)
	}

	// The Login type represents a single user login record.
	Login struct {
		// The unique identifier of the login.
		ID uuid.UUID
		// The username for the login.
		Username string
		// The password for the login.
		Password string
		// The domains this username/password combination can be used.
		Domains []string
	}
)

// NewLoginService returns a new instance of the LoginService type that will manage individual user logins using
// LoginRepository implementations provided by the given RepositoryProvider implementation.
func NewLoginService(logins RepositoryProvider[LoginRepository]) *LoginService {
	return &LoginService{
		logins: logins,
	}
}

// Create a new login record for the specified user. Returns ErrReauthenticate if the underlying individual user
// database's lifetime has expired and the caller must reauthenticate.
func (svc *LoginService) Create(userID uuid.UUID, login Login) error {
	repo, err := svc.logins.For(userID)
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

// List all login records for the specified user. Returns ErrReauthenticate if the underlying individual user
// // database's lifetime has expired and the caller must reauthenticate.
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
			ID:       result.ID,
			Username: result.Username,
			Password: result.Password,
			Domains:  result.Domains,
		}
	}

	return logins, nil
}
