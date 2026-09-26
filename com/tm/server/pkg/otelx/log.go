// Package otelx gom phần khởi tạo observability dùng chung giữa các service:
// log (slog JSON) và — ở các bước sau — tracing, metrics.
package otelx

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// ParseLevel đổi chuỗi cấu hình (debug|info|warn|error) thành slog.Level.
func ParseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	}
	return 0, fmt.Errorf("log level không hợp lệ %q (debug|info|warn|error)", s)
}

// NewLogger tạo logger JSON một dòng mỗi bản ghi, gắn sẵn tên service.
func NewLogger(w io.Writer, level slog.Level, service string) *slog.Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	return slog.New(h).With(slog.String("service", service))
}
