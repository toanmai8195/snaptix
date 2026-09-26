package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/toanmai8195/snaptix/com/tm/server/pkg/otelx"
)

type harness struct {
	h       http.Handler
	logs    *bytes.Buffer
	spans   *tracetest.SpanRecorder
	metrics *sdkmetric.ManualReader
}

func newHarness(t *testing.T, level slog.Level) harness {
	t.Helper()
	logs := &bytes.Buffer{}
	spans := tracetest.NewSpanRecorder()
	reader := sdkmetric.NewManualReader()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()); _ = mp.Shutdown(context.Background()) })

	h := NewRouter(Deps{
		Log: otelx.NewLogger(logs, level, "core"),
		DB:  pingOK,
		Routes: []func(chi.Router){func(r chi.Router) {
			r.Get("/test/{id}", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("hello")) })
			r.Get("/boom", func(http.ResponseWriter, *http.Request) { panic("nổ") })
			r.Get("/fail", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadGateway) })
		}},
		OTel: []otelhttp.Option{
			otelhttp.WithTracerProvider(tp),
			otelhttp.WithMeterProvider(mp),
			otelhttp.WithPropagators(propagation.TraceContext{}),
		},
	})
	return harness{h: h, logs: logs, spans: spans, metrics: reader}
}

func (hs harness) get(path string, header http.Header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for k, vs := range header {
		for _, v := range vs {
			req.Header.Add(k, v) // Add chuẩn hoá khoá (X-Request-ID → X-Request-Id)
		}
	}
	rec := httptest.NewRecorder()
	hs.h.ServeHTTP(rec, req)
	return rec
}

// logLines trả các dòng log JSON có msg cho trước.
func (hs harness) logLines(t *testing.T, msg string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(hs.logs.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(line, &m); err != nil {
			t.Fatalf("log không phải JSON: %s", line)
		}
		if m["msg"] == msg {
			out = append(out, m)
		}
	}
	return out
}

var hex32 = regexp.MustCompile(`^[0-9a-f]{32}$`)

// TC01
func TestRequestID(t *testing.T) {
	hs := newHarness(t, slog.LevelInfo)

	a := hs.get("/test/1", nil).Header().Get(headerRequestID)
	b := hs.get("/test/1", nil).Header().Get(headerRequestID)
	if !hex32.MatchString(a) || !hex32.MatchString(b) || a == b {
		t.Fatalf("request ID sinh mới phải là 32 hex và khác nhau: %q %q", a, b)
	}
	if got := hs.get("/test/1", http.Header{headerRequestID: {"abc-123.X_y"}}).Header().Get(headerRequestID); got != "abc-123.X_y" {
		t.Fatalf("header hợp lệ phải giữ nguyên, có %q", got)
	}
	for _, bad := range []string{strings.Repeat("a", 65), "có dấu", "a b", "<script>"} {
		if got := hs.get("/test/1", http.Header{headerRequestID: {bad}}).Header().Get(headerRequestID); got == bad || !hex32.MatchString(got) {
			t.Errorf("header không hợp lệ %q phải thay bằng ID mới, có %q", bad, got)
		}
	}
}

// TC02
func TestRecoverer(t *testing.T) {
	hs := newHarness(t, slog.LevelInfo)

	rec := hs.get("/boom", nil)
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), `"error":"internal"`) {
		t.Fatalf("panic → %d %s; muốn 500 {\"error\":\"internal\"}", rec.Code, rec.Body)
	}
	if rec := hs.get("/test/1", nil); rec.Code != http.StatusOK {
		t.Fatalf("sau panic server phải vẫn phục vụ, có %d", rec.Code)
	}
	panics := hs.logLines(t, "panic")
	if len(panics) != 1 {
		t.Fatalf("muốn 1 log panic, có %d", len(panics))
	}
	p := panics[0]
	if p["level"] != "ERROR" || p["panic"] != "nổ" || !strings.Contains(p["stack"].(string), "goroutine") || !hex32.MatchString(p["request_id"].(string)) {
		t.Fatalf("log panic thiếu thông tin: %v", p)
	}
}

