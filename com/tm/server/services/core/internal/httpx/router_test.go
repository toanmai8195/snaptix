package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type pingerFunc func(ctx context.Context) error

func (f pingerFunc) Ping(ctx context.Context) error { return f(ctx) }

var (
	pingOK   = pingerFunc(func(context.Context) error { return nil })
	pingFail = pingerFunc(func(context.Context) error { return errors.New("connection refused") })
	// pingHang chờ tới khi context hết hạn — giống DB treo.
	pingHang = pingerFunc(func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() })
)

func do(t *testing.T, h http.Handler, path string) (int, map[string]string, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	body := rec.Body.String()
	var m map[string]string
	_ = json.Unmarshal([]byte(body), &m)
	return rec.Code, m, body
}

func TestHealthAndReady(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	tests := []struct {
		name       string
		db         Pinger
		path       string
		wantCode   int
		wantStatus string
	}{
		{"healthz không phụ thuộc DB", pingFail, "/healthz", http.StatusOK, "ok"},
		{"readyz DB ok", pingOK, "/readyz", http.StatusOK, "ok"},
		{"readyz DB lỗi", pingFail, "/readyz", http.StatusServiceUnavailable, "unavailable"},
		{"readyz DB treo", pingHang, "/readyz", http.StatusServiceUnavailable, "unavailable"},
	}
	old := readyTimeout
	readyTimeout = 50 * time.Millisecond
	t.Cleanup(func() { readyTimeout = old })

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			code, m, body := do(t, NewRouter(log, tt.db), tt.path)
			if code != tt.wantCode || m["status"] != tt.wantStatus {
				t.Fatalf("%s = %d %s; muốn %d status=%s", tt.path, code, body, tt.wantCode, tt.wantStatus)
			}
			if d := time.Since(start); d > time.Second {
				t.Errorf("%s mất %v, phải trả nhanh theo readyTimeout", tt.path, d)
			}
		})
	}
}

func TestMetricsAndNotFound(t *testing.T) {
	h := NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), pingOK)

	code, _, body := do(t, h, "/metrics")
	if code != http.StatusOK || !strings.Contains(body, "go_goroutines") {
		t.Fatalf("/metrics = %d, thiếu go_goroutines", code)
	}
	if code, _, _ := do(t, h, "/khong-ton-tai"); code != http.StatusNotFound {
		t.Fatalf("route lạ = %d, muốn 404", code)
	}
}
