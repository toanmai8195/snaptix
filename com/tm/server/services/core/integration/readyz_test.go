package integration

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/toanmai8195/snaptix/com/tm/server/services/core/internal/httpx"
	"github.com/toanmai8195/snaptix/com/tm/server/services/core/internal/pgtest"
)

func discardLogger() *slog.Logger { return slog.New(slog.NewJSONHandler(io.Discard, nil)) }

// statusOf gọi GET url, trả status code (0 nếu không gọi được).
func statusOf(t *testing.T, url string) int {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Logf("GET %s: %v", url, err)
		return 0
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode
}

// eventually gọi url tới khi trả want hoặc hết timeout — hạ tầng thật cần thời gian đổi trạng thái.
func eventually(t *testing.T, url string, want int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		got := statusOf(t, url)
		if got == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("GET %s = %d, muốn %d sau %v", url, got, want, timeout)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// P0-AT02 / TC05: PG đang chạy → /readyz 200.
func TestReadyzWithPostgres(t *testing.T) {
	pool := pgtest.NewDB(t)
	srv := httptest.NewServer(httpx.NewRouter(discardLogger(), pool))
	t.Cleanup(srv.Close)

	if got := statusOf(t, srv.URL+"/readyz"); got != http.StatusOK {
		t.Fatalf("/readyz = %d, muốn 200", got)
	}
}

// P0-AT03 / TC06: dừng PG → /readyz 503, /healthz vẫn 200; bật lại PG → /readyz về 200.
// Test phá hạ tầng → container RIÊNG, không đụng container dùng chung của các test khác.
func TestReadyzWhenPostgresStops(t *testing.T) {
	pgtest.RequireDocker(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	fixedPort, err := pgtest.FixedHostPort()
	if err != nil {
		t.Fatal(err)
	}
	pg, err := pgtest.Start(ctx, fixedPort)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pg.Terminate(context.Background()) })
	pool := pg.NewDB(t)

	srv := httptest.NewServer(httpx.NewRouter(discardLogger(), pool))
	t.Cleanup(srv.Close)
	eventually(t, srv.URL+"/readyz", http.StatusOK, 5*time.Second)

	stopTimeout := 5 * time.Second
	if err := pg.Testcontainer().Stop(ctx, &stopTimeout); err != nil {
		t.Fatalf("stop postgres: %v", err)
	}
	eventually(t, srv.URL+"/readyz", http.StatusServiceUnavailable, 5*time.Second)
	if got := statusOf(t, srv.URL+"/healthz"); got != http.StatusOK {
		t.Errorf("/healthz khi PG dừng = %d, muốn 200", got)
	}

	if err := pg.Testcontainer().Start(ctx); err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	// Cùng cổng host (FixedHostPort) → pool cũ tự kết nối lại, không cần restart "core".
	eventually(t, srv.URL+"/readyz", http.StatusOK, 30*time.Second)
}
