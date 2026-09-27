// Package httpx gom phần HTTP dùng chung trong core: router, middleware, health check.
package httpx

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter trả về handler gốc của core.
func NewRouter(logger *slog.Logger, db pinger) http.Handler {
	r := chi.NewRouter()
	r.Use(logRequests(logger))
	r.Get("/healthz", healthz)
	r.Get("/readyz", readyz(db, logger))
	return r
}
