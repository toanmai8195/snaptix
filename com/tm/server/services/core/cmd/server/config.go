package main

import (
	"fmt"
	"log/slog"
	"time"
)

// config là cấu hình của core server, đọc từ biến môi trường (xem local-setup.md).
type config struct {
	HTTPAddr    string     // CORE_HTTP_ADDR
	DatabaseURL string     // CORE_DATABASE_URL
	LogLevel    slog.Level // LOG_LEVEL: debug | info | warn | error
	// CORE_SHUTDOWN_TIMEOUT: thời gian chờ request đang chạy khi nhận SIGTERM.
	// Phải nhỏ hơn grace period của K8s/docker (mặc định 30s / 10s), nếu không sẽ bị SIGKILL giữa chừng.
	ShutdownTimeout time.Duration
}

// loadConfig đọc cấu hình qua getenv (truyền os.Getenv khi chạy thật, map giả khi test).
// Biến không đặt hoặc rỗng → dùng giá trị mặc định cho môi trường local.
func loadConfig(getenv func(string) string) (config, error) {
	cfg := config{
		HTTPAddr:    envOr(getenv, "CORE_HTTP_ADDR", ":8080"),
		DatabaseURL: envOr(getenv, "CORE_DATABASE_URL", "postgres://snaptix:snaptix@localhost:5432/core?sslmode=disable"),
	}

	// UnmarshalText nhận "debug", "info", "warn", "error" (không phân biệt hoa thường).
	level := envOr(getenv, "LOG_LEVEL", "info")
	if err := cfg.LogLevel.UnmarshalText([]byte(level)); err != nil {
		return config{}, fmt.Errorf("LOG_LEVEL %q: %w", level, err)
	}

	// ParseDuration nhận "15s", "500ms", "1m30s"...
	timeout := envOr(getenv, "CORE_SHUTDOWN_TIMEOUT", "15s")
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return config{}, fmt.Errorf("CORE_SHUTDOWN_TIMEOUT %q: %w", timeout, err)
	}
	if d <= 0 {
		return config{}, fmt.Errorf("CORE_SHUTDOWN_TIMEOUT %q: phải > 0", timeout)
	}
	cfg.ShutdownTimeout = d
	return cfg, nil
}

func envOr(getenv func(string) string, key, fallback string) string {
	if v := getenv(key); v != "" {
		return v
	}
	return fallback
}
