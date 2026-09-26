package config

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load(env(nil))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := Config{HTTPAddr: ":8080", DatabaseURL: defaultDatabaseURL, LogLevel: slog.LevelInfo, ShutdownTimeout: 15 * time.Second}
	if cfg != want {
		t.Fatalf("Load() = %+v, muốn %+v", cfg, want)
	}
}

func TestLoad_Overrides(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"CORE_HTTP_ADDR":    ":9999",
		"CORE_DATABASE_URL": "postgres://x/y",
		"LOG_LEVEL":         "debug",
	}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTPAddr != ":9999" || cfg.DatabaseURL != "postgres://x/y" || cfg.LogLevel != slog.LevelDebug {
		t.Fatalf("Load() = %+v", cfg)
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	_, err := Load(env(map[string]string{"LOG_LEVEL": "verbose"}))
	if err == nil || !strings.Contains(err.Error(), "LOG_LEVEL") {
		t.Fatalf("muốn lỗi nêu LOG_LEVEL, có %v", err)
	}
}

func TestLoad_ShutdownTimeout(t *testing.T) {
	cfg, err := Load(env(map[string]string{"CORE_SHUTDOWN_TIMEOUT": "30s"}))
	if err != nil || cfg.ShutdownTimeout != 30*time.Second {
		t.Fatalf("CORE_SHUTDOWN_TIMEOUT=30s → %v, %v", cfg.ShutdownTimeout, err)
	}
	for _, bad := range []string{"abc", "-1s", "0s"} {
		if _, err := Load(env(map[string]string{"CORE_SHUTDOWN_TIMEOUT": bad})); err == nil || !strings.Contains(err.Error(), "CORE_SHUTDOWN_TIMEOUT") {
			t.Errorf("CORE_SHUTDOWN_TIMEOUT=%q phải lỗi nêu tên biến, có %v", bad, err)
		}
	}
}
