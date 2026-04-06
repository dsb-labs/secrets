// Package secrets provides the go client for the secrets api.
package secrets

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/dsb-labs/secrets/internal/server/api"
)

type (
	// The Client type is responsible for managing API requests made against the secrets API.
	Client struct {
		baseURL string
		client  *http.Client

		mux   sync.RWMutex
		token string
	}
)

// NewClient returns a new instance of the Client type that will make requests against the provided base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: time.Minute,
		},
	}
}

// SetToken sets the authentication token. This token will be sent as a bearer token in each subsequent request.
func (c *Client) SetToken(token string) {
	c.mux.Lock()
	c.token = token
	c.mux.Unlock()
}

// Token returns the current authentication token.
func (c *Client) Token() string {
	c.mux.RLock()
	defer c.mux.RUnlock()

	return c.token
}

func (c *Client) buildRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	buffer := bytes.NewBuffer(nil)
	if body != nil {
		if err := json.NewEncoder(buffer).Encode(body); err != nil {
			return nil, err
		}
	}

	r, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, buffer)
	if err != nil {
		return nil, err
	}

	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		c.mux.RLock()
		r.Header.Set("Authorization", "Bearer "+c.token)
		c.mux.RUnlock()
	}

	return r, nil
}

func doRequest[T any](client *http.Client, r *http.Request) (T, error) {
	var body T

	response, err := client.Do(r)
	if err != nil {
		return body, err
	}

	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	if response.StatusCode >= http.StatusMultipleChoices {
		var e api.Error
		if err = decoder.Decode(&e); err != nil {
			return body, err
		}

		return body, e
	}

	if err = decoder.Decode(&body); err != nil {
		return body, err
	}

	return body, nil
}

// IsNotFound returns true if the provided error is of type api.Error and has an http.StatusNotFound status code.
func IsNotFound(err error) bool {
	var e api.Error
	return errors.As(err, &e) && e.Code == http.StatusNotFound
}

// IsConflict returns true if the provided error is of type api.Error and has an http.StatusConflict status code.
func IsConflict(err error) bool {
	var e api.Error
	return errors.As(err, &e) && e.Code == http.StatusConflict
}

// IsUnauthorized returns true if the provided error is of type api.Error and has an http.StatusUnauthorized status code.
func IsUnauthorized(err error) bool {
	var e api.Error
	return errors.As(err, &e) && e.Code == http.StatusUnauthorized
}

// IsBadRequest returns true if the provided error is of type api.Error and has an http.StatusBadRequest status code.
func IsBadRequest(err error) bool {
	var e api.Error
	return errors.As(err, &e) && e.Code == http.StatusBadRequest
}
