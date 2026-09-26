package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/toanmai8195/snaptix/com/tm/server/pkg/otelx"
)

const headerRequestID = "X-Request-ID"

// Chỉ nhận request ID ngắn, an toàn để ghi log và trả lại header.
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// requestID dùng X-Request-ID gửi lên nếu hợp lệ, không thì sinh mới; trả lại qua header.
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(headerRequestID)
		if !validRequestID.MatchString(id) {
			id = newRequestID()
		}
		w.Header().Set(headerRequestID, id)
		next.ServeHTTP(w, r.WithContext(otelx.WithRequestID(r.Context(), id)))
	})
}

func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// recoverer biến panic trong handler thành 500, ghi log kèm stack, giữ server sống.
func recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				v := recover()
				if v == nil {
					return
				}
				if v == http.ErrAbortHandler { //nolint:errorlint // so sánh sentinel theo tài liệu net/http
					panic(v)
				}
				log.ErrorContext(r.Context(), "panic",
					slog.Any("panic", v), slog.String("stack", string(debug.Stack())))
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal"})
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// accessLog ghi một dòng cho mỗi request. Probe và /metrics ở mức DEBUG để không làm nhiễu log.
func accessLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}
			level := slog.LevelInfo
			switch {
			case status >= http.StatusInternalServerError:
				level = slog.LevelError
			case isInfraPath(r.URL.Path):
				level = slog.LevelDebug
			}
			log.Log(r.Context(), level, "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("route", routePattern(r)),
				slog.Int("status", status),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
			)
		})
	}
}

// routeLabel đặt tên span theo route pattern và gắn http.route cho span + metric của otelhttp.
// Pattern chỉ đầy đủ sau khi chi định tuyến xong, nên đọc sau next.ServeHTTP.
func routeLabel(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		pattern := routePattern(r)
		if pattern == "" {
			return
		}
		route := attribute.String("http.route", pattern)
		span := trace.SpanFromContext(r.Context())
		span.SetName(r.Method + " " + pattern)
		span.SetAttributes(route)
		if l, ok := otelhttp.LabelerFromContext(r.Context()); ok {
			l.Add(route)
		}
	})
}

func routePattern(r *http.Request) string {
	if rc := chi.RouteContext(r.Context()); rc != nil {
		return rc.RoutePattern()
	}
	return ""
}

func isInfraPath(p string) bool {
	return p == "/healthz" || p == "/readyz" || p == "/metrics"
}
