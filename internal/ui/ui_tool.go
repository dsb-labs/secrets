package ui

import (
	"net/http"

	"github.com/davidsbond/keeper/internal/server/token"
	toolview "github.com/davidsbond/keeper/internal/ui/view/tool"
	statusview "github.com/davidsbond/keeper/internal/ui/view/status"
)

// The ToolHandler type is responsible for serving web interface pages for the tools section.
type ToolHandler struct {
	accounts AccountService
}

// NewToolHandler returns a new instance of the ToolHandler type that will serve tool UIs using the provided
// AccountService implementation.
func NewToolHandler(accounts AccountService) *ToolHandler {
	return &ToolHandler{accounts: accounts}
}

// Register HTTP endpoints onto the provided http.ServeMux.
func (h *ToolHandler) Register(mux *http.ServeMux) {
	mux.Handle("GET /tools", requireToken(h.Index))
}

// Index renders the tools index view.
func (h *ToolHandler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, toolview.Index, toolview.ViewModel{
		DisplayName: account.DisplayName,
	})
}
