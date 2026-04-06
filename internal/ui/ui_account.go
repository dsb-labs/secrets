package ui

import (
	"encoding/base64"
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"

	"github.com/dsb-labs/secrets/internal/server/service"
	"github.com/dsb-labs/secrets/internal/server/token"
	accountview "github.com/dsb-labs/secrets/internal/ui/view/account"
	"github.com/dsb-labs/secrets/internal/ui/view/auth"
	statusview "github.com/dsb-labs/secrets/internal/ui/view/status"
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
	mux.Handle("GET /account", requireToken(h.Detail))
	mux.Handle("GET /account/password", requireToken(h.ChangePassword))
	mux.Handle("POST /account/password", requireToken(h.ChangePasswordCallback))
	mux.Handle("GET /account/delete", requireToken(h.Delete))
	mux.Handle("POST /account/delete", requireToken(h.DeleteCallback))
	mux.HandleFunc("GET /logout", h.Logout)
}

// Detail renders the account detail view.
func (h *AccountHandler) Detail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, accountview.Detail, accountview.ViewModel{
		DisplayName: account.DisplayName,
		Email:       account.Email,
	})
}

// ChangePassword renders the change password form.
func (h *AccountHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, accountview.ChangePassword, accountview.ChangePasswordViewModel{
		DisplayName: account.DisplayName,
	})
}

// The ChangePasswordForm type represents the form values submitted when calling AccountHandler.ChangePasswordCallback.
type ChangePasswordForm struct {
	// The user's current password.
	OldPassword string `form:"oldPassword"`
	// The user's desired new password.
	NewPassword string `form:"newPassword"`
	// The user's desired new password, repeated for confirmation.
	ConfirmPassword string `form:"confirmPassword"`
}

// Validate the form.
func (f ChangePasswordForm) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.OldPassword, validation.Required),
		validation.Field(&f.NewPassword, validation.Required),
		validation.Field(&f.ConfirmPassword, validation.Required),
	)
}

// ChangePasswordCallback handles the change password form submission. On success it renders the change password
// success view, presenting the user with their new restore key.
func (h *AccountHandler) ChangePasswordCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	form, err := decode[ChangePasswordForm](r)

	model := accountview.ChangePasswordViewModel{
		DisplayName: account.DisplayName,
	}

	var ve validation.Errors
	switch {
	case errors.As(err, &ve):
		model.Validation.Errors = validationErrors(ve)
		render(ctx, w, http.StatusUnprocessableEntity, accountview.ChangePassword, model)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	if form.NewPassword != form.ConfirmPassword {
		model.Error.Message = "Passwords do not match"
		render(ctx, w, http.StatusUnprocessableEntity, accountview.ChangePassword, model)
		return
	}

	restoreKey, err := h.accounts.ChangePassword(tkn.ID(), form.OldPassword, form.NewPassword)
	switch {
	case errors.Is(err, service.ErrInvalidPassword):
		model.Error.Message = "Current password is incorrect"
		render(ctx, w, http.StatusUnprocessableEntity, accountview.ChangePassword, model)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "secrets",
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	render(ctx, w, http.StatusOK, accountview.ChangePasswordSuccess, accountview.ChangePasswordSuccessViewModel{
		DisplayName: account.DisplayName,
		RestoreKey:  base64.StdEncoding.EncodeToString(restoreKey),
	})
}

// Delete renders the delete account confirmation form.
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, accountview.Delete, accountview.DeleteViewModel{
		DisplayName: account.DisplayName,
	})
}

// The DeleteAccountForm type represents the form values submitted when calling AccountHandler.DeleteCallback.
type DeleteAccountForm struct {
	// The user's current password, required to confirm account deletion.
	Password string `form:"password"`
}

// Validate the form.
func (f DeleteAccountForm) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Password, validation.Required),
	)
}

// DeleteCallback handles an account deletion request. On success, it clears the session cookie and redirects to the
// login page.
func (h *AccountHandler) DeleteCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	form, err := decode[DeleteAccountForm](r)

	model := accountview.DeleteViewModel{
		DisplayName: account.DisplayName,
	}

	var ve validation.Errors
	switch {
	case errors.As(err, &ve):
		model.Validation.Errors = validationErrors(ve)
		render(ctx, w, http.StatusUnprocessableEntity, accountview.Delete, model)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	err = h.accounts.VerifyPassword(tkn.ID(), form.Password)
	switch {
	case errors.Is(err, service.ErrInvalidPassword):
		model.Error.Message = "Password is incorrect"
		render(ctx, w, http.StatusUnprocessableEntity, accountview.Delete, model)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	if err = h.accounts.Delete(tkn.ID()); err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "secrets",
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	redirect(w, r, "/login")
}

// Logout clears the session cookie and redirects to the login page.
func (h *AccountHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "secrets",
		Value:    "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	redirect(w, r, "/login")
}

// CreateAccount renders the account creation view.
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	render(r.Context(), w, http.StatusOK, auth.Register, auth.RegisterViewModel{})
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
		model.Validation.Errors = validationErrors(ve)
		render(ctx, w, http.StatusUnprocessableEntity, auth.Register, model)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	if form.Password != form.ConfirmPassword {
		model.Error.Message = "Passwords do not match"
		render(ctx, w, http.StatusUnprocessableEntity, auth.Register, model)
		return
	}

	restoreKey, err := h.accounts.Create(service.Account{
		DisplayName: form.DisplayName,
		Email:       form.Email,
		Password:    form.Password,
	})
	switch {
	case errors.Is(err, service.ErrAccountExists):
		model.Error.Message = "An account with that email address already exists"
		render(ctx, w, http.StatusConflict, auth.Register, model)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, auth.RegisterSuccess, auth.RegisterSuccessViewModel{
		DisplayName: form.DisplayName,
		RestoreKey:  base64.StdEncoding.EncodeToString(restoreKey),
	})
}
