package httpx

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// logRequests ghi một dòng log cho mỗi request sau khi handler trả về.
func logRequests(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			// Bọc w để biết handler đã ghi status nào (ResponseWriter không có hàm đọc status).
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 { // handler không ghi gì → net/http trả 200
				status = http.StatusOK
			}
			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"duration", time.Since(start),
			)
		})
	}
}
