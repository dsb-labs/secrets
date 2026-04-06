package ui

import (
	"errors"
	"net/http"
	"time"

	"github.com/davidsbond/x/filter"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"

	"github.com/dsb-labs/secrets/internal/server/service"
	"github.com/dsb-labs/secrets/internal/server/token"
	noteview "github.com/dsb-labs/secrets/internal/ui/view/note"
	statusview "github.com/dsb-labs/secrets/internal/ui/view/status"
)

type (
	// The NoteHandler type is responsible for serving web interface pages regarding user notes.
	NoteHandler struct {
		accounts AccountService
		notes    NoteService
	}

	// The NoteService interface describes types that manage user note records.
	NoteService interface {
		// Create should store a new note record for the given user.
		Create(accountID uuid.UUID, note service.Note) error
		// List should return all notes associated with the given user id.
		List(accountID uuid.UUID, filters ...filter.Filter[service.Note]) ([]service.Note, error)
		// Get should return the note record associated with the given user and note identifiers.
		Get(accountID uuid.UUID, noteID uuid.UUID) (service.Note, error)
		// Delete should remove the note record associated with the given user and note identifiers.
		Delete(accountID uuid.UUID, noteID uuid.UUID) error
	}
)

// NewNoteHandler returns a new instance of the NoteHandler type that will serve note management UIs using
// the provided service implementations.
func NewNoteHandler(accounts AccountService, notes NoteService) *NoteHandler {
	return &NoteHandler{accounts: accounts, notes: notes}
}

// Register HTTP endpoints onto the provided http.ServeMux.
func (h *NoteHandler) Register(mux *http.ServeMux) {
	mux.Handle("GET /notes", requireToken(h.List))
	mux.Handle("GET /notes/new", requireToken(h.Create))
	mux.Handle("POST /notes", requireToken(h.CreateCallback))
	mux.Handle("GET /notes/{id}", requireToken(h.Detail))
	mux.Handle("POST /notes/{id}/delete", requireToken(h.Delete))
}

// List renders the note list view.
func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	query := r.URL.Query().Get("query")
	filters := make([]filter.Filter[service.Note], 0)
	if query != "" {
		filters = append(filters, service.NotesByQuery(query))
	}

	results, err := h.notes.List(tkn.ID(), filters...)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	items := make([]noteview.Item, len(results))
	for i, n := range results {
		items[i] = noteview.Item{
			ID:   n.ID.String(),
			Name: n.Name,
		}
	}

	render(ctx, w, http.StatusOK, noteview.List, noteview.ViewModel{
		DisplayName: account.DisplayName,
		Notes:       items,
		Query:       query,
	})
}

// Detail renders the note detail view for a single note record.
func (h *NoteHandler) Detail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	noteID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		render(ctx, w, http.StatusNotFound, statusview.NotFound, statusview.NotFoundViewModel{})
		return
	}

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	note, err := h.notes.Get(tkn.ID(), noteID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case errors.Is(err, service.ErrNoteNotFound):
		render(ctx, w, http.StatusNotFound, statusview.NotFound, statusview.NotFoundViewModel{})
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, noteview.Detail, noteview.DetailViewModel{
		DisplayName: account.DisplayName,
		ID:          note.ID.String(),
		Name:        note.Name,
		Content:     note.Content,
		CreatedAt:   note.CreatedAt.Format("2 January 2006 at 15:04"),
	})
}

// Delete handles a note deletion request, redirecting to the note list on success.
func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	noteID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		render(ctx, w, http.StatusNotFound, statusview.NotFound, statusview.NotFoundViewModel{})
		return
	}

	err = h.notes.Delete(tkn.ID(), noteID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case errors.Is(err, service.ErrNoteNotFound):
		render(ctx, w, http.StatusNotFound, statusview.NotFound, statusview.NotFoundViewModel{})
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	redirect(w, r, "/notes")
}

// Create renders the note creation form.
func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, noteview.Create, noteview.CreateViewModel{
		DisplayName: account.DisplayName,
	})
}

// The CreateNoteForm type represents the form values submitted when calling NoteHandler.CreateCallback.
type CreateNoteForm struct {
	// The note's name.
	Name string `form:"name"`
	// The note's contents.
	Content string `form:"content"`
}

// Validate the form.
func (f CreateNoteForm) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Name, validation.Required),
		validation.Field(&f.Content, validation.Required),
	)
}

// CreateCallback handles the note creation form submission, redirecting to the note detail view on success.
func (h *NoteHandler) CreateCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	form, err := decode[CreateNoteForm](r)
	model := noteview.CreateViewModel{
		DisplayName: account.DisplayName,
		Name:        form.Name,
		Content:     form.Content,
	}

	var ve validation.Errors
	switch {
	case errors.As(err, &ve):
		model.Validation.Errors = validationErrors(ve)
		render(ctx, w, http.StatusUnprocessableEntity, noteview.Create, model)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	noteID := uuid.New()
	err = h.notes.Create(tkn.ID(), service.Note{
		ID:        noteID,
		Name:      form.Name,
		Content:   form.Content,
		CreatedAt: time.Now(),
	})
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	redirect(w, r, "/notes/"+noteID.String())
}
