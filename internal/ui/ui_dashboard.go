package ui

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
	"github.com/davidsbond/keeper/internal/ui/view"
)

type (
	// The DashboardHandler type is responsible for serving web interface pages regarding the user dashboard.
	DashboardHandler struct {
		accounts DashboardAccountService
	}

	// The DashboardAccountService interface describes types that provide account details for the dashboard.
	DashboardAccountService interface {
		// Get should return the account associated with the given identifier.
		Get(id uuid.UUID) (service.Account, error)
	}
)

// NewDashboardHandler returns a new instance of the DashboardHandler type that will serve dashboard UIs using
// the provided DashboardAccountService implementation.
func NewDashboardHandler(accounts DashboardAccountService) *DashboardHandler {
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
		http.Error(w, "failed to load account", http.StatusInternalServerError)
		return
	}

	render(ctx, w, view.Dashboard, view.DashboardViewModel{
		DisplayName: account.DisplayName,
	})
}
