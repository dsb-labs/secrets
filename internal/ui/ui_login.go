package ui

import (
	"errors"
	"net/http"

	"github.com/davidsbond/x/filter"
	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
	"github.com/davidsbond/keeper/internal/ui/component"
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
		// Get should return the login record associated with the given user and login identifiers.
		Get(accountID uuid.UUID, loginID uuid.UUID) (service.Login, error)
		// Delete should remove the login record associated with the given user and login identifiers.
		Delete(accountID uuid.UUID, loginID uuid.UUID) error
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
	mux.Handle("GET /logins/{id}", requireToken(h.Detail))
	mux.Handle("POST /logins/{id}/delete", requireToken(h.Delete))
}

// List renders the login list view.
func (h *LoginHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, loginview.List, loginview.ViewModel{
			ErrorBannerProps: component.ErrorBannerProps{
				Message: "Failed to load account, please try again.",
				Detail:  err.Error(),
			},
		})
		return
	}

	results, err := h.logins.List(tkn.ID())
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case err != nil:
		render(ctx, w, loginview.List, loginview.ViewModel{
			ErrorBannerProps: component.ErrorBannerProps{
				Message: "Failed to load logins, please try again.",
				Detail:  err.Error(),
			},
			DisplayName: account.DisplayName,
		})
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

// Detail renders the login detail view for a single login record.
func (h *LoginHandler) Detail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	loginID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		render(ctx, w, loginview.Detail, loginview.DetailViewModel{
			ErrorBannerProps: component.ErrorBannerProps{
				Message: "Invalid login identifier.",
				Detail:  err.Error(),
			},
		})
		return
	}

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, loginview.Detail, loginview.DetailViewModel{
			ErrorBannerProps: component.ErrorBannerProps{
				Message: "Failed to load account, please try again.",
				Detail:  err.Error(),
			},
		})
		return
	}

	login, err := h.logins.Get(tkn.ID(), loginID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case errors.Is(err, service.ErrLoginNotFound):
		render(ctx, w, loginview.Detail, loginview.DetailViewModel{
			ErrorBannerProps: component.ErrorBannerProps{Message: "Login not found."},
			DisplayName:      account.DisplayName,
		})
		return
	case err != nil:
		render(ctx, w, loginview.Detail, loginview.DetailViewModel{
			ErrorBannerProps: component.ErrorBannerProps{
				Message: "Failed to load login, please try again.",
				Detail:  err.Error(),
			},
			DisplayName: account.DisplayName,
		})
		return
	}

	render(ctx, w, loginview.Detail, loginview.DetailViewModel{
		DisplayName: account.DisplayName,
		ID:          login.ID.String(),
		Username:    login.Username,
		Password:    login.Password,
		Domains:     login.Domains,
	})
}

// Delete handles a login deletion request, redirecting to the login list on success.
func (h *LoginHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	loginID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		render(ctx, w, loginview.Detail, loginview.DetailViewModel{
			ErrorBannerProps: component.ErrorBannerProps{
				Message: "Invalid login identifier.",
				Detail:  err.Error(),
			},
		})
		return
	}

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, loginview.Detail, loginview.DetailViewModel{
			ErrorBannerProps: component.ErrorBannerProps{
				Message: "Failed to load account, please try again.",
				Detail:  err.Error(),
			},
		})
		return
	}

	err = h.logins.Delete(tkn.ID(), loginID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case errors.Is(err, service.ErrLoginNotFound):
		render(ctx, w, loginview.Detail, loginview.DetailViewModel{
			ErrorBannerProps: component.ErrorBannerProps{Message: "Login not found."},
			DisplayName:      account.DisplayName,
		})
		return
	case err != nil:
		render(ctx, w, loginview.Detail, loginview.DetailViewModel{
			ErrorBannerProps: component.ErrorBannerProps{
				Message: "Failed to delete login, please try again.",
				Detail:  err.Error(),
			},
			DisplayName: account.DisplayName,
		})
		return
	}

	redirect(w, r, "/logins")
}
