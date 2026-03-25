package ui

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/ui/view"
)

type (
	// The AccountHandler type is responsible for serving web interface pages regarding account management.
	AccountHandler struct {
		accounts AccountService
	}

	// The AccountService interface describes types that manage user accounts.
	AccountService interface {
		// Create should create the given account, returning the restore key to be presented to the user. If an account
		// with the given email already exists, service.ErrAccountExists should be returned.
		Create(account service.Account) ([]byte, error)
	}
)

// NewAccountHandler returns a new instance of the AccountHandler type that will serve account management UIs using
// the provided AccountService implementation.
func NewAccountHandler(accounts AccountService) *AccountHandler {
	return &AccountHandler{accounts: accounts}
}

// Register HTTP endpoints onto the provided http.ServeMux.
func (h *AccountHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /register", h.CreateAccount)
	mux.HandleFunc("POST /register", h.CreateAccountCallback)
}

// CreateAccount renders the account creation view.
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	render(r.Context(), w, view.Register, view.RegisterViewModel{})
}

// CreateAccountCallback handles an account creation attempt, re-rendering the registration view on error. On success,
// it renders the registration success view, presenting the user with their restore key.
func (h *AccountHandler) CreateAccountCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	displayName := r.FormValue("displayName")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")

	model := view.RegisterViewModel{
		DisplayName: displayName,
		Email:       email,
	}

	if password != confirmPassword {
		model.Error = "Passwords do not match"
		render(ctx, w, view.Register, model)
		return
	}

	restoreKey, err := h.accounts.Create(service.Account{
		DisplayName: displayName,
		Email:       email,
		Password:    password,
	})
	switch {
	case errors.Is(err, service.ErrAccountExists):
		model.Error = "An account with that email address already exists"
		render(ctx, w, view.Register, model)
		return
	case err != nil:
		model.Error = err.Error()
		render(ctx, w, view.Register, model)
		return
	}

	render(ctx, w, view.RegisterSuccess, view.RegisterSuccessViewModel{
		DisplayName: displayName,
		RestoreKey:  base64.StdEncoding.EncodeToString(restoreKey),
	})
}
