package keeper

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/davidsbond/keeper/internal/server/api"
)

type (
	// The Account type represents a single user account.
	Account struct {
		// The user's email address.
		Email string
		// The user's display name.
		DisplayName string
		// The user's password, only set when calling Client.CreateAccount.
		Password string `json:"-"`
	}

	// The RestoreKey type represents a key that can be used to recover an account should the user forget their
	// password or in a data recovery scenario.
	RestoreKey []byte
)

func (k RestoreKey) String() string {
	return base64.StdEncoding.EncodeToString(k)
}

// CreateAccount attempts to create a new account, returning the account's restore key on success. This restore key
// must be saved by the caller to decrypt their database should they forget their password.
func (c *Client) CreateAccount(ctx context.Context, account Account) (RestoreKey, error) {
	request, err := c.buildRequest(ctx, http.MethodPost, "/api/v1/account", api.CreateAccountRequest{
		Email:       account.Email,
		Password:    account.Password,
		DisplayName: account.DisplayName,
	})
	if err != nil {
		return nil, err
	}

	response, err := doRequest[api.CreateAccountResponse](c.client, request)
	if err != nil {
		return nil, err
	}

	return response.RestoreKey, nil
}

// GetAccount attempts to return the account details for the authenticated user.
func (c *Client) GetAccount(ctx context.Context) (Account, error) {
	request, err := c.buildRequest(ctx, http.MethodGet, "/api/v1/account", nil)
	if err != nil {
		return Account{}, err
	}

	response, err := doRequest[api.GetAccountResponse](c.client, request)
	if err != nil {
		return Account{}, err
	}

	return Account{
		Email:       response.Account.Email,
		DisplayName: response.Account.DisplayName,
	}, nil
}

// ChangePassword attempts to update the caller's password to the new one. The old password must match the user's
// existing password. On success, returns the user's updated restore key which must be stored for disaster recovery
// purposes or if they forget their password.
func (c *Client) ChangePassword(ctx context.Context, oldPassword, newPassword string) (RestoreKey, error) {
	request, err := c.buildRequest(ctx, http.MethodPut, "/api/v1/account/password", api.UpdatePasswordRequest{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	})
	if err != nil {
		return nil, err
	}

	response, err := doRequest[api.UpdatePasswordResponse](c.client, request)
	if err != nil {
		return nil, err
	}

	return response.RestoreKey, nil
}

// DeleteAccount attempts to delete the caller's account.
func (c *Client) DeleteAccount(ctx context.Context) error {
	request, err := c.buildRequest(ctx, http.MethodDelete, "/api/v1/account", nil)
	if err != nil {
		return err
	}

	if _, err = doRequest[api.DeleteAccountResponse](c.client, request); err != nil {
		return err
	}

	return nil
}

// RestoreAccount attempts to update the account's password to the new one using their restore key. On success, returns
// the user's updated restore key which must be stored for disaster recovery purposes or if they forget their password.
func (c *Client) RestoreAccount(ctx context.Context, email string, restoreKey RestoreKey, newPassword string) (RestoreKey, error) {
	request, err := c.buildRequest(ctx, http.MethodPost, "/api/v1/account/restore", api.RestoreAccountRequest{
		Email:       email,
		RestoreKey:  restoreKey,
		NewPassword: newPassword,
	})
	if err != nil {
		return nil, err
	}

	response, err := doRequest[api.RestoreAccountResponse](c.client, request)
	if err != nil {
		return nil, err
	}

	return response.RestoreKey, nil
}
