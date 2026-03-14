package ui

import (
	"embed"
	"net/http"
)

type (
	AssetHandler struct {
	}
)

var (
	//go:embed asset
	assets embed.FS
)

func NewAssetHandler() *AssetHandler {
	return &AssetHandler{}
}

func (h *AssetHandler) Register(mux *http.ServeMux) {
	mux.Handle("GET /asset/", http.FileServer(http.FS(assets)))
}
