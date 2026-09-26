// Command server chạy HTTP API của core service.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/toanmai8195/snaptix/com/tm/server/pkg/otelx"
	"github.com/toanmai8195/snaptix/com/tm/server/pkg/postgres"
	"github.com/toanmai8195/snaptix/com/tm/server/services/core/internal/config"
	"github.com/toanmai8195/snaptix/com/tm/server/services/core/internal/httpx"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "core:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// SIGTERM (orchestrator) / SIGINT (Ctrl+C) huỷ ctx → Serve bắt đầu dừng êm.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, os.Interrupt)
	defer stop()

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	log := otelx.NewLogger(os.Stdout, cfg.LogLevel, "core")
	// Đọc/ghi traceparent (W3C) để trace_id xuyên service; SDK exporter thêm ở P0-T10.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	// Đóng pool SAU khi server dừng hẳn: request đang chạy vẫn dùng được DB.
	defer func() {
		pool.Close()
		log.Info("database pool closed")
	}()

	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", cfg.HTTPAddr, err)
	}
	srv := &http.Server{
		Handler:           httpx.NewRouter(httpx.Deps{Log: log, DB: pool}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.InfoContext(ctx, "core started", slog.String("addr", cfg.HTTPAddr))
	return httpx.Serve(ctx, srv, ln, log, cfg.ShutdownTimeout)
}
