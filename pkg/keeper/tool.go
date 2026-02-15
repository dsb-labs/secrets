package keeper

import (
	"context"
	"net/http"

	"github.com/davidsbond/x/convert"

	"github.com/davidsbond/keeper/internal/server/api"
)

type (
	// The Export type represents the entire contents of a user's keeper database.
	Export struct {
		// The user's logins.
		Logins []Login
		// The user's notes.
		Notes []Note
	}
)

// Export the caller's entire keeper database.
func (c *Client) Export(ctx context.Context) (Export, error) {
	request, err := c.buildRequest(ctx, http.MethodGet, "/api/v1/export", nil)
	if err != nil {
		return Export{}, err
	}

	response, err := doRequest[api.ExportResponse](c.client, request)
	if err != nil {
		return Export{}, err
	}

	return Export{
		Logins: convert.Slice(response.Logins, func(login api.Login) Login {
			return Login{
				ID:       login.ID,
				Username: login.Username,
				Password: login.Password,
				Domains:  login.Domains,
			}
		}),
		Notes: convert.Slice(response.Notes, func(note api.Note) Note {
			return Note{
				ID:      note.ID,
				Name:    note.Name,
				Content: note.Content,
			}
		}),
	}, nil
}
