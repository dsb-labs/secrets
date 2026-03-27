package ui

import (
	"net/http"

	"github.com/davidsbond/keeper/internal/server/token"
	"github.com/davidsbond/keeper/internal/ui/view/dashboard"
	statusview "github.com/davidsbond/keeper/internal/ui/view/status"
)

// The DashboardHandler type is responsible for serving web interface pages regarding the user dashboard.
type DashboardHandler struct {
	accounts AccountService
}

// NewDashboardHandler returns a new instance of the DashboardHandler type that will serve dashboard UIs using
// the provided AccountService implementation.
func NewDashboardHandler(accounts AccountService) *DashboardHandler {
	return &DashboardHandler{accounts: accounts}
}

// Register HTTP endpoints onto the provided http.ServeMux.
func (h *DashboardHandler) Register(mux *http.ServeMux) {
	mux.Handle("GET /dashboard", requireToken(h.Dashboard))
}

// Dashboard renders the dashboard view.
func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, dashboard.Dashboard, dashboard.ViewModel{
		DisplayName: account.DisplayName,
	})
}
