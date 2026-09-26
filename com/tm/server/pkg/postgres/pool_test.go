package postgres

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewPool_InvalidURL(t *testing.T) {
	_, err := NewPool(context.Background(), "postgres://%zz")
	if err == nil || !strings.Contains(err.Error(), "parse database url") {
		t.Fatalf("muốn lỗi parse database url, có %v", err)
	}
}

func TestNewPool_LazyAndDefaultTimeout(t *testing.T) {
	// Cổng 1 không có DB: pool vẫn tạo được vì kết nối mở lười.
	pool, err := NewPool(context.Background(), "postgres://u:p@127.0.0.1:1/db?sslmode=disable")
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close()

	if got := pool.Config().ConnConfig.ConnectTimeout; got != 2*time.Second {
		t.Errorf("ConnectTimeout = %v, muốn 2s", got)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err == nil {
		t.Error("Ping tới cổng không có DB phải lỗi")
	}
}

func TestNewPool_KeepsExplicitTimeout(t *testing.T) {
	pool, err := NewPool(context.Background(), "postgres://u:p@127.0.0.1:1/db?connect_timeout=7")
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close()
	if got := pool.Config().ConnConfig.ConnectTimeout; got != 7*time.Second {
		t.Errorf("ConnectTimeout = %v, muốn 7s từ URL", got)
	}
}
