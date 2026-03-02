package keeper

import (
	"context"
	"net/http"
	"path"

	"github.com/davidsbond/x/convert"

	"github.com/davidsbond/keeper/internal/server/api"
)

type (
	// The Note type represents a single username/password combination stored for a user.
	Note struct {
		// The unique identifier of the note.
		ID string
		// The note's name.
		Name string
		// The note's contents
		Content string
	}
)

// CreateNote attempts to create a new note record for the authenticated user.
func (c *Client) CreateNote(ctx context.Context, note Note) error {
	request, err := c.buildRequest(ctx, http.MethodPost, "/api/v1/note", api.CreateNoteRequest{
		Name:    note.Name,
		Content: note.Content,
	})
	if err != nil {
		return err
	}

	if _, err = doRequest[api.CreateNoteResponse](c.client, request); err != nil {
		return err
	}

	return nil
}

// ListNotes attempts to return all note records stored for the authenticated user. If the "query" parameter is set,
// the server will filter the results to notes that contain the query string in their name or content.
func (c *Client) ListNotes(ctx context.Context, query string) ([]Note, error) {
	p := "/api/v1/note"
	if query != "" {
		p += "?query=" + query
	}

	request, err := c.buildRequest(ctx, http.MethodGet, p, nil)
	if err != nil {
		return nil, err
	}

	response, err := doRequest[api.ListNotesResponse](c.client, request)
	if err != nil {
		return nil, err
	}

	return convert.Slice(response.Notes, func(in api.Note) Note {
		return Note{
			ID:      in.ID,
			Name:    in.Name,
			Content: in.Content,
		}
	}), nil
}

// DeleteNote attempts to delete the note record with the specified id for the authenticated user.
func (c *Client) DeleteNote(ctx context.Context, id string) error {
	request, err := c.buildRequest(ctx, http.MethodDelete, path.Join("/api/v1/note", id), nil)
	if err != nil {
		return err
	}

	if _, err = doRequest[api.DeleteNoteResponse](c.client, request); err != nil {
		return err
	}

	return nil
}

// GetNote attempts to obtain the note record with the specified id for the authenticated user.
func (c *Client) GetNote(ctx context.Context, id string) (Note, error) {
	request, err := c.buildRequest(ctx, http.MethodGet, path.Join("/api/v1/note", id), nil)
	if err != nil {
		return Note{}, err
	}

	response, err := doRequest[api.GetNoteResponse](c.client, request)
	if err != nil {
		return Note{}, err
	}

	return Note{
		ID:      response.Note.ID,
		Name:    response.Note.Name,
		Content: response.Note.Content,
	}, nil
}
