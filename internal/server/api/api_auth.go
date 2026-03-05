package api

import (
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
)

type (
	// The AuthAPI exposes HTTP endpoints for authenticating individual user accounts.
	AuthAPI struct {
		auth AuthService
	}

	// The AuthService interface describes types that manage authentication of users.
	AuthService interface {
		// Login should return a token.Token if the provided email and password combination is correct. This Token should
		// be given to the user for subsequent API calls.
		Login(string, string) (token.Token, error)
		// Logout should close any account databases associated with the given UUID.
		Logout(uuid.UUID) error
	}
)

// NewAuthAPI returns a new instance of the AuthAPI type that manages authenticating user accounts via the
// given AuthService implementation.
func NewAuthAPI(auth AuthService) *AuthAPI {
	return &AuthAPI{auth: auth}
}

// Register the HTTP endpoints onto the given http.ServeMux.
func (api *AuthAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth", api.Login)
	mux.Handle("DELETE /api/v1/auth", requireToken(api.Logout))
}

type (
	// The LoginRequest type represents the request body given when calling AuthAPI.Login
	LoginRequest struct {
		// The user's email address.
		Email string `json:"email"`
		// The user's password.
		Password string `json:"password"`
	}

	// The LoginResponse type represents the response body returned when calling AuthAPI.Login
	LoginResponse struct {
		// The user's authentication token.
		Token string `json:"token"`
	}
)

// Validate the request.
func (r LoginRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email, validation.Required, is.Email),
		validation.Field(&r.Password, validation.Required),
	)
}

// Login handles an inbound HTTP request to generate a new authentication token for a user. On success, it responds with
// an http.StatusOK code and a JSON-encoded LoginResponse.
func (api *AuthAPI) Login(w http.ResponseWriter, r *http.Request) {
	request, err := decode[LoginRequest](r.Body)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to decode request: %v", err)
		return
	}

	tkn, err := api.auth.Login(request.Email, request.Password)
	switch {
	case errors.Is(err, service.ErrAccountNotFound):
		writeErrorf(w, http.StatusNotFound, "account %q does not exist", request.Email)
		return
	case errors.Is(err, service.ErrInvalidPassword):
		writeErrorf(w, http.StatusBadRequest, "invalid password for account %q", request.Email)
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to login user %q: %v", request.Email, err)
		return
	}

	write(w, http.StatusOK, LoginResponse{
		Token: tkn.String(),
	})
}

type (
	// The LogoutResponse type represents the response body returned when calling AuthAPI.Logout
	LogoutResponse struct{}
)

// Logout handles an inbound HTTP request to lock the caller's individual account database. On success, it responds with
// an http.StatusOK code and a JSON-encoded LogoutResponse.
func (api *AuthAPI) Logout(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	if err := api.auth.Logout(tkn.ID()); err != nil {
		writeErrorf(w, http.StatusInternalServerError, "failed to logout: %v", err)
		return
	}

	write(w, http.StatusOK, LogoutResponse{})
}