// TC03
func TestAccessLog(t *testing.T) {
	hs := newHarness(t, slog.LevelInfo)
	hs.get("/test/42", http.Header{headerRequestID: {"req-1"}})
	hs.get("/healthz", nil)
	hs.get("/fail", nil)

	lines := hs.logLines(t, "http request")
	if len(lines) != 2 {
		t.Fatalf("muốn 2 dòng access log (healthz ở DEBUG bị ẩn), có %d: %s", len(lines), hs.logs)
	}
	l := lines[0]
	want := map[string]any{"level": "INFO", "method": "GET", "path": "/test/42", "route": "/test/{id}",
		"status": float64(200), "bytes": float64(5), "request_id": "req-1"}
	for k, v := range want {
		if l[k] != v {
			t.Errorf("%s = %v, muốn %v", k, l[k], v)
		}
	}
	if _, ok := l["duration_ms"].(float64); !ok {
		t.Error("thiếu duration_ms")
	}
	if lines[1]["level"] != "ERROR" || lines[1]["status"] != float64(502) {
		t.Errorf("5xx phải log ERROR: %v", lines[1])
	}

	dbg := newHarness(t, slog.LevelDebug)
	dbg.get("/healthz", nil)
	if l := dbg.logLines(t, "http request"); len(l) != 1 || l[0]["level"] != "DEBUG" {
		t.Fatalf("probe phải log ở DEBUG: %v", l)
	}
}

// TC04
func TestTracing(t *testing.T) {
	hs := newHarness(t, slog.LevelInfo)
	const traceID = "4bf92f3577b34da6a3ce929d0e0e4736"
	const parentID = "00f067aa0ba902b7"
	hs.get("/test/7", http.Header{"Traceparent": {"00-" + traceID + "-" + parentID + "-01"}})
	hs.get("/healthz", nil)

	ended := hs.spans.Ended()
	if len(ended) != 1 {
		t.Fatalf("muốn 1 span (probe bị lọc), có %d", len(ended))
	}
	s := ended[0]
	if s.Name() != "GET /test/{id}" {
		t.Errorf("tên span = %q", s.Name())
	}
	if s.SpanContext().TraceID().String() != traceID || s.Parent().SpanID().String() != parentID {
		t.Errorf("span phải nối vào traceparent: trace=%s parent=%s", s.SpanContext().TraceID(), s.Parent().SpanID())
	}
	attrs := map[attribute.Key]attribute.Value{}
	for _, kv := range s.Attributes() {
		attrs[kv.Key] = kv.Value
	}
	if attrs["http.route"].AsString() != "/test/{id}" || attrs["http.response.status_code"].AsInt64() != 200 {
		t.Errorf("thiếu thuộc tính http.route / status: %v", attrs)
	}
	lines := hs.logLines(t, "http request")
	if len(lines) == 0 || lines[0]["trace_id"] != traceID || lines[0]["span_id"] != s.SpanContext().SpanID().String() {
		t.Errorf("access log phải có trace_id/span_id của span: %v", lines)
	}
}

// TC05
func TestMetrics(t *testing.T) {
	hs := newHarness(t, slog.LevelInfo)
	hs.get("/test/1", nil)
	hs.get("/fail", nil)

	var rm metricdata.ResourceMetrics
	if err := hs.metrics.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "http.server.request.duration" {
				continue
			}
			hist, ok := m.Data.(metricdata.Histogram[float64])
			if !ok {
				t.Fatalf("http.server.request.duration không phải histogram: %T", m.Data)
			}
			routes := map[string]int64{}
			for _, dp := range hist.DataPoints {
				route, _ := dp.Attributes.Value("http.route")
				status, _ := dp.Attributes.Value("http.response.status_code")
				routes[route.AsString()+" "+status.Emit()] += int64(dp.Count)
			}
			if routes["/test/{id} 200"] != 1 || routes["/fail 502"] != 1 {
				t.Fatalf("datapoint theo route/status sai: %v", routes)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("không có metric http.server.request.duration")
	}
}
