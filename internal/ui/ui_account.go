package ui

import (
	"encoding/base64"
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"

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

// The CreateAccountForm type represents the form values submitted when calling AccountHandler.CreateAccountCallback.
type CreateAccountForm struct {
	// The user's display name.
	DisplayName string `form:"displayName"`
	// The user's email address.
	Email string `form:"email"`
	// The user's password.
	Password string `form:"password"`
	// The user's password, repeated for confirmation.
	ConfirmPassword string `form:"confirmPassword"`
}

// Validate the form.
func (f CreateAccountForm) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.DisplayName, validation.Required),
		validation.Field(&f.Email, validation.Required, is.Email),
		validation.Field(&f.Password, validation.Required),
		validation.Field(&f.ConfirmPassword, validation.Required),
	)
}

// CreateAccountCallback handles an account creation attempt, re-rendering the registration view on error. On success,
// it renders the registration success view, presenting the user with their restore key.
func (h *AccountHandler) CreateAccountCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	form, err := decode[CreateAccountForm](r)

	model := auth.RegisterViewModel{
		DisplayName: form.DisplayName,
		Email:       form.Email,
	}

	var ve validation.Errors
	switch {
	case errors.As(err, &ve):
		model.Errors = validationErrors(ve)
		render(ctx, w, auth.Register, model)
		return
	case err != nil:
		model.Message = "An unexpected error occurred, please try again."
		model.Detail = err.Error()
		render(ctx, w, auth.Register, model)
		return
	}

	if form.Password != form.ConfirmPassword {
		model.Message = "Passwords do not match"
		render(ctx, w, auth.Register, model)
		return
	}

	restoreKey, err := h.accounts.Create(service.Account{
		DisplayName: form.DisplayName,
		Email:       form.Email,
		Password:    form.Password,
	})
	switch {
	case errors.Is(err, service.ErrAccountExists):
		model.Message = "An account with that email address already exists"
		render(ctx, w, auth.Register, model)
		return
	case err != nil:
		model.Message = "An unexpected error occurred, please try again."
		model.Detail = err.Error()
		render(ctx, w, auth.Register, model)
		return
	}

	render(ctx, w, auth.RegisterSuccess, auth.RegisterSuccessViewModel{
		DisplayName: form.DisplayName,
		RestoreKey:  base64.StdEncoding.EncodeToString(restoreKey),
	})
}
