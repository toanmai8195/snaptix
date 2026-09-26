// Package httpx dựng router HTTP của core: endpoint hạ tầng (health, metrics) và
// chỗ gắn route của các module nghiệp vụ.
package httpx

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Pinger là thứ /readyz cần: kiểm tra kết nối tới phụ thuộc (vd pgxpool.Pool).
type Pinger interface {
	Ping(ctx context.Context) error
}

// readyTimeout giới hạn thời gian /readyz chờ DB, để load balancer nhận 503 nhanh.
var readyTimeout = 2 * time.Second

// NewRouter trả về handler gốc của core.
func NewRouter(log *slog.Logger, db Pinger) http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), readyTimeout)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			log.WarnContext(ctx, "readyz: database unavailable", slog.Any("error", err))
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable", "error": "database"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Handle("/metrics", promhttp.Handler())
	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
