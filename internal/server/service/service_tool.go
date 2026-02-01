package service

import (
	"errors"
	"fmt"

	"github.com/davidsbond/x/convert"
	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/database"
)

type (
	// The ToolService type provides common user tool implementations, like data export/import.
	ToolService struct {
		logins RepositoryProvider[LoginRepository]
		notes  RepositoryProvider[NoteRepository]
	}

	// The Export type represents a user's entire dataset.
	Export struct {
		// The user's logins.
		Logins []Login
		// The user's notes.
		Notes []Note
	}
)

// NewToolService returns a new instance of the ToolService type that will query logins and notes from the given
// repository provider implementations.
func NewToolService(logins RepositoryProvider[LoginRepository], notes RepositoryProvider[NoteRepository]) *ToolService {
	return &ToolService{
		notes:  notes,
		logins: logins,
	}
}

// Export the specified user's data.  Returns ErrReauthenticate if the underlying individual user database's lifetime
// has expired and the caller must reauthenticate.
func (svc *ToolService) Export(userID uuid.UUID) (Export, error) {
	loginRepo, err := svc.logins.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return Export{}, ErrReauthenticate
	case err != nil:
		return Export{}, fmt.Errorf("failed to get database for user: %w", err)
	}

	noteRepo, err := svc.notes.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return Export{}, ErrReauthenticate
	case err != nil:
		return Export{}, fmt.Errorf("failed to get database for user: %w", err)
	}

	logins, err := loginRepo.List()
	switch {
	case errors.Is(err, database.ErrClosed):
		return Export{}, ErrReauthenticate
	case err != nil:
		return Export{}, fmt.Errorf("failed to list logins: %w", err)
	}

	notes, err := noteRepo.List()
	switch {
	case errors.Is(err, database.ErrClosed):
		return Export{}, ErrReauthenticate
	case err != nil:
		return Export{}, fmt.Errorf("failed to list notes: %w", err)
	}

	return Export{
		Notes: convert.Slice(notes, func(in database.Note) Note {
			return Note{
				ID:      in.ID,
				Name:    in.Name,
				Content: in.Content,
			}
		}),
		Logins: convert.Slice(logins, func(in database.Login) Login {
			return Login{
				ID:       in.ID,
				Username: in.Username,
				Password: in.Password,
				Domains:  in.Domains,
			}
		}),
	}, nil
}
