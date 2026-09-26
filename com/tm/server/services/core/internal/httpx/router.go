// Package httpx dựng router HTTP của core: middleware chung, endpoint hạ tầng
// (health, metrics) và chỗ gắn route của các module nghiệp vụ.
package httpx

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Pinger là thứ /readyz cần: kiểm tra kết nối tới phụ thuộc (vd pgxpool.Pool).
type Pinger interface {
	Ping(ctx context.Context) error
}

// Deps là phụ thuộc để dựng router.
type Deps struct {
	Log *slog.Logger
	DB  Pinger
	// Routes: mỗi module nghiệp vụ gắn route của mình.
	Routes []func(chi.Router)
	// OTel: tuỳ chọn cho otelhttp (test truyền TracerProvider/MeterProvider riêng); rỗng = provider global.
	OTel []otelhttp.Option
}

// readyTimeout giới hạn thời gian /readyz chờ DB, để load balancer nhận 503 nhanh.
var readyTimeout = 2 * time.Second

// NewRouter trả về handler gốc của core.
//
// Thứ tự middleware (ngoài → trong): otelhttp → request ID → access log → recover → route label → handler.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(requestID, accessLog(d.Log), recoverer(d.Log), routeLabel)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), readyTimeout)
		defer cancel()
		if err := d.DB.Ping(ctx); err != nil {
			d.Log.WarnContext(ctx, "readyz: database unavailable", slog.Any("error", err))
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable", "error": "database"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Handle("/metrics", promhttp.Handler())

	for _, mount := range d.Routes {
		mount(r)
	}

	opts := append([]otelhttp.Option{
		// Probe và /metrics không tạo span/metric để không làm nhiễu dữ liệu.
		otelhttp.WithFilter(func(req *http.Request) bool { return !isInfraPath(req.URL.Path) }),
		otelhttp.WithSpanNameFormatter(func(_ string, req *http.Request) string { return req.Method }),
	}, d.OTel...)
	return otelhttp.NewHandler(r, "core", opts...)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
