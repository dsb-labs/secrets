package ui

import (
	"errors"
	"net/http"

	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
	"github.com/davidsbond/keeper/internal/server/ui/view"
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
	render(r.Context(), w, "Login", view.Login, view.LoginViewModel{})
}

// LoginCallback handles a login attempt, rerendering the login view on error. On success, it redirects to the dashboard
// page.
func (h *AuthHandler) LoginCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email := r.FormValue("email")
	password := r.FormValue("password")

	model := view.LoginViewModel{
		Email:    email,
		Password: password,
	}

	tkn, err := h.auth.Login(email, password)
	switch {
	case errors.Is(err, service.ErrAccountNotFound):
		model.Error = "Account not found"
		render(ctx, w, "Login", view.Login, model)
		return
	case errors.Is(err, service.ErrInvalidPassword):
		model.Error = "Invalid password"
		render(ctx, w, "Login", view.Login, model)
		return
	case err != nil:
		model.Error = err.Error()
		render(ctx, w, "Login", view.Login, model)
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

	redirect(w, r, "/dashboard")
}
