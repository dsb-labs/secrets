package ui

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/ui/view/auth"
)

// The AccountHandler type is responsible for serving web interface pages regarding account management.
type AccountHandler struct {
	accounts AccountService
}

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
	render(r.Context(), w, auth.Register, auth.RegisterViewModel{})
}

// CreateAccountCallback handles an account creation attempt, re-rendering the registration view on error. On success,
// it renders the registration success view, presenting the user with their restore key.
func (h *AccountHandler) CreateAccountCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	displayName := r.FormValue("displayName")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")

	model := auth.RegisterViewModel{
		DisplayName: displayName,
		Email:       email,
	}

	if password != confirmPassword {
		model.Error = "Passwords do not match"
		render(ctx, w, auth.Register, model)
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
		render(ctx, w, auth.Register, model)
		return
	case err != nil:
		model.Error = "An unexpected error occurred, please try again."
		model.ErrorDetail = err.Error()
		render(ctx, w, auth.Register, model)
		return
	}

	render(ctx, w, auth.RegisterSuccess, auth.RegisterSuccessViewModel{
		DisplayName: displayName,
		RestoreKey:  base64.StdEncoding.EncodeToString(restoreKey),
	})
}
