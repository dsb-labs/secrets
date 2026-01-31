// Package ui provides the keeper user-interface via an http.Handler implementation.
package ui

import (
	"embed"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

type (
	// The Handler type is an http.Handler implementation that serves the application ui from an embedded file
	// system.
	Handler struct {
		assets fs.FS
	}
)

var (
	//go:embed dist/*
	assets embed.FS
)

// NewHandler returns a new instance of the Handler type.
func NewHandler() *Handler {
	sub, err := fs.Sub(assets, "dist")
	if err != nil {
		panic(err)
	}

	return &Handler{
		assets: sub,
	}
}

// Register the Handler onto the http.ServeMux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("/", h)
}

// ServeHTTP handles inbound HTTP requests for the user-interface. Unless an explicit file extension is present, it
// always serves the index.html page.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	location := strings.TrimPrefix(r.URL.Path, "/")
	ext := path.Ext(location)
	if ext == "" {
		location = "index.html"
		ext = ".html"
	}

	file, err := h.assets.Open(location)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer file.Close()

	w.Header().Set("Content-Type", mime.TypeByExtension(ext))
	if _, err = io.Copy(w, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
