package otelx

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// fakeCollector ghi lại request OTLP HTTP nhận được.
type fakeCollector struct {
	mu     sync.Mutex
	bodies map[string][][]byte
	srv    *httptest.Server
}

func newFakeCollector(t *testing.T) *fakeCollector {
	t.Helper()
	fc := &fakeCollector{bodies: map[string][][]byte{}}
	fc.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		fc.mu.Lock()
		fc.bodies[r.URL.Path] = append(fc.bodies[r.URL.Path], b)
		fc.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(fc.srv.Close)
	return fc
}

func (fc *fakeCollector) get(path string) [][]byte {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.bodies[path]
}

func discardLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func emit(ctx context.Context) {
	_, span := otel.Tracer("test").Start(ctx, "work")
	span.End()
	c, _ := otel.Meter("test").Int64Counter("test.ops")
	c.Add(ctx, 1)
}

// TC01
func TestSetup_ExportsTracesAndMetrics(t *testing.T) {
	fc := newFakeCollector(t)
	shutdown, err := Setup(context.Background(), Config{
		ServiceName: "core", ServiceVersion: "1.2.3", Endpoint: fc.srv.URL, MetricInterval: time.Hour,
	}, discardLogger())
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	emit(context.Background())
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	for _, path := range []string{"/v1/traces", "/v1/metrics"} {
		bodies := fc.get(path)
		if len(bodies) == 0 {
			t.Fatalf("collector không nhận %s", path)
		}
		all := bytes.Join(bodies, nil)
		for _, want := range []string{"service.name", "core", "service.version", "1.2.3"} {
			if !bytes.Contains(all, []byte(want)) {
				t.Errorf("%s thiếu %q trong resource", path, want)
			}
		}
	}
	if !bytes.Contains(bytes.Join(fc.get("/v1/metrics"), nil), []byte("test.ops")) {
		t.Error("metric test.ops không được export khi shutdown")
	}
}

// TC02
func TestSetup_Disabled(t *testing.T) {
	fc := newFakeCollector(t)
	shutdown, err := Setup(context.Background(), Config{ServiceName: "core", Endpoint: fc.srv.URL, Disabled: true}, discardLogger())
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	emit(context.Background())
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if n := len(fc.get("/v1/traces")) + len(fc.get("/v1/metrics")); n != 0 {
		t.Fatalf("Disabled mà vẫn gửi %d request", n)
	}
}

// TC03
func TestSetup_Propagator(t *testing.T) {
	shutdown, err := Setup(context.Background(), Config{ServiceName: "core", Disabled: true}, discardLogger())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	tid, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	sid, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	ctx := trace.ContextWithSpanContext(context.Background(),
		trace.NewSpanContext(trace.SpanContextConfig{TraceID: tid, SpanID: sid, TraceFlags: trace.FlagsSampled}))
	member, _ := baggage.NewMember("tenant", "snaptix")
	bag, _ := baggage.New(member)
	ctx = baggage.ContextWithBaggage(ctx, bag)

	carrier := propagation.HeaderCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	if got := carrier.Get("traceparent"); got != "00-"+tid.String()+"-"+sid.String()+"-01" {
		t.Errorf("traceparent = %q", got)
	}
	if got := carrier.Get("baggage"); got != "tenant=snaptix" {
		t.Errorf("baggage = %q", got)
	}
	out := otel.GetTextMapPropagator().Extract(context.Background(), carrier)
	if trace.SpanContextFromContext(out).TraceID() != tid {
		t.Error("extract traceparent không ra trace ID")
	}
}

// TC04
func TestSetup_CollectorUnreachable(t *testing.T) {
	start := time.Now()
	shutdown, err := Setup(context.Background(), Config{ServiceName: "core", Endpoint: "http://127.0.0.1:1", MetricInterval: time.Hour}, discardLogger())
	if err != nil {
		t.Fatalf("Setup phải thành công khi collector chưa chạy: %v", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("Setup bị chặn %v", d)
	}
	emit(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	start = time.Now()
	_ = shutdown(ctx) // có thể trả lỗi export — chỉ cần không treo, không panic
	if d := time.Since(start); d > 4*time.Second {
		t.Errorf("shutdown mất %v, vượt thời hạn ctx", d)
	}
}

func TestConfigFromEnv(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

	cfg, err := ConfigFromEnv("core", "dev", env(nil))
	if err != nil || cfg.Endpoint != "http://localhost:4318" || cfg.Disabled || cfg.MetricInterval != 10*time.Second {
		t.Fatalf("mặc định sai: %+v %v", cfg, err)
	}
	cfg, err = ConfigFromEnv("core", "dev", env(map[string]string{
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://collector:4318", "OTEL_SDK_DISABLED": "true", "OTEL_METRIC_EXPORT_INTERVAL": "1500",
	}))
	if err != nil || cfg.Endpoint != "http://collector:4318" || !cfg.Disabled || cfg.MetricInterval != 1500*time.Millisecond {
		t.Fatalf("override sai: %+v %v", cfg, err)
	}
	if _, err := ConfigFromEnv("core", "dev", env(map[string]string{"OTEL_METRIC_EXPORT_INTERVAL": "nhanh"})); err == nil {
		t.Error("OTEL_METRIC_EXPORT_INTERVAL sai phải lỗi")
	}
}
