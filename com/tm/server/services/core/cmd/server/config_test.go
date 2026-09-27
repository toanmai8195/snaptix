package main

import (
	"log/slog"
	"testing"
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
			want: config{HTTPAddr: ":8080", DatabaseURL: defaultDB, LogLevel: slog.LevelInfo},
		},
		{
			name: "biến rỗng → mặc định",
			env:  map[string]string{"CORE_HTTP_ADDR": "", "LOG_LEVEL": ""},
			want: config{HTTPAddr: ":8080", DatabaseURL: defaultDB, LogLevel: slog.LevelInfo},
		},
		{
			name: "đặt đủ biến",
			env: map[string]string{
				"CORE_HTTP_ADDR":    ":9090",
				"CORE_DATABASE_URL": "postgres://u:p@db:5432/x",
				"LOG_LEVEL":         "debug",
			},
			want: config{HTTPAddr: ":9090", DatabaseURL: "postgres://u:p@db:5432/x", LogLevel: slog.LevelDebug},
		},
		{
			name: "LOG_LEVEL không phân biệt hoa thường",
			env:  map[string]string{"LOG_LEVEL": "WARN"},
			want: config{HTTPAddr: ":8080", DatabaseURL: defaultDB, LogLevel: slog.LevelWarn},
		},
		{
			name: "LOG_LEVEL error",
			env:  map[string]string{"LOG_LEVEL": "error"},
			want: config{HTTPAddr: ":8080", DatabaseURL: defaultDB, LogLevel: slog.LevelError},
		},
		{
			name:    "LOG_LEVEL sai → lỗi",
			env:     map[string]string{"LOG_LEVEL": "verbose"},
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
