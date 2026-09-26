// Command server chạy HTTP API của core service.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
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
	defer pool.Close()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpx.NewRouter(httpx.Deps{Log: log, DB: pool}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.InfoContext(ctx, "core started", slog.String("addr", cfg.HTTPAddr))
	return srv.ListenAndServe()
}
