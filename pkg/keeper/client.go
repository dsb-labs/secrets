// Package keeper provides the go client for the keeper api.
package keeper

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/davidsbond/keeper/internal/server/api"
)

type (
	// The Client type is responsible for managing API requests made against the keeper API.
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
