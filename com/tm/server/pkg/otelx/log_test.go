package otelx

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
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
