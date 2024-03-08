package api

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"gomess/internal/handler"
	"gomess/pkg/config"
	"net/http"
	"sync"
	"time"
)

var (
	cleanupWaitGroup sync.WaitGroup
)

type API struct {
	Handler      http.Handler
	Version      string
	OverrideTime func() time.Time
}

func (a *API) Now() time.Time {
	if a.OverrideTime != nil {
		return a.OverrideTime()
	}
	return time.Now()
}

func NewAPIWithVersion(ctx context.Context, conf *config.GlobalConfiguration, version string) *API {
	api := &API{Version: version}
	r := chi.NewRouter()
	h := handler.NewHandler()
	ws := handler.NewWsHandler()

	corsHandler := cors.New(cors.Options{
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowedHeaders:   conf.CORS.AllAllowedHeaders([]string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}),
		ExposedHeaders:   []string{"X-Total-Count", "Link"},
		AllowCredentials: true,
	})

	r.Group(func(r chi.Router) {
		r.Get("/health", h.HealthCheck)
	})
	r.HandleFunc("/apiws", ws.WsServe)

	api.Handler = corsHandler.Handler(r)
	return api

}

func WaitForCleanup(ctx context.Context) {
	cleanupDone := make(chan struct{})

	go func() {
		defer close(cleanupDone)
		cleanupWaitGroup.Wait()
	}()

	select {
	case <-ctx.Done():
		return

	case <-cleanupDone:
		return
	}
}
