package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func BuildHTTPRouter() http.Handler {
	route := chi.NewRouter()
	route.Use(middleware.Logger)
	route.Get("/", Echo)

	// server static files
	fs := http.FileServer(http.Dir("./web"))
	route.Handle("/static/", http.StripPrefix("/static/", fs))
	return route
}
