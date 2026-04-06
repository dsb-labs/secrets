package ui

import (
	"errors"
	"net/http"

	"github.com/dsb-labs/secrets/internal/server/service"
	"github.com/dsb-labs/secrets/internal/server/token"
	"github.com/dsb-labs/secrets/internal/ui/view/dashboard"
	statusview "github.com/dsb-labs/secrets/internal/ui/view/status"
)

// The DashboardHandler type is responsible for serving web interface pages regarding the user dashboard.
type DashboardHandler struct {
	accounts AccountService
	logins   LoginService
}

// NewDashboardHandler returns a new instance of the DashboardHandler type that will serve dashboard UIs using
// the provided AccountService and LoginService implementations.
func NewDashboardHandler(accounts AccountService, logins LoginService) *DashboardHandler {
	return &DashboardHandler{accounts: accounts, logins: logins}
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

	duplicates, err := h.logins.ListReusedPasswords(tkn.ID())
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	weak, err := h.logins.ListWeakPasswords(tkn.ID())
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, dashboard.Dashboard, dashboard.ViewModel{
		DisplayName:            account.DisplayName,
		DuplicatePasswordCount: len(duplicates),
		WeakPasswordCount:      len(weak),
	})
}
