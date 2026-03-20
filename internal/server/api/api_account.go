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
	// The AccountAPI exposes HTTP endpoints for managing individual user accounts.
	AccountAPI struct {
		accounts AccountService
	}

	// The AccountService interface describes types that manage individual user accounts.
	AccountService interface {
		// Create should create the given account, returning the restore key to be used should the user enter a
		// disaster recovery scenario and need to manually decrypt their data. If an account with the given email
		// already exists, service.ErrAccountExists should be returned.
		Create(account service.Account) ([]byte, error)
		// Get should return the account associated with the given identifier. Returning service.ErrAccountNotFound
		// if the account does not exist.
		Get(id uuid.UUID) (service.Account, error)
		// Delete should delete the account associated with the given identifier. Returning service.ErrAccountNotFound
		// if the account does not exist.
		Delete(id uuid.UUID) error
		// ChangePassword should update the account's password to the new value if the old password provided is
		// correct, returning the restore key to be used should the user enter a disaster recovery scenario and need to
		// manually decrypt their data.
		ChangePassword(id uuid.UUID, oldPassword string, newPassword string) ([]byte, error)
		// Restore should update the account associated with the given email address' password after verifying the
		// provided restore key is valid. Returning service.ErrAccountNotFound if the account does not exist or
		// service.ErrInvalidRestoreKey if the given restore key is invalid. On success, it should return the new
		// restore key to be used should the user enter a disaster recovery scenario and need to manually decrypt their
		// data.
		Restore(email string, restoreKey []byte, newPassword string) ([]byte, error)
	}

	// The Account type represents an individual user account as returned by the API.
	Account struct {
		// The user's display name.
		DisplayName string `json:"displayName"`
		// The user's email address.
		Email string `json:"email"`
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
	mux.Handle("GET /api/v1/account", requireToken(api.Get))
	mux.Handle("DELETE /api/v1/account", requireToken(api.Delete))
	mux.Handle("PUT /api/v1/account/password", requireToken(api.UpdatePassword))
	mux.HandleFunc("POST /api/v1/account/restore", api.Restore)
}

type (
	// The CreateAccountRequest type represents the request body given when calling AccountAPI.Create.
	CreateAccountRequest struct {
		// The user's email address.
		Email string `json:"email"`
		// The user's password.
		Password string `json:"password"`
		// The user's display name.
		DisplayName string `json:"displayName"`
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
		validation.Field(&r.DisplayName, validation.Required),
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
		Email:       request.Email,
		Password:    request.Password,
		DisplayName: request.DisplayName,
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

type (
	// The GetAccountResponse type represents the response body returned when calling AccountAPI.Get
	GetAccountResponse struct {
		// The user's account details.
		Account Account `json:"account"`
	}
)

// Get handles an inbound HTTP request to query the caller's account details. On success, it responds with
// an http.StatusOK code and a JSON-encoded GetAccountResponse.
func (api *AccountAPI) Get(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	account, err := api.accounts.Get(tkn.ID())
	switch {
	case errors.Is(err, service.ErrAccountNotFound):
		writeError(w, http.StatusNotFound, "account not found")
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to get account %v", err)
		return
	}

	write(w, http.StatusOK, GetAccountResponse{
		Account: Account{
			DisplayName: account.DisplayName,
			Email:       account.Email,
		},
	})
}

type (
	// The DeleteAccountResponse type represents the response body returned when calling AccountAPI.Delete
	DeleteAccountResponse struct{}
)

// Delete handles an inbound HTTP request to delete the caller's account. On success, it responds with
// an http.StatusOK code and a JSON-encoded DeleteAccountResponse.
func (api *AccountAPI) Delete(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	err := api.accounts.Delete(tkn.ID())
	switch {
	case errors.Is(err, service.ErrAccountNotFound):
		writeError(w, http.StatusNotFound, "account not found")
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to delete account %v", err)
		return
	}

	write(w, http.StatusOK, DeleteAccountResponse{})
}

type (
	// The UpdatePasswordRequest type represents the request body given when calling AccountAPI.UpdatePassword.
	UpdatePasswordRequest struct {
		// The user's current password.
		OldPassword string `json:"oldPassword"`
		// The user's new password.
		NewPassword string `json:"newPassword"`
	}

	// The UpdatePasswordResponse type represents the response body returned when calling AccountAPI.UpdatePassword
	UpdatePasswordResponse struct {
		// The key to use if manual data decryption is required.
		RestoreKey []byte `json:"restoreKey"`
	}
)

// Validate the request.
func (r UpdatePasswordRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.OldPassword, validation.Required),
		validation.Field(&r.NewPassword, validation.Required, validation.NotIn(r.OldPassword)),
	)
}

// UpdatePassword handles an inbound HTTP request to change the caller's password. On success, it responds with
// an http.StatusOK code and a JSON-encoded UpdatePasswordResponse.
func (api *AccountAPI) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	request, err := decode[UpdatePasswordRequest](r.Body)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to decode request: %v", err)
		return
	}

	restoreKey, err := api.accounts.ChangePassword(tkn.ID(), request.OldPassword, request.NewPassword)
	switch {
	case errors.Is(err, service.ErrAccountNotFound):
		writeError(w, http.StatusNotFound, "account not found")
		return
	case errors.Is(err, service.ErrInvalidPassword):
		writeError(w, http.StatusBadRequest, "invalid password for account")
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to update password: %v", err)
		return
	}

	write(w, http.StatusOK, UpdatePasswordResponse{
		RestoreKey: restoreKey,
	})
}

type (
	// The RestoreAccountRequest type represents the request body given when calling AccountAPI.Restore.
	RestoreAccountRequest struct {
		// The account's email address.
		Email string `json:"email"`
		// The account's restore key.
		RestoreKey []byte `json:"restoreKey"`
		// The user's new password.
		NewPassword string `json:"newPassword"`
	}

	// The RestoreAccountResponse type represents the response body returned when calling AccountAPI.Restore
	RestoreAccountResponse struct {
		// The account's new restore key.
		RestoreKey []byte `json:"restoreKey"`
	}
)

// Validate the request.
func (r RestoreAccountRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email, validation.Required, is.Email),
		validation.Field(&r.RestoreKey, validation.Required),
		validation.Field(&r.NewPassword, validation.Required),
	)
}

// Restore handles an inbound HTTP request to change an account's password using its restore key. On success, it responds
// with an http.StatusOK code and a JSON-encoded RestoreAccountResponse.
func (api *AccountAPI) Restore(w http.ResponseWriter, r *http.Request) {
	request, err := decode[RestoreAccountRequest](r.Body)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "failed to decode request: %v", err)
		return
	}

	restoreKey, err := api.accounts.Restore(request.Email, request.RestoreKey, request.NewPassword)
	switch {
	case errors.Is(err, service.ErrAccountNotFound):
		writeError(w, http.StatusNotFound, "account not found")
		return
	case errors.Is(err, service.ErrInvalidRestoreKey):
		writeError(w, http.StatusBadRequest, "invalid restore key for account")
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to restore account %v", err)
		return
	}

	write(w, http.StatusOK, RestoreAccountResponse{
		RestoreKey: restoreKey,
	})
}
