package rest

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed web/*
var statics embed.FS

func BuildHTTPRouter() http.Handler {
	route := chi.NewRouter()
	route.Use(middleware.Logger)
	route.Get("/", Echo)

	content, err := fs.Sub(statics, "web")
	if err != nil {
		panic(err)
	}
	route.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(content))))
	return route
}
