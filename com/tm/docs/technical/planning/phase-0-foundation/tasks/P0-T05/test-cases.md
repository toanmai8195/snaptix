# Test cases — P0-T05: Skeleton core (HTTP server, chi, /healthz, config, slog, pgxpool, /readyz)

> Viết ở bước 1, chờ người dùng duyệt trước khi code. Task: xem [README phase](../../README.md).

Mọi lệnh `go`/`bazel` chạy trong `com/tm/server`; PG khởi động bằng `docker compose -f deploy/docker-compose.yml up -d --wait` từ gốc repo. Biến môi trường theo [local-setup](../../../../local-setup.md): `CORE_HTTP_ADDR` (mặc định `:8080`), `CORE_DATABASE_URL` (mặc định trỏ PG compose), `LOG_LEVEL` (mặc định `info`).

**Ngoài phạm vi**: dừng êm khi SIGTERM (P0-T06) — ở task này Ctrl-C chỉ giết process; image (P0-T07); integration test tự động với testcontainers (P0-T08).

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T05-TC01 | Manual | Không đặt biến env, `go run ./services/core/cmd/server` | Process chạy, lắng nghe `:8080`; có một dòng log khởi động ghi địa chỉ lắng nghe | ✅ |
| P0-T05-TC02 | Manual | `curl -i localhost:8080/healthz` | `200`, body `ok` | ✅ |
| P0-T05-TC03 | Manual | `curl -i localhost:8080/khong-co`; `curl -i -X POST localhost:8080/healthz` | Lần lượt `404` và `405` (chi tự xử lý) | ✅ |
| P0-T05-TC04 | Manual | `CORE_HTTP_ADDR=:9090 go run ...`, gọi `/healthz` ở 9090 | Lắng nghe `:9090`, `/healthz` 200; cổng 8080 không mở | ✅ |
| P0-T05-TC05 | Manual | `LOG_LEVEL=verbose go run ...` | Thoát mã khác 0 ngay khi khởi động, log lỗi nêu rõ giá trị `LOG_LEVEL` không hợp lệ | ✅ |
| P0-T05-TC06 | Manual | Gọi 1 request `/healthz`, xem stdout | Mỗi dòng log là một object JSON có `time`, `level`, `msg`; có dòng log request gồm `method`, `path`, `status`, `duration` | ✅ |
| P0-T05-TC07 | Manual | `LOG_LEVEL=debug` so với mặc định | Chỉ khi `debug` mới thấy các dòng `"level":"DEBUG"` | ✅ |
| P0-T05-TC08 | Manual | PG đang chạy, `curl -i localhost:8080/readyz` | `200`, body `ok` | ✅ |
| P0-T05-TC09 | Manual | Đang chạy core, `docker compose ... stop postgres-core`, gọi `/readyz` và `/healthz`; rồi `start` lại PG, gọi `/readyz` | Khi PG dừng: `/readyz` `503` trong ≤ 2 giây (có timeout, không treo), `/healthz` vẫn `200`. PG chạy lại: `/readyz` về `200` mà **không** restart core | ✅ |
| P0-T05-TC10 | Manual | PG đang dừng, khởi động core; và `CORE_DATABASE_URL=khong-phai-url go run ...` | PG dừng: core vẫn khởi động (pool kết nối lười), `/healthz` 200, `/readyz` 503. URL sai cú pháp: thoát mã khác 0, log lỗi cấu hình | ✅ |
| P0-T05-TC11 | Manual | `bazel run //:gazelle` (2 lần), `go vet ./...`, `go build ./...`, `bazel build //...` | Gazelle lần 2 không đổi gì; build bằng cả `go` và Bazel pass; dependency mới (`chi`, `pgx`) chỉ khai báo trong `go.mod`, `MODULE.bazel` chỉ thêm `use_repo` | ✅ |

Test nghiệm thu liên quan: P0-AT02 (`/readyz` 200 khi PG chạy), P0-AT03 (PG dừng → `/readyz` 503, `/healthz` 200), P0-AT06 (log JSON có `time`, `level`, `msg`) — ở task này kiểm tra manual; bản integration tự động làm ở P0-T08.

