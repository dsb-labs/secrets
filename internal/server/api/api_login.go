package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/davidsbond/x/convert"
	"github.com/davidsbond/x/filter"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
)

type (
	// The LoginAPI exposes HTTP endpoints for managing individual user logins.
	LoginAPI struct {
		logins LoginService
	}

	// The LoginService interface describes types that manage user logins.
	LoginService interface {
		// Create should create a new login record for the given user id.
		Create(accountID uuid.UUID, login service.Login) error
		// List should return all logins associated with the given user id.
		List(accountID uuid.UUID, filters ...filter.Filter[service.Login]) ([]service.Login, error)
		// Delete should remove the login record associated with the given user and login id. Returning
		// service.ErrLoginNotFound if it does not exist.
		Delete(accountID uuid.UUID, loginID uuid.UUID) error
		// Get should return the login record associated with the given user and login id. Returning
		// service.ErrLoginNotFound if it does not exist.
		Get(accountID uuid.UUID, loginID uuid.UUID) (service.Login, error)
	}

	// The Login type represents a single username/password combination.
	Login struct {
		// The unique identifier of the login.
		ID string `json:"id"`
		// The username.
		Username string `json:"username"`
		// The password.
		Password string `json:"password"`
		// The domains this password can be used.
		Domains []string `json:"domains"`
		// When the login was created.
		CreatedAt time.Time `json:"createdAt"`
	}
)

// NewLoginAPI returns a new instance of the LoginAPI type that manages user logins via the
// given LoginService implementation.
func NewLoginAPI(logins LoginService) *LoginAPI {
	return &LoginAPI{logins: logins}
}

// Register the HTTP endpoints onto the given http.ServeMux.
func (api *LoginAPI) Register(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/login", requireToken(api.Create))
	mux.Handle("GET /api/v1/login", requireToken(api.List))
	mux.Handle("GET /api/v1/login/{id}", requireToken(api.Get))
	mux.Handle("DELETE /api/v1/login/{id}", requireToken(api.Delete))
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
	CreateLoginResponse struct {
		ID string `json:"id"`
	}
)

// Validate the request.
func (r CreateLoginRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Username, validation.Required),
		validation.Field(&r.Password, validation.Required),
	)
}

// Create handles an inbound HTTP request to store a new login record for a user. On success, it responds with
// an http.StatusCreated code and a JSON-encoded CreateLoginResponse.
func (api *LoginAPI) Create(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	request, err := decode[CreateLoginRequest](r.Body)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to decode request: %v", err)
		return
	}

	login := service.Login{
		ID:        uuid.New(),
		Username:  request.Username,
		Password:  request.Password,
		Domains:   request.Domains,
		CreatedAt: time.Now(),
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

	write(w, http.StatusCreated, CreateLoginResponse{
		ID: login.ID.String(),
	})
}

type (
	// The ListLoginsResponse type represents the response body returned when calling LoginAPI.List
	ListLoginsResponse struct {
		// The logins stored for the account.
		Logins []Login `json:"logins"`
	}
)

// List handles an inbound HTTP request to list all login records for a user. On success, it responds with
// an http.StatusOK code and a JSON-encoded ListLoginsResponse.
func (api *LoginAPI) List(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	filters := make([]filter.Filter[service.Login], 0)
	if domain := r.URL.Query().Get("domain"); domain != "" {
		filters = append(filters, service.LoginsByDomain(domain))
	}

	results, err := api.logins.List(tkn.ID(), filters...)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to list logins: %v", err)
		return
	}

	write(w, http.StatusOK, ListLoginsResponse{
		Logins: convert.Slice(results, func(in service.Login) Login {
			return Login{
				ID:        in.ID.String(),
				Username:  in.Username,
				Password:  in.Password,
				Domains:   in.Domains,
				CreatedAt: in.CreatedAt,
			}
		}),
	})
}

type (
	// The DeleteLoginResponse type represents the response body returned when calling LoginAPI.Delete
	DeleteLoginResponse struct{}
)

// Delete handles an inbound HTTP request to delete a login record for a user. On success, it responds with
// an http.StatusOK code and a JSON-encoded DeleteLoginResponse.
func (api *LoginAPI) Delete(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	loginID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to parse login id: %v", err)
		return
	}

	err = api.logins.Delete(tkn.ID(), loginID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case errors.Is(err, service.ErrLoginNotFound):
		writeErrorf(w, http.StatusNotFound, "login %q does not exist", loginID)
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to delete login: %v", err)
		return
	}

	write(w, http.StatusOK, DeleteLoginResponse{})
}

type (
	// The GetLoginResponse type represents the response body returned when calling LoginAPI.Get
	GetLoginResponse struct {
		// The requested login details.
		Login Login `json:"login"`
	}
)

// Get handles an inbound HTTP request to query a login record for a user. On success, it responds with
// an http.StatusOK code and a JSON-encoded GetLoginResponse.
func (api *LoginAPI) Get(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	loginID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to parse login id: %v", err)
		return
	}

	result, err := api.logins.Get(tkn.ID(), loginID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case errors.Is(err, service.ErrLoginNotFound):
		writeErrorf(w, http.StatusNotFound, "login %q does not exist", loginID)
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to get login: %v", err)
		return
	}

	write(w, http.StatusOK, GetLoginResponse{
		Login: Login{
			ID:        result.ID.String(),
			Username:  result.Username,
			Password:  result.Password,
			Domains:   result.Domains,
			CreatedAt: result.CreatedAt,
		},
	})
}
