// Command server là HTTP server của core service.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/toanmai8195/snaptix/com/tm/server/services/core/internal/httpx"
)

func main() {
	cfg, err := loadConfig(os.Getenv)
	if err != nil {
		// Chưa có cấu hình → log ở mức mặc định, vẫn giữ định dạng JSON.
		newLogger(slog.LevelInfo).Error("load config", "err", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.LogLevel)
	logger.Debug("config loaded", "http_addr", cfg.HTTPAddr, "log_level", cfg.LogLevel.String())

	// pgxpool.New chỉ parse URL và tạo pool; kết nối mở lười khi cần.
	// → core vẫn khởi động khi PG chưa lên, /readyz báo 503 tới khi PG sẵn sàng.
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("create db pool", "err", err) // không log URL: chứa mật khẩu
		os.Exit(1)
	}
	defer pool.Close()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpx.NewRouter(logger, pool),
		ReadHeaderTimeout: 5 * time.Second, // client gửi header quá chậm → cắt (chống slowloris)
	}

	logger.Info("core server listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("http server", "err", err)
		os.Exit(1)
	}
}

// newLogger tạo logger JSON một dòng ra stdout; dòng dưới level bị bỏ.
func newLogger(level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
