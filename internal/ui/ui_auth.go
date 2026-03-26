package ui

import (
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"

	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
	"github.com/davidsbond/keeper/internal/ui/view/auth"
)

type (
	// The AuthHandler type is responsible for serving web interface pages regarding authentication.
	AuthHandler struct {
		auth AuthService
	}

	// The AuthService interface describes types that handler user authentication.
	AuthService interface {
		// Login should return a token.Token if the provided email and password combination is correct. This Token should
		// be given to the user for subsequent calls.
		Login(string, string) (token.Token, error)
	}
)

// NewAuthHandler returns a new instance of the AuthHandler type that will serve authentication UIs using the provided
// AuthService implementation.
func NewAuthHandler(auth AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register HTTP endpoints onto the provided http.ServeMux.
func (h *AuthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /login", h.Login)
	mux.HandleFunc("POST /login", h.LoginCallback)
}

// Login renders the login view.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	render(r.Context(), w, auth.Login, auth.LoginViewModel{
		Redirect: r.URL.Query().Get("redirect"),
	})
}

// The LoginForm type represents the form values submitted when calling AuthHandler.LoginCallback.
type LoginForm struct {
	// The user's email address.
	Email string `form:"email"`
	// The user's password.
	Password string `form:"password"`
}

// Validate the form.
func (f LoginForm) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Email, validation.Required, is.Email),
		validation.Field(&f.Password, validation.Required),
	)
}

// LoginCallback handles a login attempt, rerendering the login view on error. On success, it redirects to the
// originally requested page, or the dashboard if none was captured.
func (h *AuthHandler) LoginCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	redirectTo := r.FormValue("redirect")

	form, err := decode[LoginForm](r)
	model := auth.LoginViewModel{
		Email:    form.Email,
		Password: form.Password,
		Redirect: redirectTo,
	}

	var ve validation.Errors
	switch {
	case errors.As(err, &ve):
		model.Errors = validationErrors(ve)
		render(ctx, w, auth.Login, model)
		return
	case err != nil:
		model.Message = "An unexpected error occurred, please try again."
		model.Detail = err.Error()
		render(ctx, w, auth.Login, model)
		return
	}

	tkn, err := h.auth.Login(form.Email, form.Password)
	switch {
	case errors.Is(err, service.ErrAccountNotFound):
		model.Message = "Account not found"
		render(ctx, w, auth.Login, model)
		return
	case errors.Is(err, service.ErrInvalidPassword):
		model.Message = "Invalid password"
		render(ctx, w, auth.Login, model)
		return
	case err != nil:
		model.Message = "An unexpected error occurred, please try again."
		model.Detail = err.Error()
		render(ctx, w, auth.Login, model)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "keeper",
		Value:    tkn.String(),
		Expires:  tkn.ExpiresAt(),
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	destination := "/dashboard"
	if redirectTo != "" {
		destination = redirectTo
	}

	redirect(w, r, destination)
}
