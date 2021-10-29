package ultron

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	chimiddleware "github.com/jacexh/gopkg/chi-middleware"
)

//go:embed web/*
var statics embed.FS

func homepage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello, ultron"))
}

func buildHTTPRouter() http.Handler {
	route := chi.NewRouter()

	{ // middlewares
		route.Use(middleware.RequestID)
		route.Use(middleware.RealIP)
		route.Use(chimiddleware.RequestZapLog(Logger))
		route.Use(middleware.Recoverer)
	}

	// http api
	{
		route.Get("/", homepage)
	}

	// static files
	content, err := fs.Sub(statics, "web")
	if err != nil {
		panic(err)
	}
	route.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(content))))
	return route
}
