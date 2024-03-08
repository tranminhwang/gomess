package api

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"gomess/internal/handler"
	"gomess/pkg/config"
	"net/http"
	"time"
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
	h := handler.NewHandler()
	r := chi.NewRouter()

	corsHandler := cors.New(cors.Options{
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowedHeaders:   conf.CORS.AllAllowedHeaders([]string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}),
		ExposedHeaders:   []string{"X-Total-Count", "Link"},
		AllowCredentials: true,
	})

	r.Group(func(r chi.Router) {
		r.Get("/health", h.HealthCheck)
	})

	api.Handler = corsHandler.Handler(r)
	return api

}
