package httpx

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// readyTimeout: /readyz không được treo khi DB không phản hồi.
const readyTimeout = 2 * time.Second

// pinger là thứ /readyz cần từ DB — *pgxpool.Pool thoả mãn, test dùng fake.
type pinger interface {
	Ping(ctx context.Context) error
}

// healthz (liveness): process còn sống và phục vụ được HTTP. Không kiểm tra dependency.
func healthz(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, "ok") // lỗi ghi = client đã ngắt, không còn gì để làm
}

// readyz (readiness): sẵn sàng nhận traffic — ping được DB trong readyTimeout.
func readyz(db pinger, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readyTimeout)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			logger.Warn("readyz: db ping failed", "err", err)
			http.Error(w, "db unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = io.WriteString(w, "ok")
	}
}
