// Package config đọc cấu hình core từ biến môi trường.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/toanmai8195/snaptix/com/tm/server/pkg/otelx"
)

// Config là cấu hình của core service.
type Config struct {
	HTTPAddr    string
	DatabaseURL string
	LogLevel    slog.Level
	// ShutdownTimeout: thời gian tối đa chờ request đang chạy khi dừng service.
	ShutdownTimeout time.Duration
	OTel            otelx.Config
}

const (
	defaultHTTPAddr    = ":8080"
	defaultDatabaseURL = "postgres://snaptix:snaptix@localhost:5432/core?sslmode=disable"
	defaultShutdown    = 15 * time.Second
)

// Load đọc cấu hình qua getenv (truyền os.Getenv; test truyền map) và áp giá trị mặc định.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		HTTPAddr:    or(getenv("CORE_HTTP_ADDR"), defaultHTTPAddr),
		DatabaseURL: or(getenv("CORE_DATABASE_URL"), defaultDatabaseURL),
	}
	var errs []error
	level, err := otelx.ParseLevel(getenv("LOG_LEVEL"))
	if err != nil {
		errs = append(errs, fmt.Errorf("LOG_LEVEL: %w", err))
	}
	cfg.LogLevel = level

	cfg.ShutdownTimeout = defaultShutdown
	if v := getenv("CORE_SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			errs = append(errs, fmt.Errorf("CORE_SHUTDOWN_TIMEOUT: %q không phải thời lượng dương (vd 15s)", v))
		} else {
			cfg.ShutdownTimeout = d
		}
	}
	cfg.OTel, err = otelx.ConfigFromEnv("core", or(getenv("CORE_VERSION"), "dev"), getenv)
	if err != nil {
		errs = append(errs, err)
	}
	return cfg, errors.Join(errs...)
}

func or(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
