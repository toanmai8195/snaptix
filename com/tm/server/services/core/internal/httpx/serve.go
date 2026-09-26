package httpx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Serve phục vụ srv trên ln cho tới khi ctx bị huỷ (vd nhận SIGTERM), rồi dừng êm:
// ngừng nhận kết nối mới, chờ request đang chạy xong trong tối đa timeout.
// Quá timeout thì đóng cưỡng bức và trả lỗi bọc context.DeadlineExceeded.
func Serve(ctx context.Context, srv *http.Server, ln net.Listener, log *slog.Logger, timeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
	}

	// ctx đã bị huỷ: dùng context riêng (giữ value, bỏ cancel) để đếm thời gian dừng.
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()
	log.InfoContext(shutdownCtx, "shutting down", slog.Duration("timeout", timeout))

	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close()
		log.WarnContext(shutdownCtx, "shutdown timeout, đóng cưỡng bức kết nối còn lại", slog.Any("error", err))
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}
	log.InfoContext(shutdownCtx, "http server stopped")
	return nil
}
