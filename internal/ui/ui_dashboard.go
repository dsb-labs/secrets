package ui

import (
	"net/http"

	"github.com/davidsbond/keeper/internal/ui/view"
)

type (
	// The DashboardHandler type is responsible for serving web interface pages regarding the user dashboard.
	DashboardHandler struct {
	}
)

// NewDashboardHandler returns a new instance of the DashboardHandler type that will serve dashboard UIs.
func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

// Register HTTP endpoints onto the provided http.ServeMux.
func (h *DashboardHandler) Register(mux *http.ServeMux) {
	mux.Handle("GET /dashboard", requireToken(h.Dashboard))
}

// Dashboard renders the dashboard view.
func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	render(r.Context(), w, view.Dashboard, view.DashboardViewModel{})
}
