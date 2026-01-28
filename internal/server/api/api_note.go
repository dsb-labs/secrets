package api

import (
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"

	"github.com/davidsbond/passwords/internal/server/service"
	"github.com/davidsbond/passwords/internal/server/token"
)

type (
	// The NoteAPI exposes HTTP endpoints for managing individual user notes.
	NoteAPI struct {
		notes NoteService
	}

	// The NoteService interface describes types that manage user notes.
	NoteService interface {
		// Create should create a new note record for the given user id.
		Create(uuid.UUID, service.Note) error
		// List should return all notes associated with the given user id.
		List(uuid.UUID) ([]service.Note, error)
		// Delete should remove the note record associated with the given user and note id. Returning
		// service.ErrNoteNotFound if it does not exist.
		Delete(uuid.UUID, uuid.UUID) error
	}

	// The Note type represents a single password.
	Note struct {
		// The unique identifier of the note.
		ID string `json:"id"`
		// The note's name.
		Name string `json:"name"`
		// The note's contents
		Content string `json:"content"`
	}
)

// NewNoteAPI returns a new instance of the NoteAPI type that manages user notes via the
// given NoteService implementation.
func NewNoteAPI(notes NoteService) *NoteAPI {
	return &NoteAPI{notes: notes}
}

// Register the HTTP endpoints onto the given http.ServeMux.
func (api *NoteAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/note", api.Create)
	mux.HandleFunc("GET /api/v1/note", api.List)
	mux.HandleFunc("DELETE /api/v1/note/{id}", api.Delete)
}

type (
	// The CreateNoteRequest type represents the request body given when calling NoteAPI.Create
	CreateNoteRequest struct {
		// The note's name.
		Name string `json:"name"`
		// The note's contents
		Content string `json:"content"`
	}

	// The CreateNoteResponse type represents the response body returned when calling NoteAPI.Create
	CreateNoteResponse struct{}
)

// Validate the request.
func (r CreateNoteRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required),
		validation.Field(&r.Content, validation.Required),
	)
}

// Create handles an inbound HTTP request to store a new note record for a user. On success, it responds with
// an http.StatusCreated code and a JSON-encoded CreateNoteResponse.
func (api *NoteAPI) Create(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	if !tkn.Valid() {
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	request, err := decode[CreateNoteRequest](r.Body)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to decode request: %v", err)
		return
	}

	note := service.Note{
		Name:    request.Name,
		Content: request.Content,
	}

	err = api.notes.Create(tkn.ID(), note)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to create note: %v", err)
		return
	}

	write(w, http.StatusCreated, CreateNoteResponse{})
}

type (
	// The ListNotesResponse type represents the response body returned when calling NoteAPI.List
	ListNotesResponse struct {
		// The notes stored for the account.
		Notes []Note `json:"notes"`
	}
)

// List handles an inbound HTTP request to list all note records for a user. On success, it responds with
// an http.StatusOK code and a JSON-encoded ListNotesResponse.
func (api *NoteAPI) List(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	if !tkn.Valid() {
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	results, err := api.notes.List(tkn.ID())
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to list notes: %v", err)
		return
	}

	notes := make([]Note, len(results))
	for i, result := range results {
		notes[i] = Note{
			ID:      result.ID.String(),
			Name:    result.Name,
			Content: result.Content,
		}
	}

	write(w, http.StatusOK, ListNotesResponse{Notes: notes})
}

type (
	// The DeleteNoteResponse type represents the response body returned when calling NoteAPI.Delete
	DeleteNoteResponse struct{}
)

// Delete handles an inbound HTTP request to delete a note record for a user. On success, it responds with
// an http.StatusOK code and a JSON-encoded ListNotesResponse.
func (api *NoteAPI) Delete(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	if !tkn.Valid() {
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	noteID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to parse note id: %v", err)
		return
	}

	err = api.notes.Delete(tkn.ID(), noteID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case errors.Is(err, service.ErrNoteNotFound):
		writeErrorf(w, http.StatusNotFound, "note %q does not exist", noteID)
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to delete note: %v", err)
		return
	}

	write(w, http.StatusOK, DeleteNoteResponse{})
}
