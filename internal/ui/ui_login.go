package ui

import (
	"errors"
	"net/http"

	"github.com/davidsbond/x/filter"
	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
	loginview "github.com/davidsbond/keeper/internal/ui/view/login"
)

type (
	// The LoginHandler type is responsible for serving web interface pages regarding user logins.
	LoginHandler struct {
		accounts AccountService
		logins   LoginService
	}

	// The LoginService interface describes types that manage user login records.
	LoginService interface {
		// List should return all logins associated with the given user id.
		List(accountID uuid.UUID, filters ...filter.Filter[service.Login]) ([]service.Login, error)
	}
)

// NewLoginHandler returns a new instance of the LoginHandler type that will serve login management UIs using
// the provided service implementations.
func NewLoginHandler(accounts AccountService, logins LoginService) *LoginHandler {
	return &LoginHandler{accounts: accounts, logins: logins}
}

// Register HTTP endpoints onto the provided http.ServeMux.
func (h *LoginHandler) Register(mux *http.ServeMux) {
	mux.Handle("GET /logins", requireToken(h.List))
}

// List renders the login list view.
func (h *LoginHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		http.Error(w, "failed to load account", http.StatusInternalServerError)
		return
	}

	results, err := h.logins.List(tkn.ID())
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirect(w, r, "/login")
		return
	case err != nil:
		http.Error(w, "failed to load logins", http.StatusInternalServerError)
		return
	}

	items := make([]loginview.Item, len(results))
	for i, l := range results {
		items[i] = loginview.Item{
			ID:       l.ID.String(),
			Username: l.Username,
			Domains:  l.Domains,
		}
	}

	render(ctx, w, loginview.List, loginview.ViewModel{
		DisplayName: account.DisplayName,
		Logins:      items,
	})
}
