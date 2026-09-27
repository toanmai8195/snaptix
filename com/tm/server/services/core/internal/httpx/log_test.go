package httpx

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogRequests(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus float64 // JSON number → float64 khi decode vào map[string]any
	}{
		{
			name:       "handler ghi body, không gọi WriteHeader → 200",
			handler:    func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "hi") },
			wantStatus: 200,
		},
		{
			name:       "handler không ghi gì → 200",
			handler:    func(http.ResponseWriter, *http.Request) {},
			wantStatus: 200,
		},
		{
			name:       "handler gọi WriteHeader(503)",
			handler:    func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) },
			wantStatus: 503,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))
			h := logRequests(logger)(tt.handler)

			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/x?q=1", nil))

			var line map[string]any
			if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
				t.Fatalf("log không phải một dòng JSON: %v\n%s", err, buf.String())
			}
			for _, key := range []string{"time", "level", "msg", "duration"} {
				if _, ok := line[key]; !ok {
					t.Errorf("log thiếu key %q: %v", key, line)
				}
			}
			if line["method"] != "POST" || line["path"] != "/x" {
				t.Errorf("method/path = %v %v, muốn POST /x", line["method"], line["path"])
			}
			if line["status"] != tt.wantStatus {
				t.Errorf("status = %v, muốn %v", line["status"], tt.wantStatus)
			}
		})
	}
}
