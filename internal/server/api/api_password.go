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
	// The PasswordAPI exposes HTTP endpoints for managing individual user passwords.
	PasswordAPI struct {
		passwords PasswordService
	}

	// The PasswordService interface describes types that manage user passwords.
	PasswordService interface {
		// Create should create a new password record.
		Create(service.Password) error
		// List should return all passwords associated with the given user id.
		List(uuid.UUID) ([]service.Password, error)
	}

	// The Password type represents a single password.
	Password struct {
		// The username.
		Username string `json:"username"`
		// The password.
		Password string `json:"password"`
		// The domains this password can be used.
		Domains []string `json:"domains"`
	}
)

func NewPasswordAPI(passwords PasswordService) *PasswordAPI {
	return &PasswordAPI{passwords: passwords}
}

func (api *PasswordAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/password", api.Create)
	mux.HandleFunc("GET /api/v1/password", api.List)
}

type (
	CreatePasswordRequest struct {
		Username string   `json:"username"`
		Password string   `json:"password"`
		Domains  []string `json:"domains"`
	}

	CreatePasswordResponse struct{}
)

func (r CreatePasswordRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Username, validation.Required),
		validation.Field(&r.Password, validation.Required),
	)
}

func (api *PasswordAPI) Create(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	if !tkn.Valid() {
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	request, err := decode[CreatePasswordRequest](r.Body)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to decode request: %v", err)
		return
	}

	password := service.Password{
		UserID:   tkn.ID(),
		Username: request.Username,
		Password: request.Password,
		Domains:  request.Domains,
	}

	err = api.passwords.Create(password)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to create password: %v", err)
		return
	}

	write(w, http.StatusCreated, CreatePasswordResponse{})
}

type (
	ListPasswordsResponse struct {
		Passwords []Password `json:"passwords"`
	}
)

func (api *PasswordAPI) List(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	if !tkn.Valid() {
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	results, err := api.passwords.List(tkn.ID())
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to list passwords: %v", err)
		return
	}

	passwords := make([]Password, len(results))
	for i, result := range results {
		passwords[i] = Password{
			Username: result.Username,
			Password: result.Password,
			Domains:  result.Domains,
		}
	}

	write(w, http.StatusOK, ListPasswordsResponse{Passwords: passwords})
}
