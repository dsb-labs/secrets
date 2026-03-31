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
		// The user's payment cards.
		Cards []Card
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
				ID:        login.ID,
				Username:  login.Username,
				Password:  login.Password,
				Domains:   login.Domains,
				CreatedAt: login.CreatedAt,
				Name:      login.Name,
			}
		}),
		Notes: convert.Slice(response.Notes, func(note api.Note) Note {
			return Note{
				ID:        note.ID,
				Name:      note.Name,
				Content:   note.Content,
				CreatedAt: note.CreatedAt,
			}
		}),
		Cards: convert.Slice(response.Cards, func(card api.Card) Card {
			return Card{
				ID:          card.ID,
				HolderName:  card.HolderName,
				Number:      card.Number,
				ExpiryMonth: card.ExpiryMonth,
				ExpiryYear:  card.ExpiryYear,
				CVV:         card.CVV,
				CreatedAt:   card.CreatedAt,
				Name:        card.Name,
			}
		}),
	}, nil
}
