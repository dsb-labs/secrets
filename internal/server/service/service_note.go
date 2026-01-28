package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/davidsbond/passwords/internal/server/database"
)

type (
	// The NoteService type responsible for managing individual user note records.
	NoteService struct {
		notes RepositoryProvider[NoteRepository]
	}

	// The NoteRepository interface describes types that persist note records.
	NoteRepository interface {
		// Create should store a new note record.
		Create(database.Note) error
		// List should return all note records.
		List() ([]database.Note, error)
		// Delete should remove a note record, returning database.ErrNoteNotFound if it does not exist.
		Delete(uuid.UUID) error
	}

	// The Note type represents a single user note record.
	Note struct {
		// The unique identifier of the note.
		ID uuid.UUID
		// The note's name.
		Name string
		// The note's contents
		Content string
	}
)

var (
	// ErrNoteNotFound is the error given when trying to perform an operation against a note record that does not
	// exist.
	ErrNoteNotFound = errors.New("note not found")
)

// NewNoteService returns a new instance of the NoteService type that will manage individual user notes using
// NoteRepository implementations provided by the given RepositoryProvider implementation.
func NewNoteService(notes RepositoryProvider[NoteRepository]) *NoteService {
	return &NoteService{
		notes: notes,
	}
}

// Create a new note record for the specified user. Returns ErrReauthenticate if the underlying individual user
// database's lifetime has expired and the caller must reauthenticate.
func (svc *NoteService) Create(userID uuid.UUID, note Note) error {
	repo, err := svc.notes.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to get database for user: %w", err)
	}

	record := database.Note{
		ID:      uuid.New(),
		Name:    note.Name,
		Content: note.Content,
	}

	err = repo.Create(record)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to create note record: %w", err)
	default:
		return nil
	}
}

// List all note records for the specified user. Returns ErrReauthenticate if the underlying individual user
// database's lifetime has expired and the caller must reauthenticate.
func (svc *NoteService) List(userID uuid.UUID) ([]Note, error) {
	repo, err := svc.notes.For(userID)
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
		return nil, fmt.Errorf("failed to list note records: %w", err)
	}

	notes := make([]Note, len(results))
	for i, result := range results {
		notes[i] = Note{
			ID:      result.ID,
			Name:    result.Name,
			Content: result.Content,
		}
	}

	return notes, nil
}

// Delete a note record for the given user. Returns ErrReauthenticate if the underlying individual user database's
// lifetime has expired and the caller must reauthenticate. Returns ErrNoteNotFound if the specified note record does
// not exist.
func (svc *NoteService) Delete(userID uuid.UUID, noteID uuid.UUID) error {
	repo, err := svc.notes.For(userID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case err != nil:
		return fmt.Errorf("failed to get database for user: %w", err)
	}

	err = repo.Delete(noteID)
	switch {
	case errors.Is(err, database.ErrClosed):
		return ErrReauthenticate
	case errors.Is(err, database.ErrNoteNotFound):
		return ErrNoteNotFound
	case err != nil:
		return fmt.Errorf("failed to delete note record: %w", err)
	}

	return nil
}
