package keeper

import (
	"context"
	"net/http"

	"github.com/davidsbond/keeper/internal/server/api"
)

// Login attempts to obtain an authentication token for the given email and password combination. On success, a token
// is stored within the Client and can be accessed via the Client.Token method.
func (c *Client) Login(ctx context.Context, email, password string) error {
	request, err := c.buildRequest(ctx, http.MethodPost, "/api/v1/auth", api.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return err
	}

	response, err := doRequest[api.LoginResponse](c.client, request)
	if err != nil {
		return err
	}

	c.mux.Lock()
	c.token = response.Token
	c.mux.Unlock()

	return nil
}

// Logout invalidates the current authentication token.
func (c *Client) Logout(ctx context.Context) error {
	request, err := c.buildRequest(ctx, http.MethodDelete, "/api/v1/auth", nil)
	if err != nil {
		return err
	}

	if _, err = doRequest[api.LogoutResponse](c.client, request); err != nil {
		return err
	}

	c.mux.Lock()
	c.token = ""
	c.mux.Unlock()

	return nil
}
