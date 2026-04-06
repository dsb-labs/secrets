package ui

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/davidsbond/x/filter"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"

	"github.com/dsb-labs/secrets/internal/server/password"
	"github.com/dsb-labs/secrets/internal/server/service"
	"github.com/dsb-labs/secrets/internal/server/token"
	loginview "github.com/dsb-labs/secrets/internal/ui/view/login"
	statusview "github.com/dsb-labs/secrets/internal/ui/view/status"
)

type (
	// The LoginHandler type is responsible for serving web interface pages regarding user logins.
	LoginHandler struct {
		accounts AccountService
		logins   LoginService
	}

	// The LoginService interface describes types that manage user login records.
	LoginService interface {
		// Create should store a new login record for the given user.
		Create(accountID uuid.UUID, login service.Login) error
		// List should return all logins associated with the given user id.
		List(accountID uuid.UUID, filters ...filter.Filter[service.Login]) ([]service.Login, error)
		// Get should return the login record associated with the given user and login identifiers.
		Get(accountID uuid.UUID, loginID uuid.UUID) (service.Login, error)
		// Delete should remove the login record associated with the given user and login identifiers.
		Delete(accountID uuid.UUID, loginID uuid.UUID) error
		// ListReusedPasswords should return all logins that share a password with at least one other login.
		ListReusedPasswords(accountID uuid.UUID, filters ...filter.Filter[service.Login]) ([]service.Login, error)
		// ListSamePassword should return all logins that share the same password as the given login.
		ListSamePassword(accountID uuid.UUID, loginID uuid.UUID) ([]service.Login, error)
		// ListWeakPasswords should return all logins whose password is rated below Good.
		ListWeakPasswords(accountID uuid.UUID, filters ...filter.Filter[service.Login]) ([]service.Login, error)
	}
)

// NewLoginHandler returns a new instance of the LoginHandler type that will serve login management UIs using
// the provided service implementations.
func NewLoginHandler(accounts AccountService, logins LoginService) *LoginHandler {
	return &LoginHandler{accounts: accounts, logins: logins}
}

// Register HTTP endpoints onto the provided http.ServeMux.
func (h *LoginHandler) Register(mux *http.ServeMux) {
	mux.Handle("GET /logins", requireToken(h.List))
	mux.Handle("GET /logins/new", requireToken(h.Create))
	mux.Handle("GET /logins/reused", requireToken(h.Reused))
	mux.Handle("GET /logins/weak", requireToken(h.Weak))
	mux.Handle("POST /logins", requireToken(h.CreateCallback))
	mux.Handle("GET /logins/{id}", requireToken(h.Detail))
	mux.Handle("POST /logins/{id}/delete", requireToken(h.Delete))
}

// List renders the login list view.
func (h *LoginHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	query := r.URL.Query().Get("query")
	filters := make([]filter.Filter[service.Login], 0)
	if query != "" {
		filters = append(filters,
			service.LoginsByDomain(query),
			service.LoginsByName(query),
		)
	}

	results, err := h.logins.List(tkn.ID(), filters...)
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

	items := make([]loginview.Item, len(results))
	for i, l := range results {
		items[i] = loginview.Item{
			ID:       l.ID.String(),
			Username: l.Username,
			Domains:  l.Domains,
			Name:     l.Name,
		}
	}

	render(ctx, w, http.StatusOK, loginview.List, loginview.ViewModel{
		DisplayName: account.DisplayName,
		Logins:      items,
		Query:       query,
	})
}

// Reused renders the reused passwords view, listing all logins that share a password with at least one other.
func (h *LoginHandler) Reused(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	query := r.URL.Query().Get("query")
	filters := make([]filter.Filter[service.Login], 0)
	if query != "" {
		filters = append(filters,
			service.LoginsByDomain(query),
			service.LoginsByName(query),
		)
	}

	results, err := h.logins.ListReusedPasswords(tkn.ID(), filters...)
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

	items := make([]loginview.Item, len(results))
	for i, l := range results {
		items[i] = loginview.Item{
			ID:       l.ID.String(),
			Username: l.Username,
			Domains:  l.Domains,
			Name:     l.Name,
		}
	}

	render(ctx, w, http.StatusOK, loginview.Reused, loginview.ReusedViewModel{
		DisplayName: account.DisplayName,
		Logins:      items,
		Query:       query,
	})
}

