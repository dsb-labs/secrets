// Package ui provides HTTP handlers for serving the application's web interface.
package ui

import (
	"context"
	"io"
	"net/http"

	"github.com/a-h/templ"

	"github.com/davidsbond/keeper/internal/server/token"
)

type (
	View[T any] func(model T) templ.Component
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
