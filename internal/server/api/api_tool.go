package api

import (
	"errors"
	"net/http"

	"github.com/davidsbond/x/convert"
	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
)

type (
	// The ToolAPI exposes HTTP endpoints for common user tools, such as import/export.
	ToolAPI struct {
		tools ToolService
	}

	// The ToolService interface describes types that provide user tool implementations.
	ToolService interface {
		// Export should return all the specified user's data as a service.Export type.
		Export(uuid.UUID) (service.Export, error)
	}
)

// NewToolAPI returns a new instance of the ToolAPI type that provider user tools via the given ToolService
// implementation.
func NewToolAPI(tools ToolService) *ToolAPI {
	return &ToolAPI{
		tools: tools,
	}
}

// Register the HTTP endpoints onto the given http.ServeMux.
func (api *ToolAPI) Register(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/export", requireToken(api.Export))
}

type (
	// The ExportResponse type represents the response body returned when calling ToolAPI.Export
	ExportResponse struct {
		// The user's logins.
		Logins []Login `json:"logins"`
		// The user's notes.
		Notes []Note `json:"notes"`
		// The user's payment cards.
		Cards []Card `json:"cards"`
	}
)

// Export handles an inbound HTTP request to export all of a user's data. On success, it responds with an http.StatusOK
// code and a JSON-encoded ExportResponse.
func (api *ToolAPI) Export(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	export, err := api.tools.Export(tkn.ID())
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	case err != nil:
		writeErrorf(w, http.StatusInternalServerError, "failed to export: %v", err)
		return
	}

	write(w, http.StatusOK, ExportResponse{
		Logins: convert.Slice(export.Logins, func(in service.Login) Login {
			return Login{
				ID:       in.ID.String(),
				Username: in.Username,
				Password: in.Password,
				Domains:  in.Domains,
			}
		}),
		Notes: convert.Slice(export.Notes, func(in service.Note) Note {
			return Note{
				ID:      in.ID.String(),
				Name:    in.Name,
				Content: in.Content,
			}
		}),
		Cards: convert.Slice(export.Cards, func(in service.Card) Card {
			return Card{
				ID:          in.ID.String(),
				HolderName:  in.HolderName,
				Number:      in.Number,
				ExpiryMonth: in.ExpiryMonth,
				ExpiryYear:  in.ExpiryYear,
				CVV:         in.CVV,
			}
		}),
	})
}
