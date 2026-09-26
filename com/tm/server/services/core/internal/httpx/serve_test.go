package httpx

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

// startServe chạy Serve trên cổng ngẫu nhiên với handler cho trước.
func startServe(t *testing.T, h http.HandlerFunc, timeout time.Duration) (addr string, cancel context.CancelFunc, done <-chan error, logs *bytes.Buffer) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	logs = &bytes.Buffer{}
	log := slog.New(slog.NewJSONHandler(logs, nil))
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error, 1)
	go func() { ch <- Serve(ctx, &http.Server{Handler: h}, ln, log, timeout) }()
	t.Cleanup(cancel)
	return ln.Addr().String(), cancel, ch, logs
}

func waitDone(t *testing.T, done <-chan error, max time.Duration) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(max):
		t.Fatalf("Serve không trả về trong %v", max)
		return nil
	}
}

// sleepHandler ngủ d rồi trả "done"; dừng sớm nếu request bị huỷ.
func sleepHandler(d time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(d):
			_, _ = io.WriteString(w, "done")
		case <-r.Context().Done():
		}
	}
}

// TC01
func TestServe_StopsWhenIdle(t *testing.T) {
	_, cancel, done, logs := startServe(t, sleepHandler(0), 5*time.Second)
	time.Sleep(50 * time.Millisecond)

	start := time.Now()
	cancel()
	if err := waitDone(t, done, time.Second); err != nil {
		t.Fatalf("Serve = %v, muốn nil", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("dừng mất %v", d)
	}
	out := logs.String()
	if i, j := strings.Index(out, "shutting down"), strings.Index(out, "http server stopped"); i < 0 || j < i {
		t.Errorf("log phải có 'shutting down' rồi 'http server stopped': %s", out)
	}
}

// TC02 + TC03
func TestServe_WaitsInFlightAndRefusesNew(t *testing.T) {
	addr, cancel, done, _ := startServe(t, sleepHandler(500*time.Millisecond), 5*time.Second)

	type result struct {
		code int
		body string
		err  error
	}
	resCh := make(chan result, 1)
	go func() {
		resp, err := http.Get("http://" + addr + "/slow")
		if err != nil {
			resCh <- result{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		b, _ := io.ReadAll(resp.Body)
		resCh <- result{code: resp.StatusCode, body: string(b)}
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	// TC03: listener đóng ngay khi bắt đầu dừng → kết nối mới bị từ chối.
	time.Sleep(50 * time.Millisecond)
	if conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond); err == nil {
		_ = conn.Close()
		t.Error("kết nối mới sau khi bắt đầu dừng phải bị từ chối")
	}

	select {
	case err := <-done:
		t.Fatalf("Serve trả về (%v) khi request còn đang chạy", err)
	default:
	}

	r := <-resCh
	if r.err != nil || r.code != http.StatusOK || r.body != "done" {
		t.Fatalf("request đang chạy phải hoàn thành 200 'done', có %d %q %v", r.code, r.body, r.err)
	}
	if err := waitDone(t, done, 2*time.Second); err != nil {
		t.Fatalf("Serve = %v, muốn nil", err)
	}
}

// TC04
func TestServe_TimeoutForcesClose(t *testing.T) {
	addr, cancel, done, logs := startServe(t, sleepHandler(5*time.Second), 200*time.Millisecond)

	reqErr := make(chan error, 1)
	go func() {
		resp, err := http.Get("http://" + addr + "/very-slow")
		if err == nil {
			_, err = io.ReadAll(resp.Body)
			_ = resp.Body.Close()
		}
		reqErr <- err
	}()
	time.Sleep(100 * time.Millisecond)

	start := time.Now()
	cancel()
	err := waitDone(t, done, 2*time.Second)
	if d := time.Since(start); d > time.Second {
		t.Errorf("phải trả về sau ≈ timeout 200ms, mất %v", d)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Serve = %v, muốn bọc context.DeadlineExceeded", err)
	}
	if !strings.Contains(logs.String(), `"level":"WARN"`) {
		t.Errorf("phải log WARN khi quá timeout: %s", logs)
	}
	select {
	case e := <-reqErr:
		if e == nil {
			t.Error("request quá timeout phải bị đóng kết nối (lỗi phía client)")
		}
	case <-time.After(2 * time.Second):
		t.Error("kết nối không bị đóng cưỡng bức")
	}
}
