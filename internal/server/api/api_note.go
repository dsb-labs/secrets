package api

import (
	"errors"
	"net/http"

	"github.com/davidsbond/x/convert"
	"github.com/davidsbond/x/filter"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
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
		List(uuid.UUID, ...filter.Filter[service.Note]) ([]service.Note, error)
		// Delete should remove the note record associated with the given user and note id. Returning
		// service.ErrNoteNotFound if it does not exist.
		Delete(uuid.UUID, uuid.UUID) error
		// Get should return the note record associated with the given user and note id. Returning
		// service.ErrNoteNotFound if it does not exist.
		Get(uuid.UUID, uuid.UUID) (service.Note, error)
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
	mux.Handle("POST /api/v1/note", requireToken(api.Create))
	mux.Handle("GET /api/v1/note", requireToken(api.List))
	mux.Handle("DELETE /api/v1/note/{id}", requireToken(api.Delete))
	mux.Handle("GET /api/v1/note/{id}", requireToken(api.Get))
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
	CreateNoteResponse struct {
		ID string `json:"id"`
	}
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
	request, err := decode[CreateNoteRequest](r.Body)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to decode request: %v", err)
		return
	}

	note := service.Note{
		ID:      uuid.New(),
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

	write(w, http.StatusCreated, CreateNoteResponse{
		ID: note.ID.String(),
	})
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
	filters := make([]filter.Filter[service.Note], 0)
	if query := r.URL.Query().Get("query"); query != "" {
		filters = append(filters, service.NotesByQuery(query))
	}

	results, err := api.notes.List(tkn.ID(), filters...)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to list notes: %v", err)
		return
	}

	write(w, http.StatusOK, ListNotesResponse{
		Notes: convert.Slice(results, func(in service.Note) Note {
			return Note{
				ID:      in.ID.String(),
				Name:    in.Name,
				Content: in.Content,
			}
		}),
	})
}

type (
	// The DeleteNoteResponse type represents the response body returned when calling NoteAPI.Delete
	DeleteNoteResponse struct{}
)

// Delete handles an inbound HTTP request to delete a note record for a user. On success, it responds with
// an http.StatusOK code and a JSON-encoded ListNotesResponse.
func (api *NoteAPI) Delete(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
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

type (
	// The GetNoteResponse type represents the response body returned when calling NoteAPI.Get
	GetNoteResponse struct {
		// The requested note details.
		Note Note `json:"note"`
	}
)

// Get handles an inbound HTTP request to query a note record for a user. On success, it responds with
// an http.StatusOK code and a JSON-encoded GetNoteResponse.
func (api *NoteAPI) Get(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	noteID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to parse note id: %v", err)
		return
	}

	result, err := api.notes.Get(tkn.ID(), noteID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case errors.Is(err, service.ErrNoteNotFound):
		writeErrorf(w, http.StatusNotFound, "note %q does not exist", noteID)
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to get note: %v", err)
		return
	}

	write(w, http.StatusOK, GetNoteResponse{
		Note: Note{
			ID:      result.ID.String(),
			Name:    result.Name,
			Content: result.Content,
		},
	})
}
