// Package ui provides HTTP handlers for serving the application's web interface.
package ui

import (
	"context"
	"io"
	"net/http"

	"github.com/a-h/templ"
	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
)

type (
	// View is a function that produces a templ.Component from a view model.
	View[T any] func(model T) templ.Component

	// The AccountService interface describes types that manage user accounts. It is shared across handlers that
	// need to read or create account data.
	AccountService interface {
		// Get should return the account associated with the given identifier.
		Get(id uuid.UUID) (service.Account, error)
		// Create should create the given account, returning the restore key to be presented to the user. If an
		// account with the given email already exists, service.ErrAccountExists should be returned.
		Create(account service.Account) ([]byte, error)
	}
)

func redirect(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, path, http.StatusFound)
}

func render[T any](ctx context.Context, w io.Writer, view View[T], model T) {
	err := view(model).Render(ctx, w)
	if err != nil {
		panic(err)
	}
}

func requireToken(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !token.FromContext(r.Context()).Valid() {
			redirect(w, r, "/login")
			return
		}

		next.ServeHTTP(w, r)
	})
}
