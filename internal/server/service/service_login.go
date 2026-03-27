package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/davidsbond/x/convert"
	"github.com/davidsbond/x/filter"
	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/database"
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
		// Delete should remove a login record, returning database.ErrLoginNotFound if it does not exist.
		Delete(uuid.UUID) error
		// Get should return the login record associated with the given id, returning database.ErrLoginNotFound if it
		// does not exist.
		Get(uuid.UUID) (database.Login, error)
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
		// When the login was created.
		CreatedAt time.Time
	}
)

var (
	// ErrLoginNotFound is the error given when trying to perform an operation against a login record that does not
	// exist.
	ErrLoginNotFound = errors.New("login not found")
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
		ID:        login.ID,
		Username:  login.Username,
		Password:  login.Password,
		Domains:   login.Domains,
		CreatedAt: login.CreatedAt,
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
// database's lifetime has expired and the caller must reauthenticate.
func (svc *LoginService) List(userID uuid.UUID, filters ...filter.Filter[Login]) ([]Login, error) {
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

	logins := convert.Slice(results, func(in database.Login) Login {
		return Login{
			ID:        in.ID,
			Username:  in.Username,
			Password:  in.Password,
			Domains:   in.Domains,
			CreatedAt: in.CreatedAt,
		}
	})

	if len(filters) == 0 {
		return logins, nil
	}

	return filter.All(logins, filters...), nil
}

// Delete a login record for the given user. Returns ErrReauthenticate if the underlying individual user database's
// lifetime has expired and the caller must reauthenticate. Returns ErrLoginNotFound if the specified login record does
// not exist.
func (svc *LoginService) Delete(userID uuid.UUID, loginID uuid.UUID) error {
	repo, err := svc.logins.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to get database for user: %w", err)
	}

	err = repo.Delete(loginID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case errors.Is(err, database.ErrLoginNotFound):
		return ErrLoginNotFound
	case err != nil:
		return fmt.Errorf("failed to delete login record: %w", err)
	}

	return nil
}

// Get a login record associated with the given user and login identifiers. Returns ErrReauthenticate if the underlying
// individual user database's lifetime has expired and the caller must reauthenticate. Returns ErrLoginNotFound if the
// specified login record does not exist.
func (svc *LoginService) Get(userID uuid.UUID, loginID uuid.UUID) (Login, error) {
	repo, err := svc.logins.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return Login{}, ErrReauthenticate
	case err != nil:
		return Login{}, fmt.Errorf("failed to get database for user: %w", err)
	}

	result, err := repo.Get(loginID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return Login{}, ErrReauthenticate
	case errors.Is(err, database.ErrLoginNotFound):
		return Login{}, ErrLoginNotFound
	case err != nil:
		return Login{}, fmt.Errorf("failed to get login record: %w", err)
	}

	return Login{
		ID:        result.ID,
		Username:  result.Username,
		Password:  result.Password,
		Domains:   result.Domains,
		CreatedAt: result.CreatedAt,
	}, nil
}
