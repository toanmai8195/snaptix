package main

import (
	"log/slog"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	const defaultDB = "postgres://snaptix:snaptix@localhost:5432/core?sslmode=disable"

	tests := []struct {
		name    string
		env     map[string]string
		want    config
		wantErr bool
	}{
		{
			name: "không đặt biến nào → mặc định",
			env:  map[string]string{},
			want: config{HTTPAddr: ":8080", DatabaseURL: defaultDB, LogLevel: slog.LevelInfo, ShutdownTimeout: 15 * time.Second},
		},
		{
			name: "biến rỗng → mặc định",
			env:  map[string]string{"CORE_HTTP_ADDR": "", "LOG_LEVEL": ""},
			want: config{HTTPAddr: ":8080", DatabaseURL: defaultDB, LogLevel: slog.LevelInfo, ShutdownTimeout: 15 * time.Second},
		},
		{
			name: "đặt đủ biến",
			env: map[string]string{
				"CORE_HTTP_ADDR":        ":9090",
				"CORE_DATABASE_URL":     "postgres://u:p@db:5432/x",
				"LOG_LEVEL":             "debug",
				"CORE_SHUTDOWN_TIMEOUT": "5s",
			},
			want: config{HTTPAddr: ":9090", DatabaseURL: "postgres://u:p@db:5432/x", LogLevel: slog.LevelDebug, ShutdownTimeout: 5 * time.Second},
		},
		{
			name: "LOG_LEVEL không phân biệt hoa thường",
			env:  map[string]string{"LOG_LEVEL": "WARN"},
			want: config{HTTPAddr: ":8080", DatabaseURL: defaultDB, LogLevel: slog.LevelWarn, ShutdownTimeout: 15 * time.Second},
		},
		{
			name: "LOG_LEVEL error",
			env:  map[string]string{"LOG_LEVEL": "error"},
			want: config{HTTPAddr: ":8080", DatabaseURL: defaultDB, LogLevel: slog.LevelError, ShutdownTimeout: 15 * time.Second},
		},
		{
			name:    "LOG_LEVEL sai → lỗi",
			env:     map[string]string{"LOG_LEVEL": "verbose"},
			wantErr: true,
		},
		{
			name: "CORE_SHUTDOWN_TIMEOUT dạng 500ms",
			env:  map[string]string{"CORE_SHUTDOWN_TIMEOUT": "500ms"},
			want: config{HTTPAddr: ":8080", DatabaseURL: defaultDB, LogLevel: slog.LevelInfo, ShutdownTimeout: 500 * time.Millisecond},
		},
		{
			name:    "CORE_SHUTDOWN_TIMEOUT không phải duration → lỗi",
			env:     map[string]string{"CORE_SHUTDOWN_TIMEOUT": "abc"},
			wantErr: true,
		},
		{
			name:    "CORE_SHUTDOWN_TIMEOUT thiếu đơn vị → lỗi",
			env:     map[string]string{"CORE_SHUTDOWN_TIMEOUT": "15"},
			wantErr: true,
		},
		{
			name:    "CORE_SHUTDOWN_TIMEOUT âm → lỗi",
			env:     map[string]string{"CORE_SHUTDOWN_TIMEOUT": "-1s"},
			wantErr: true,
		},
		{
			name:    "CORE_SHUTDOWN_TIMEOUT bằng 0 → lỗi",
			env:     map[string]string{"CORE_SHUTDOWN_TIMEOUT": "0s"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string { return tt.env[key] } // key không có → ""

			got, err := loadConfig(getenv)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("loadConfig() err = nil, muốn có lỗi")
				}
				return
			}
			if err != nil {
				t.Fatalf("loadConfig() err = %v", err)
			}
			if got != tt.want {
				t.Errorf("loadConfig() = %+v, muốn %+v", got, tt.want)
			}
		})
	}
}