## Kế hoạch subtask

Mỗi subtask chạy được bằng `go run ./services/core/cmd/server` và thêm **một** khái niệm mới.

| # | Làm gì | File | Kiến thức mới |
|---|---|---|---|
| 2.1 | Hello world: `main` in một dòng ra stdout; gazelle sinh `go_binary` | `services/core/cmd/server/main.go` | Package `main` và hàm `main`; vì sao binary nằm trong `cmd/<tên>/` (một service có thể có nhiều binary: `server`, `worker`); `go run` vs `go build` vs `bazel run` |
| 2.2 | HTTP server bằng thư viện chuẩn: `http.Server{Addr: ":8080", Handler: mux}` với một `http.HandleFunc("/healthz", ...)` trả `ok` | `main.go` | `net/http`: `Handler` là interface một method `ServeHTTP`, `HandlerFunc` là adapter; `ResponseWriter`, `*Request`; mỗi request chạy trên goroutine riêng; vì sao tạo `http.Server` thay vì `http.ListenAndServe` (để P0-T06 gọi `Shutdown`, đặt timeout); `ReadHeaderTimeout` chống slowloris |
| 2.3 | Thay mux chuẩn bằng **chi**: `chi.NewRouter()`, `r.Get("/healthz", ...)`; 404/405 tự có. Tách handler sang package `httpx` | `internal/httpx/router.go`, `internal/httpx/health.go`, `main.go`, `go.mod`, `MODULE.bazel` | Vì sao chi (tương thích `net/http`, handler vẫn là `http.Handler`, middleware là `func(http.Handler) http.Handler`) thay vì gin/echo (API riêng, không theo chuẩn); `internal/` giới hạn ai được import; lần đầu thêm dependency thật: `go get` → `bazel mod tidy` → gazelle |
| 2.4 | Config từ env: struct `config` với `HTTPAddr`, `DatabaseURL`, `LogLevel`; hàm `loadConfig(getenv func(string) string) (config, error)` có giá trị mặc định; lỗi → in ra stderr, `os.Exit(1)` | `cmd/server/config.go`, `main.go` | `os.Getenv` vs `os.LookupEnv`; truyền `getenv` như tham số để test không đụng env thật; error value thay cho exception, `fmt.Errorf("...: %w", err)`; vì sao config nằm trong package `main` (chỉ `main` dùng — quy tắc 3 tầng, không tạo `pkg/config`); không dùng viper/envconfig — 3 biến chưa đáng thêm thư viện |
| 2.5 | Log `slog` JSON: `slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: ...}))`, parse `LOG_LEVEL`; middleware chi ghi log mỗi request (`method`, `path`, `status`, `duration`) | `main.go`, `config.go`, `internal/httpx/log.go` | `log/slog` (thư viện chuẩn từ Go 1.21): logger có cấu trúc, key-value, handler JSON; truyền `*slog.Logger` tường minh thay vì biến global; viết middleware: bọc `ResponseWriter` để bắt status code (dùng `middleware.WrapResponseWriter` của chi) |
| 2.6 | pgxpool + `/readyz`: `pgxpool.New(ctx, url)` trong `main`, `defer pool.Close()`; `/readyz` gọi `Ping` với `context.WithTimeout` 2s → 200 hoặc 503. Handler nhận interface `pinger` (1 method) khai báo ở `httpx`, không nhận thẳng `*pgxpool.Pool` | `main.go`, `internal/httpx/health.go`, `go.mod`, `MODULE.bazel` | `pgx/v5` và `pgxpool` (vì sao pgx thay vì `database/sql` + lib/pq); pool kết nối **lười** — `New` không mở kết nối nên core chạy được khi PG chưa lên; `context.Context`: truyền xuống, deadline, `r.Context()`; interface ở phía dùng — handler chỉ cần `Ping`, test dùng fake không cần DB; liveness vs readiness (K8s không restart pod khi chỉ DB chết) |
