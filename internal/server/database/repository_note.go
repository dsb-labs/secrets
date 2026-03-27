package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

type (
	// The NoteRepository type is responsible for managing the persistence of user notes. This should
	// be instantiated against a user's individual database.
	NoteRepository struct {
		db *badger.DB
	}

	// The Note type represents a secure note as stored in a user's individual database.
	Note struct {
		// The note's unique identifier.
		ID uuid.UUID
		// The note's name.
		Name string
		// The note's contents
		Content string
		// When the note was created.
		CreatedAt time.Time
	}
)

var (
	// ErrNoteNotFound is the error given when performing an operation on a note record that does not exist.
	ErrNoteNotFound = errors.New("note not found")
)

func (p Note) key() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("note/")
	buf.Write(p.ID[:])

	return buf.Bytes()
}

// NewNoteRepository returns a new instance of the NoteRepository type that will persist note data using the provided
// badger.DB database.
func NewNoteRepository(db *badger.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// Create a new note record.
func (r *NoteRepository) Create(note Note) error {
	data, err := json.Marshal(note)
	if err != nil {
		return fmt.Errorf("failed to marshal note %q: %w", note.ID, err)
	}

	return update(r.db, func(txn *badger.Txn) error {
		return txn.Set(note.key(), data)
	})
}

// List all note records.
func (r *NoteRepository) List() ([]Note, error) {
	notes := make([]Note, 0)
	err := iterate(r.db, "note/", func(note Note) error {
		notes = append(notes, note)
		return nil
	})

	return notes, err
}

// Delete a note record, returns ErrNoteNotFound if the note record does not exist.
func (r *NoteRepository) Delete(id uuid.UUID) error {
	return update(r.db, func(txn *badger.Txn) error {
		key := Note{ID: id}.key()

		if _, err := txn.Get(key); errors.Is(err, badger.ErrKeyNotFound) {
			return ErrNoteNotFound
		}

		return txn.Delete(key)
	})
}

// Get a note record by its id, returns ErrNoteNotFound if the note record does not exist.
func (r *NoteRepository) Get(id uuid.UUID) (Note, error) {
	return view(r.db, func(txn *badger.Txn) (Note, error) {
		note := Note{
			ID: id,
		}

		item, err := txn.Get(note.key())
		switch {
		case errors.Is(err, badger.ErrKeyNotFound):
			return Note{}, ErrNoteNotFound
		case err != nil:
			return Note{}, err
		}

		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &note)
		})

		return note, err
	})
}
