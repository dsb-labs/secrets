package api

import (
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"

	"github.com/davidsbond/passwords/internal/server/service"
)

type (
	// The AccountAPI exposes HTTP endpoints for managing individual user accounts.
	AccountAPI struct {
		accounts AccountService
	}

	// The AccountService interface describes types that manage individual user accounts.
	AccountService interface {
		// Create should create the given account, returning the restore key to be used should the user enter a
		// disaster recovery scenario and need to manually decrypt their data. If an account with the given email
		// already exists, service.ErrAccountExists should be returned.
		Create(service.Account) ([]byte, error)
	}
)

// NewAccountAPI returns a new instance of the AccountAPI type that manages individual user accounts via the
// given AccountService implementation.
func NewAccountAPI(accounts AccountService) *AccountAPI {
	return &AccountAPI{
		accounts: accounts,
	}
}

// Register the HTTP endpoints onto the given http.ServeMux.
func (api *AccountAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/account", api.Create)
}

type (
	// The CreateAccountRequest type represents the request body given when calling AccountAPI.Create.
	CreateAccountRequest struct {
		// The user's email address.
		Email string `json:"email"`
		// The user's password.
		Password string `json:"password"`
	}

	// The CreateAccountResponse type represents the response body returned when calling AccountAPI.Create.
	CreateAccountResponse struct {
		// The key to use if manual data decryption is required.
		RestoreKey []byte `json:"restoreKey"`
	}
)

// Validate the request.
func (r CreateAccountRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email, validation.Required, is.Email),
		validation.Field(&r.Password, validation.Required),
	)
}

// Create handles an inbound HTTP request to create a new account. On success, it responds with an http.StatusCreated
// code and a JSON-encoded CreateAccountResponse.
func (api *AccountAPI) Create(w http.ResponseWriter, r *http.Request) {
	request, err := decode[CreateAccountRequest](r.Body)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to decode request: %v", err)
		return
	}

	account := service.Account{
		Email:    request.Email,
		Password: request.Password,
	}

	restoreKey, err := api.accounts.Create(account)
	switch {
	case errors.Is(err, service.ErrAccountExists):
		writeErrorf(w, http.StatusConflict, "account %q already exists", account.Email)
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to create account %q: %v", account.Email, err)
		return
	}

	write(w, http.StatusCreated, CreateAccountResponse{
		RestoreKey: restoreKey,
	})
}
