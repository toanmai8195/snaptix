// Package postgres gom phần kết nối PostgreSQL dùng chung giữa các service.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool tạo pool kết nối. Pool mở kết nối lười: service vẫn khởi động được khi DB
// chưa sẵn sàng, còn /readyz báo trạng thái thật qua Ping.
// Truy vấn trong một trace đang có được ghi thành span con (OpenTelemetry).
func NewPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	if cfg.ConnConfig.ConnectTimeout == 0 {
		cfg.ConnConfig.ConnectTimeout = 2 * time.Second
	}
	cfg.ConnConfig.Tracer = newQueryTracer()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	return pool, nil
}