// Weak renders the weak passwords view, listing all logins whose password is rated below Good.
func (h *LoginHandler) Weak(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	query := r.URL.Query().Get("query")
	filters := make([]filter.Filter[service.Login], 0)
	if query != "" {
		filters = append(filters,
			service.LoginsByDomain(query),
			service.LoginsByName(query),
		)
	}

	results, err := h.logins.ListWeakPasswords(tkn.ID(), filters...)
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

	items := make([]loginview.Item, len(results))
	for i, l := range results {
		items[i] = loginview.Item{
			ID:       l.ID.String(),
			Username: l.Username,
			Domains:  l.Domains,
			Name:     l.Name,
		}
	}

	render(ctx, w, http.StatusOK, loginview.Weak, loginview.WeakViewModel{
		DisplayName: account.DisplayName,
		Logins:      items,
		Query:       query,
	})
}

// Detail renders the login detail view for a single login record.
func (h *LoginHandler) Detail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	loginID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		render(ctx, w, http.StatusNotFound, statusview.NotFound, statusview.NotFoundViewModel{})
		return
	}

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	login, err := h.logins.Get(tkn.ID(), loginID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case errors.Is(err, service.ErrLoginNotFound):
		render(ctx, w, http.StatusNotFound, statusview.NotFound, statusview.NotFoundViewModel{})
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	shared, err := h.logins.ListSamePassword(tkn.ID(), loginID)
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

	render(ctx, w, http.StatusOK, loginview.Detail, loginview.DetailViewModel{
		DisplayName:         account.DisplayName,
		ID:                  login.ID.String(),
		Username:            login.Username,
		Password:            login.Password,
		Domains:             login.Domains,
		CreatedAt:           login.CreatedAt.Format("2 January 2006 at 15:04"),
		SharedPasswordCount: len(shared),
		PasswordRating:      password.Rate(login.Password),
		Name:                login.Name,
	})
}

// Delete handles a login deletion request, redirecting to the login list on success.
func (h *LoginHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	loginID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		render(ctx, w, http.StatusNotFound, statusview.NotFound, statusview.NotFoundViewModel{})
		return
	}

	err = h.logins.Delete(tkn.ID(), loginID)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case errors.Is(err, service.ErrLoginNotFound):
		render(ctx, w, http.StatusNotFound, statusview.NotFound, statusview.NotFoundViewModel{})
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	redirect(w, r, "/logins")
}

// Create renders the login creation form.
func (h *LoginHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, loginview.Create, loginview.CreateViewModel{
		DisplayName: account.DisplayName,
	})
}

// The CreateLoginForm type represents the form values submitted when calling LoginHandler.CreateCallback.
type CreateLoginForm struct {
	// The optional name for the login.
	Name string `form:"name"`
	// The username.
	Username string `form:"username"`
	// The password.
	Password string `form:"password"`
	// The domains where this username/password combination can be used, as a comma-separated string.
	Domains string `form:"domains"`
}

// Validate the form.
func (f CreateLoginForm) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Username, validation.Required),
		validation.Field(&f.Password, validation.Required),
	)
}

// CreateCallback handles the login creation form submission, redirecting to the login detail view on success.
func (h *LoginHandler) CreateCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	form, err := decode[CreateLoginForm](r)
	model := loginview.CreateViewModel{
		DisplayName: account.DisplayName,
		Name:        form.Name,
		Username:    form.Username,
		Password:    form.Password,
		Domains:     form.Domains,
	}

	var ve validation.Errors
	switch {
	case errors.As(err, &ve):
		model.Validation.Errors = validationErrors(ve)
		render(ctx, w, http.StatusUnprocessableEntity, loginview.Create, model)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	var domains []string
	for d := range strings.SplitSeq(form.Domains, ",") {
		if strings.TrimSpace(d) == "" {
			continue
		}

		domains = append(domains, d)
	}

	loginID := uuid.New()
	err = h.logins.Create(tkn.ID(), service.Login{
		ID:        loginID,
		Name:      form.Name,
		Username:  form.Username,
		Password:  form.Password,
		Domains:   domains,
		CreatedAt: time.Now(),
	})
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

	redirect(w, r, "/logins/"+loginID.String())
}
