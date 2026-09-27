package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

// testServer chạy serve trên cổng ngẫu nhiên (:0); cancel() đóng vai signal dừng.
type testServer struct {
	url    string
	addr   string
	cancel context.CancelFunc
	done   chan error // nhận kết quả của serve
}

func startServe(t *testing.T, h http.Handler, timeout time.Duration) *testServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0") // OS chọn cổng trống
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	ts := &testServer{
		url:    "http://" + ln.Addr().String(),
		addr:   ln.Addr().String(),
		cancel: cancel,
		done:   make(chan error, 1),
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	srv := &http.Server{Handler: h, ReadHeaderTimeout: time.Second}
	go func() { ts.done <- serve(ctx, srv, ln, timeout, logger) }()
	return ts
}

// wait chờ serve return, fail nếu quá max.
func (ts *testServer) wait(t *testing.T, max time.Duration) error {
	t.Helper()
	select {
	case err := <-ts.done:
		return err
	case <-time.After(max):
		t.Fatalf("serve chưa return sau %v", max)
		return nil
	}
}

type result struct {
	status int
	body   string
	err    error
}

// get gọi url trong goroutine, trả kết quả qua channel.
func get(url string) <-chan result {
	ch := make(chan result, 1)
	go func() {
		resp, err := http.Get(url)
		if err != nil {
			ch <- result{err: err}
			return
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		ch <- result{status: resp.StatusCode, body: string(b), err: err}
	}()
	return ch
}

// slowHandler báo qua started khi request bắt đầu, rồi ngủ d mới trả "done".
func slowHandler(d time.Duration, started chan<- struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-time.After(d):
			_, _ = io.WriteString(w, "done")
		case <-r.Context().Done(): // bị Close cắt ngang
		}
	})
}

func TestServeStopsWhenIdle(t *testing.T) {
	ts := startServe(t, http.NotFoundHandler(), 5*time.Second)

	ts.cancel()

	if err := ts.wait(t, time.Second); err != nil {
		t.Fatalf("serve() = %v, muốn nil", err)
	}
}

// TC03: request đang chạy được chờ xong rồi serve mới return nil.
func TestServeWaitsForInFlightRequest(t *testing.T) {
	const handlerTime = 1 * time.Second
	started := make(chan struct{})
	ts := startServe(t, slowHandler(handlerTime, started), 5*time.Second)

	resCh := get(ts.url)
	<-started
	start := time.Now()
	ts.cancel()

	res := <-resCh
	if res.err != nil || res.status != http.StatusOK || res.body != "done" {
		t.Fatalf("request đang chạy = %+v, muốn 200 \"done\"", res)
	}
	if err := ts.wait(t, 2*time.Second); err != nil {
		t.Fatalf("serve() = %v, muốn nil", err)
	}
	if elapsed := time.Since(start); elapsed < handlerTime/2 {
		t.Errorf("serve return sau %v — không chờ request (≈ %v)", elapsed, handlerTime)
	}
}

// TC04: đã bắt đầu shutdown thì kết nối mới bị từ chối, request cũ vẫn xong.
func TestServeRejectsNewConnectionsDuringShutdown(t *testing.T) {
	started := make(chan struct{})
	ts := startServe(t, slowHandler(1*time.Second, started), 5*time.Second)

	resCh := get(ts.url)
	<-started
	ts.cancel()

	// Shutdown chạy trong goroutine của serve → thử kết nối lại tới khi bị từ chối.
	deadline := time.Now().Add(time.Second)
	for {
		conn, err := net.DialTimeout("tcp", ts.addr, 100*time.Millisecond)
		if err != nil {
			break // listener đã đóng
		}
		conn.Close()
		if time.Now().After(deadline) {
			t.Fatal("vẫn kết nối mới được sau khi bắt đầu shutdown")
		}
		time.Sleep(10 * time.Millisecond)
	}

	if res := <-resCh; res.err != nil || res.status != http.StatusOK {
		t.Fatalf("request đang chạy = %+v, muốn 200", res)
	}
	if err := ts.wait(t, 2*time.Second); err != nil {
		t.Fatalf("serve() = %v, muốn nil", err)
	}
}

// TC05: request lâu hơn timeout → serve return lỗi sau ≈ timeout, request bị cắt.
func TestServeShutdownTimeout(t *testing.T) {
	const timeout = 200 * time.Millisecond
	started := make(chan struct{})
	ts := startServe(t, slowHandler(10*time.Second, started), timeout)

	resCh := get(ts.url)
	<-started
	start := time.Now()
	ts.cancel()

	err := ts.wait(t, 2*time.Second)
	elapsed := time.Since(start)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("serve() = %v, muốn lỗi bọc context.DeadlineExceeded", err)
	}
	if elapsed < timeout || elapsed > timeout+time.Second {
		t.Errorf("serve return sau %v, muốn ≈ %v", elapsed, timeout)
	}
	if res := <-resCh; res.err == nil {
		t.Errorf("request vượt timeout = %+v, muốn bị cắt (lỗi phía client)", res)
	}
}

// Listener hỏng → serve trả lỗi ngay, không cần ctx bị huỷ.
func TestServeListenerError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ln.Close()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	err = serve(context.Background(), &http.Server{ReadHeaderTimeout: time.Second}, ln, time.Second, logger)
	if err == nil {
		t.Fatal("serve() = nil, muốn lỗi khi listener đã đóng")
	}
}
