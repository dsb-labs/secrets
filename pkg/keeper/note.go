package keeper

import (
	"context"
	"net/http"
	"path"
	"time"

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
		// When the note was created.
		CreatedAt time.Time
	}
)

// CreateNote attempts to create a new note record for the authenticated user, returning its identifier on success.
func (c *Client) CreateNote(ctx context.Context, note Note) (string, error) {
	request, err := c.buildRequest(ctx, http.MethodPost, "/api/v1/note", api.CreateNoteRequest{
		Name:    note.Name,
		Content: note.Content,
	})
	if err != nil {
		return "", err
	}

	response, err := doRequest[api.CreateNoteResponse](c.client, request)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

type (
	// The NoteListOptions type contains fields used to filter the results of listing note records.
	NoteListOptions struct {
		// The query to match notes to, checked against name and content.
		Query string
	}
)

// ListNotes attempts to return all note records stored for the authenticated user. The NoteListOptions struct
// can be used to filter by name or content.
func (c *Client) ListNotes(ctx context.Context, options NoteListOptions) ([]Note, error) {
	p := "/api/v1/note"
	if options.Query != "" {
		p += "?query=" + options.Query
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
			ID:        in.ID,
			Name:      in.Name,
			Content:   in.Content,
			CreatedAt: in.CreatedAt,
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
		ID:        response.Note.ID,
		Name:      response.Note.Name,
		Content:   response.Note.Content,
		CreatedAt: response.Note.CreatedAt,
	}, nil
}
