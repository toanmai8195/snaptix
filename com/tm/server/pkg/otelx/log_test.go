package otelx

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		in      string
		want    slog.Level
		wantErr bool
	}{
		{in: "", want: slog.LevelInfo},
		{in: "info", want: slog.LevelInfo},
		{in: "DEBUG", want: slog.LevelDebug},
		{in: " warn ", want: slog.LevelWarn},
		{in: "warning", want: slog.LevelWarn},
		{in: "error", want: slog.LevelError},
		{in: "verbose", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseLevel(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseLevel(%q) = nil error, muốn lỗi", tt.in)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("ParseLevel(%q) = %v, %v; muốn %v", tt.in, got, err, tt.want)
			}
		})
	}
}

func TestNewLogger_JSONOneLine(t *testing.T) {
	var buf bytes.Buffer
	log := NewLogger(&buf, slog.LevelInfo, "core")

	log.Debug("ẩn vì dưới level")
	log.Info("core started", slog.String("addr", ":8080"))

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	if len(lines) != 1 {
		t.Fatalf("muốn 1 dòng log (debug bị lọc), có %d: %s", len(lines), buf.String())
	}
	var rec map[string]any
	if err := json.Unmarshal(lines[0], &rec); err != nil {
		t.Fatalf("log không phải JSON: %v", err)
	}
	for k, want := range map[string]string{"level": "INFO", "msg": "core started", "service": "core", "addr": ":8080"} {
		if rec[k] != want {
			t.Errorf("%s = %v, muốn %q", k, rec[k], want)
		}
	}
	if _, ok := rec["time"]; !ok {
		t.Error("thiếu trường time")
	}
}

func TestNewLogger_CorrelationIDs(t *testing.T) {
	var buf bytes.Buffer
	log := NewLogger(&buf, slog.LevelInfo, "core").With(slog.String("component", "test"))

	tid, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	sid, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	ctx := trace.ContextWithSpanContext(context.Background(),
		trace.NewSpanContext(trace.SpanContextConfig{TraceID: tid, SpanID: sid, TraceFlags: trace.FlagsSampled}))
	ctx = WithRequestID(ctx, "req-9")

	log.InfoContext(ctx, "có ID")
	log.Info("không context")

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	var with, without map[string]any
	_ = json.Unmarshal(lines[0], &with)
	_ = json.Unmarshal(lines[1], &without)
	if with["trace_id"] != tid.String() || with["span_id"] != sid.String() || with["request_id"] != "req-9" || with["component"] != "test" {
		t.Fatalf("thiếu ID tương quan: %v", with)
	}
	for _, k := range []string{"trace_id", "span_id", "request_id"} {
		if _, ok := without[k]; ok {
			t.Errorf("log không có context không được có %s", k)
		}
	}
}
