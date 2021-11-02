package ultron

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	chimiddleware "github.com/jacexh/gopkg/chi-middleware"
	"go.uber.org/zap"
)

type (
	restServer struct {
		runner *masterRunner
	}

	restResponse struct {
		Result       bool   `json:"result,omitempty"`
		ErrorMessage string `json:"error_message,omitempty"`
	}
)

//go:embed web/*
var statics embed.FS

func homepage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello, ultron!"))
}

func (rest *restServer) handleStartNewPlan() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req []*V1StageConfig
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Logger.Error("failed to parse request body", zap.Error(err))
			renderResponse(err, rw, r)
			return
		}

		plan := NewPlan("")
		for _, stage := range req {
			plan.AddStages(stage)
		}
		err := rest.runner.StartPlan(plan)
		renderResponse(err, rw, r)
	}
}

func (rest *restServer) handleStopPlan() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rest.runner.StopPlan()
		renderResponse(nil, rw, r)
	}
}

func renderResponse(err error, w http.ResponseWriter, r *http.Request) {
	ret := &restResponse{}
	if err == nil {
		ret.Result = true
	} else {
		ret.ErrorMessage = err.Error()
	}

	data, err := json.Marshal(ret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func buildHTTPRouter(runner *masterRunner) http.Handler {
	route := chi.NewRouter()
	rest := &restServer{runner: runner}

	{ // middlewares
		route.Use(middleware.RequestID)
		route.Use(middleware.RealIP)
		route.Use(chimiddleware.RequestZapLog(Logger))
		route.Use(middleware.Recoverer)
	}

	// http api
	{
		route.Get("/", homepage)
		route.Post("/api/v1/plan", rest.handleStartNewPlan())
		route.Delete("/api/v1/plan", rest.handleStopPlan())
	}

	// static files
	content, err := fs.Sub(statics, "web")
	if err != nil {
		panic(err)
	}
	route.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(content))))
	return route
}
