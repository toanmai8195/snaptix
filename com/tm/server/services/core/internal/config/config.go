// Package config đọc cấu hình core từ biến môi trường.
package config

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/toanmai8195/snaptix/com/tm/server/pkg/otelx"
)

// Config là cấu hình của core service.
type Config struct {
	HTTPAddr    string
	DatabaseURL string
	LogLevel    slog.Level
}

const (
	defaultHTTPAddr    = ":8080"
	defaultDatabaseURL = "postgres://snaptix:snaptix@localhost:5432/core?sslmode=disable"
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
	return cfg, errors.Join(errs...)
}

func or(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
