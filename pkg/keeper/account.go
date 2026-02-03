package keeper

import (
	"context"
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
)

// CreateAccount attempts to create a new account.
func (c *Client) CreateAccount(ctx context.Context, account Account) error {
	request, err := c.buildRequest(ctx, http.MethodPost, "/api/v1/account", api.CreateAccountRequest{
		Email:       account.Email,
		Password:    account.Password,
		DisplayName: account.DisplayName,
	})
	if err != nil {
		return err
	}

	if _, err = doRequest[api.CreateAccountResponse](c.client, request); err != nil {
		return err
	}

	return nil
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
