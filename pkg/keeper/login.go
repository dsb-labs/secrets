package keeper

import (
	"context"
	"net/http"
	"path"

	"github.com/davidsbond/keeper/internal/server/api"
)

type (
	// The Login type represents a single username/password combination stored for a user.
	Login struct {
		// The login's unique identifier.
		ID string
		// The username for this login.
		Username string
		// The password for this login.
		Password string
		// Domains where this login can be used.
		Domains []string
	}
)

// CreateLogin attempts to create a new login record for the authenticated user.
func (c *Client) CreateLogin(ctx context.Context, login Login) error {
	request, err := c.buildRequest(ctx, http.MethodPost, "/api/v1/login", api.CreateLoginRequest{
		Username: login.Username,
		Password: login.Password,
		Domains:  login.Domains,
	})
	if err != nil {
		return err
	}

	if _, err = doRequest[api.CreateLoginResponse](c.client, request); err != nil {
		return err
	}

	return nil
}

// ListLogins attempts to return all login records stored for the authenticated user. If the "domain" parameter is set,
// the server will filter the results to credentials that may be usable on the domain.
func (c *Client) ListLogins(ctx context.Context, domain string) ([]Login, error) {
	p := "/api/v1/login"
	if domain != "" {
		p += "?domain=" + domain
	}

	request, err := c.buildRequest(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	response, err := doRequest[api.ListLoginsResponse](c.client, request)
	if err != nil {
		return nil, err
	}

	logins := make([]Login, len(response.Logins))
	for i, login := range response.Logins {
		logins[i] = Login{
			ID:       login.ID,
			Username: login.Username,
			Password: login.Password,
			Domains:  login.Domains,
		}
	}

	return logins, nil
}

// DeleteLogin attempts to delete the login record with the specified id for the authenticated user.
func (c *Client) DeleteLogin(ctx context.Context, id string) error {
	request, err := c.buildRequest(ctx, http.MethodDelete, path.Join("/api/v1/login", id), nil)
	if err != nil {
		return err
	}

	if _, err = doRequest[api.DeleteLoginResponse](c.client, request); err != nil {
		return err
	}

	return nil
}

// GetLogin attempts to obtain the login record with the specified id for the authenticated user.
func (c *Client) GetLogin(ctx context.Context, id string) (Login, error) {
	request, err := c.buildRequest(ctx, http.MethodGet, path.Join("/api/v1/login", id), nil)
	if err != nil {
		return Login{}, err
	}

	response, err := doRequest[api.GetLoginResponse](c.client, request)
	if err != nil {
		return Login{}, err
	}

	return Login{
		ID:       response.Login.ID,
		Username: response.Login.Username,
		Password: response.Login.Password,
		Domains:  response.Login.Domains,
	}, nil
}
