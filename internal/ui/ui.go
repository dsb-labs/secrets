// Package ui provides HTTP handlers for serving the application's web interface.
package ui

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/a-h/templ"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-playground/form/v4"
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

	// The Validatable interface describes types that can validate their own fields, returning an error if any
	// field values are invalid.
	Validatable interface {
		// Validate should return an error if any field values are invalid.
		Validate() error
	}
)

func redirect(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, path, http.StatusFound)
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	target := r.URL.RequestURI()
	redirect(w, r, "/login?redirect="+url.QueryEscape(target))
}

func render[T any](ctx context.Context, w io.Writer, view View[T], model T) {
	err := view(model).Render(ctx, w)
	if err != nil {
		panic(err)
	}
}

var decoder = form.NewDecoder()

func decode[T Validatable](r *http.Request) (T, error) {
	var body T
	if err := r.ParseForm(); err != nil {
		return body, err
	}

	if err := decoder.Decode(&body, r.PostForm); err != nil {
		return body, err
	}
	
	if err := body.Validate(); err != nil {
		return body, err
	}

	return body, nil
}

func validationErrors(ve validation.Errors) map[string]string {
	out := make(map[string]string, len(ve))
	for k, v := range ve {
		out[k] = v.Error()
	}
	return out
}

func requireToken(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !token.FromContext(r.Context()).Valid() {
			redirectToLogin(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
