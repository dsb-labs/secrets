package ui

import (
	"net/http"
	"strings"

	statusview "github.com/dsb-labs/secrets/internal/ui/view/status"
)

// The NotFoundHandler type serves a 404 page for any route not matched by a more specific handler.
// Requests to paths prefixed with /api or /asset receive a plain 404 status instead.
type NotFoundHandler struct{}

// NewNotFoundHandler returns a new instance of the NotFoundHandler type.
func NewNotFoundHandler() *NotFoundHandler {
	return &NotFoundHandler{}
}

// Register HTTP endpoints onto the provided http.ServeMux.
func (h *NotFoundHandler) Register(mux *http.ServeMux) {
	mux.Handle("/", h)
}

func (h *NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/asset") {
		http.NotFound(w, r)
		return
	}

	render(r.Context(), w, http.StatusNotFound, statusview.NotFound, statusview.NotFoundViewModel{})
}
