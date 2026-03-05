package ui

import (
	"net/http"

	"github.com/davidsbond/keeper/internal/server/token"
	"github.com/davidsbond/keeper/internal/server/ui/view"
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
	mux.HandleFunc("GET /dashboard", h.Dashboard)
}

// Dashboard renders the dashboard view.
func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	tkn := token.FromContext(r.Context())
	if !tkn.Valid() {
		redirect(w, r, "/login")
		return
	}

	render(r.Context(), w, "Dashboard", view.Dashboard, view.DashboardViewModel{})
}
