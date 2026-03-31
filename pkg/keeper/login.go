package keeper

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/davidsbond/x/convert"

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
		// When the login was created.
		CreatedAt time.Time
		// A user supplied name for the login.
		Name string
	}
)

// CreateLogin attempts to create a new login record for the authenticated user, returning its identifier on success.
func (c *Client) CreateLogin(ctx context.Context, login Login) (string, error) {
	request, err := c.buildRequest(ctx, http.MethodPost, "/api/v1/login", api.CreateLoginRequest{
		Username: login.Username,
		Password: login.Password,
		Domains:  login.Domains,
		Name:     login.Name,
	})
	if err != nil {
		return "", err
	}

	response, err := doRequest[api.CreateLoginResponse](c.client, request)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

type (
	// The LoginListOptions type contains fields used to filter the results of listing login records.
	LoginListOptions struct {
		// The domain to match logins to.
		Domain string
		// The name to match logins to.
		Name string
	}
)

// ListLogins attempts to return all login records stored for the authenticated user. The LoginListOptions struct
// can be used to filter by name or domain.
func (c *Client) ListLogins(ctx context.Context, options LoginListOptions) ([]Login, error) {
	values := url.Values{}
	if options.Domain != "" {
		values.Add("domain", options.Domain)
	}
	if options.Name != "" {
		values.Add("name", options.Name)
	}

	p := "/api/v1/login"
	v := values.Encode()
	if v != "" {
		p += "?" + v
	}

	request, err := c.buildRequest(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	response, err := doRequest[api.ListLoginsResponse](c.client, request)
	if err != nil {
		return nil, err
	}

	return convert.Slice(response.Logins, func(in api.Login) Login {
		return Login{
			ID:        in.ID,
			Username:  in.Username,
			Password:  in.Password,
			Domains:   in.Domains,
			CreatedAt: in.CreatedAt,
			Name:      in.Name,
		}
	}), nil
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
		ID:        response.Login.ID,
		Username:  response.Login.Username,
		Password:  response.Login.Password,
		Domains:   response.Login.Domains,
		CreatedAt: response.Login.CreatedAt,
		Name:      response.Login.Name,
	}, nil
}
