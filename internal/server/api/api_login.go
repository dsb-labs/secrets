package api

import (
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"

	"github.com/davidsbond/passwords/internal/server/service"
	"github.com/davidsbond/passwords/internal/server/token"
)

type (
	// The LoginAPI exposes HTTP endpoints for managing individual user passwords.
	LoginAPI struct {
		logins LoginService
	}

	// The LoginService interface describes types that manage user passwords.
	LoginService interface {
		// Create should create a new password record for the given user id.
		Create(uuid.UUID, service.Login) error
		// List should return all passwords associated with the given user id.
		List(uuid.UUID) ([]service.Login, error)
	}

	// The Login type represents a single password.
	Login struct {
		// The unique identifier of the login.
		ID string `json:"id"`
		// The username.
		Username string `json:"username"`
		// The password.
		Password string `json:"password"`
		// The domains this password can be used.
		Domains []string `json:"domains"`
	}
)

// NewLoginAPI returns a new instance of the LoginAPI type that manages user passwords via the
// given LoginService implementation.
func NewLoginAPI(logins LoginService) *LoginAPI {
	return &LoginAPI{logins: logins}
}

// Register the HTTP endpoints onto the given http.ServeMux.
func (api *LoginAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/login", api.Create)
	mux.HandleFunc("GET /api/v1/login", api.List)
}

type (
	// The CreateLoginRequest type represents the request body given when calling LoginAPI.Create
	CreateLoginRequest struct {
		// The username.
		Username string `json:"username"`
		// The password.
		Password string `json:"password"`
		// The domains where this username/password combination can be used.
		Domains []string `json:"domains"`
	}

	// The CreateLoginResponse type represents the response body returned when calling LoginAPI.Create
	CreateLoginResponse struct{}
)

// Validate the request.
func (r CreateLoginRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Username, validation.Required),
		validation.Field(&r.Password, validation.Required),
	)
}

// Create handles an inbound HTTP request to store a new password record for a user. On success, it responds with
// an http.StatusCreated code and a JSON-encoded CreateLoginResponse.
func (api *LoginAPI) Create(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	if !tkn.Valid() {
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	request, err := decode[CreateLoginRequest](r.Body)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to decode request: %v", err)
		return
	}

	login := service.Login{
		Username: request.Username,
		Password: request.Password,
		Domains:  request.Domains,
	}

	err = api.logins.Create(tkn.ID(), login)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to create login: %v", err)
		return
	}

	write(w, http.StatusCreated, CreateLoginResponse{})
}

type (
	// The ListLoginsResponse type represents the response body returned when calling LoginAPI.List
	ListLoginsResponse struct {
		// The logins stored for the account.
		Logins []Login `json:"logins"`
	}
)

// List handles an inbound HTTP request to list all password records for a user. On success, it responds with
// an http.StatusOK code and a JSON-encoded ListLoginsResponse.
func (api *LoginAPI) List(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	if !tkn.Valid() {
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	results, err := api.logins.List(tkn.ID())
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to list logins: %v", err)
		return
	}

	logins := make([]Login, len(results))
	for i, result := range results {
		logins[i] = Login{
			ID:       result.ID.String(),
			Username: result.Username,
			Password: result.Password,
			Domains:  result.Domains,
		}
	}

	write(w, http.StatusOK, ListLoginsResponse{Logins: logins})
}
