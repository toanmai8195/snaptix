// Command worker chạy job nền của core. Hiện chỉ là khung: khởi tạo log, OpenTelemetry,
// pool DB rồi chờ tín hiệu dừng. Job thật (hết hạn hold, relay outbox...) thêm ở các phase sau.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/toanmai8195/snaptix/com/tm/server/pkg/otelx"
	"github.com/toanmai8195/snaptix/com/tm/server/pkg/postgres"
	"github.com/toanmai8195/snaptix/com/tm/server/services/core/internal/config"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "worker:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, os.Interrupt)
	defer stop()

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	cfg.OTel.ServiceName = "core-worker"
	log := otelx.NewLogger(os.Stdout, cfg.LogLevel, "core-worker")

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	otelShutdown, err := otelx.Setup(ctx, cfg.OTel, log)
	if err != nil {
		return err
	}
	defer func() {
		sctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = otelShutdown(sctx)
	}()

	log.InfoContext(ctx, "worker started")
	<-ctx.Done()
	log.Info("worker stopped")
	return nil
}
