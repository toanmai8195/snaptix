package httpx

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// fakeDB thoả pinger: trả err cố định, hoặc chờ tới khi ctx hết hạn nếu block = true.
type fakeDB struct {
	err   error
	block bool
}

func (f fakeDB) Ping(ctx context.Context) error {
	if f.block {
		<-ctx.Done()
		return ctx.Err()
	}
	return f.err
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestRouter(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		db         fakeDB
		wantStatus int
		wantBody   string
	}{
		{"healthz ok", http.MethodGet, "/healthz", fakeDB{}, http.StatusOK, "ok"},
		{"healthz không phụ thuộc DB", http.MethodGet, "/healthz", fakeDB{err: errors.New("down")}, http.StatusOK, "ok"},
		{"readyz DB ok", http.MethodGet, "/readyz", fakeDB{}, http.StatusOK, "ok"},
		{"readyz DB lỗi", http.MethodGet, "/readyz", fakeDB{err: errors.New("down")}, http.StatusServiceUnavailable, "db unavailable\n"},
		{"path không có", http.MethodGet, "/khong-co", fakeDB{}, http.StatusNotFound, "404 page not found\n"},
		{"sai method", http.MethodPost, "/healthz", fakeDB{}, http.StatusMethodNotAllowed, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter(discardLogger(), tt.db)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, muốn %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Errorf("body = %q, muốn %q", got, tt.wantBody)
			}
		})
	}
}

func TestReadyzTimeout(t *testing.T) {
	router := NewRouter(discardLogger(), fakeDB{block: true})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	start := time.Now()
	router.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, muốn %d", rec.Code, http.StatusServiceUnavailable)
	}
	if elapsed < readyTimeout || elapsed > readyTimeout+time.Second {
		t.Errorf("readyz mất %v, muốn ≈ %v", elapsed, readyTimeout)
	}
}
