package main

import (
	"fmt"
	"log/slog"
)

// config là cấu hình của core server, đọc từ biến môi trường (xem local-setup.md).
type config struct {
	HTTPAddr    string     // CORE_HTTP_ADDR
	DatabaseURL string     // CORE_DATABASE_URL
	LogLevel    slog.Level // LOG_LEVEL: debug | info | warn | error
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
	return cfg, nil
}

func envOr(getenv func(string) string, key, fallback string) string {
	if v := getenv(key); v != "" {
		return v
	}
	return fallback
}
