// Command server là HTTP server của core service.
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

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/toanmai8195/snaptix/com/tm/server/services/core/internal/httpx"
)

// main là chỗ DUY NHẤT gọi os.Exit. os.Exit thoát ngay, bỏ qua mọi defer,
// nên toàn bộ việc cần dọn dẹp nằm trong run — run return thì defer đã chạy xong.
func main() {
	cfg, err := loadConfig(os.Getenv)
	if err != nil {
		// Chưa có cấu hình → log ở mức mặc định, vẫn giữ định dạng JSON.
		newLogger(slog.LevelInfo).Error("load config", "err", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.LogLevel)
	if err := run(cfg, logger); err != nil {
		logger.Error("core server exited", "err", err)
		os.Exit(1)
	}
}

func run(cfg config, logger *slog.Logger) error {
	logger.Debug("config loaded", "http_addr", cfg.HTTPAddr, "log_level", cfg.LogLevel.String(),
		"shutdown_timeout", cfg.ShutdownTimeout.String())

	// pgxpool.New chỉ parse URL và tạo pool; kết nối mở lười khi cần.
	// → core vẫn khởi động khi PG chưa lên, /readyz báo 503 tới khi PG sẵn sàng.
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create db pool: %w", err) // không đưa URL vào lỗi: chứa mật khẩu
	}
	defer func() {
		pool.Close()
		logger.Info("db pool closed")
	}()

	srv := &http.Server{
		Handler:           httpx.NewRouter(logger, pool),
		ReadHeaderTimeout: 5 * time.Second, // client gửi header quá chậm → cắt (chống slowloris)
	}

	// Mở cổng trước, TRONG goroutine chính: cổng bị chiếm → lỗi trả về ngay,
	// và chỉ log "listening" khi đã nghe thật.
	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	logger.Info("core server listening", "addr", ln.Addr().String())

	ctx, stop := shutdownContext(logger)
	defer stop()

	// serve chỉ return khi HTTP đã dừng hẳn → sau đó defer mới đóng pool.
	return serve(ctx, srv, ln, cfg.ShutdownTimeout, logger)
}

// shutdownContext trả về context bị huỷ khi nhận SIGINT (Ctrl-C) hoặc SIGTERM (docker stop / K8s).
// Không dùng signal.NotifyContext vì ở Go 1.24 nó không cho biết signal nào tới.
func shutdownContext(logger *slog.Logger) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	// Buffer 1: signal tới lúc chưa ai đọc vẫn không bị mất.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigCh:
			logger.Info("shutting down", "signal", sig.String())
		case <-ctx.Done(): // run kết thúc vì lý do khác (vd server lỗi)
		}
		// Bỏ đăng ký: signal lần 2 quay về mặc định của Go → giết process ngay.
		signal.Stop(sigCh)
		cancel()
	}()
	return ctx, cancel
}

// serve phục vụ HTTP trên ln cho tới khi ctx bị huỷ, rồi dừng êm:
// ngừng nhận kết nối mới, chờ request đang chạy tối đa timeout, quá hạn thì cắt.
// Nhận ctx + listener thay vì tự bắt signal / tự mở cổng → test gọi được với cổng :0 và cancel().
func serve(ctx context.Context, srv *http.Server, ln net.Listener, timeout time.Duration, logger *slog.Logger) error {
	// Serve block → chạy trong goroutine riêng, kết quả gửi về qua channel.
	// Buffer 1: goroutine gửi xong là thoát, kể cả khi không còn ai nhận.
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	// Chờ cái nào tới trước: server lỗi hoặc yêu cầu dừng.
	select {
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
	}

	// Dừng êm: Shutdown đóng listener (request mới bị từ chối), đóng kết nối idle,
	// rồi CHỜ request đang chạy xong — tối đa timeout.
	// Context mới từ Background: ctx đã bị huỷ, dùng lại thì Shutdown trả về ngay.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		// Hết giờ mà vẫn còn request → cắt thô những request còn lại.
		logger.Warn("shutdown timeout exceeded, closing remaining connections",
			"timeout", timeout.String(), "err", err)
		_ = srv.Close()
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	logger.Info("http server stopped")
	return nil
}

// newLogger tạo logger JSON một dòng ra stdout; dòng dưới level bị bỏ.
func newLogger(level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
