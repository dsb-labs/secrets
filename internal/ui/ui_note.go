package ui

import (
	"errors"
	"net/http"

	"github.com/davidsbond/x/filter"
	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
	"github.com/davidsbond/keeper/internal/ui/component"
	noteview "github.com/davidsbond/keeper/internal/ui/view/note"
)

type (
	// The NoteHandler type is responsible for serving web interface pages regarding user notes.
	NoteHandler struct {
		accounts AccountService
		notes    NoteService
	}

	// The NoteService interface describes types that manage user note records.
	NoteService interface {
		// List should return all notes associated with the given user id.
		List(accountID uuid.UUID, filters ...filter.Filter[service.Note]) ([]service.Note, error)
		// Get should return the note record associated with the given user and note identifiers.
		Get(accountID uuid.UUID, noteID uuid.UUID) (service.Note, error)
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
	mux.Handle("GET /notes/{id}", requireToken(h.Detail))
}

// List renders the note list view.
func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, noteview.List, noteview.ViewModel{
			ErrorBannerProps: component.ErrorBannerProps{
				Message: "Failed to load account, please try again.",
				Detail:  err.Error(),
			},
		})
		return
	}

	results, err := h.notes.List(tkn.ID())
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirect(w, r, "/login")
		return
	case err != nil:
		render(ctx, w, noteview.List, noteview.ViewModel{
			ErrorBannerProps: component.ErrorBannerProps{
				Message: "Failed to load notes, please try again.",
				Detail:  err.Error(),
			},
			DisplayName: account.DisplayName,
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

	render(ctx, w, noteview.List, noteview.ViewModel{
		DisplayName: account.DisplayName,
		Notes:       items,
	})
}

// Detail renders the note detail view for a single note record.
func (h *NoteHandler) Detail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	noteID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		render(ctx, w, noteview.Detail, noteview.DetailViewModel{
			ErrorBannerProps: component.ErrorBannerProps{
				Message: "Invalid note identifier.",
				Detail:  err.Error(),
			},
		})
		return
	}

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, noteview.Detail, noteview.DetailViewModel{
			ErrorBannerProps: component.ErrorBannerProps{
				Message: "Failed to load account, please try again.",
				Detail:  err.Error(),
			},
		})
		return
	}

	note, err := h.notes.Get(tkn.ID(), noteID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirect(w, r, "/login")
		return
	case errors.Is(err, service.ErrNoteNotFound):
		render(ctx, w, noteview.Detail, noteview.DetailViewModel{
			ErrorBannerProps: component.ErrorBannerProps{Message: "Note not found."},
			DisplayName:      account.DisplayName,
		})
		return
	case err != nil:
		render(ctx, w, noteview.Detail, noteview.DetailViewModel{
			ErrorBannerProps: component.ErrorBannerProps{
				Message: "Failed to load note, please try again.",
				Detail:  err.Error(),
			},
			DisplayName: account.DisplayName,
		})
		return
	}

	render(ctx, w, noteview.Detail, noteview.DetailViewModel{
		DisplayName: account.DisplayName,
		ID:          note.ID.String(),
		Name:        note.Name,
		Content:     note.Content,
	})
}
