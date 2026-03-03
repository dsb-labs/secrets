// Package ui provides HTTP handlers for serving the application's web interface.
package ui

import (
	"context"
	"io"
	"net/http"

	"github.com/a-h/templ"

	"github.com/davidsbond/keeper/internal/server/ui/layout"
)

func redirect(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, path, http.StatusFound)
}

func render[T any](ctx context.Context, w io.Writer, title string, view func(T) templ.Component, model T) {
	err := layout.
		Main(title, view(model)).
		Render(ctx, w)
	if err != nil {
		panic(err)
	}
}
