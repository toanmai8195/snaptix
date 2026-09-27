# Chạy local

> Đang phát triển — hướng dẫn sẽ cập nhật theo từng phase.

## Yêu cầu

- Docker + Docker Compose
- Go 1.24.1 (theo `com/tm/server/go.mod`), Bazelisk (đọc phiên bản Bazel từ `.bazelversion`: 8.7.0)
- Node.js 22+ và pnpm
- Không cần cài goose: `scripts/migrate.sh` chạy goose đã pin bằng `go run`

## Các bước

```bash
git clone <repo-url> snaptix && cd snaptix

# 1 + 2. Hạ tầng + migration (xem `make` để biết mọi lệnh)
make up        # = docker compose -f deploy/docker-compose.yml up -d --wait
make migrate   # = scripts/migrate.sh up   (PG core; PG analytics thêm ở Phase 6)
# tạo migration mới: scripts/migrate.sh create <ten> sql   (đánh số tuần tự)

# 3. Server (Go) — chạy trực tiếp bằng go khi dev
cd com/tm/server
go run ./services/core/cmd/server          # curl localhost:8080/readyz → 200; Ctrl-C dừng êm
# (từ Phase sau) go run ./services/core/cmd/worker, go run ./services/stats-worker/cmd/worker

# 4. App (Node + React) — từ chặng C
cd com/tm/app
pnpm install
pnpm dev   # chạy song song bff, web-client, web-admin
```

## Biến môi trường

Core đọc thẳng biến môi trường, có mặc định cho local (Phase 0 dùng 4 biến `CORE_HTTP_ADDR`, `CORE_DATABASE_URL`, `LOG_LEVEL`, `CORE_SHUTDOWN_TIMEOUT`; các biến khác dành cho phase sau).

| Biến | Dùng bởi | Ví dụ |
|---|---|---|
| `CORE_HTTP_ADDR` | core | `:8080` (mặc định) |
| `CORE_DATABASE_URL` | core | `postgres://snaptix:snaptix@localhost:5432/core?sslmode=disable` (mặc định) |
| `LOG_LEVEL` | core | `debug` \| `info` (mặc định) \| `warn` \| `error` |
| `CORE_SHUTDOWN_TIMEOUT` | core | `15s` (mặc định) — thời gian chờ request đang chạy khi nhận SIGTERM |
| `CORE_VERSION` | core | `dev` (mặc định) — `service.version` trong telemetry |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | core, bff | `http://localhost:4318` (mặc định) |
| `OTEL_SDK_DISABLED` | core, bff | `true` để tắt export |
| `OTEL_METRIC_EXPORT_INTERVAL` | core, bff | `10000` (ms, mặc định) |
| `ANALYTICS_DATABASE_URL` | stats-worker, bff | `postgres://snaptix:snaptix@localhost:5433/analytics` |
| `REDIS_URL` | core, bff | `redis://localhost:6379` |
| `BFF_HOST` / `BFF_PORT` | bff | `0.0.0.0` / `3000` (mặc định) |
| `MONGODB_URI` | bff | `mongodb://localhost:27017/snaptix` (mặc định) |
| `CORE_BASE_URL` | bff | `http://localhost:8080` |
| `CORE_SERVICE_TOKEN` | bff, core | chuỗi ngẫu nhiên |
| `AUTH_GOOGLE_ID`, `AUTH_GOOGLE_SECRET` | bff | từ Google Cloud Console |
| `SESSION_SECRET` | bff | `openssl rand -base64 32` |
| `VITE_BFF_URL` | web-client, web-admin | `http://localhost:3000` |

## Cổng mặc định

| Service | Cổng |
|---|---|
| bff | 3000 |
| web-client (Vite) | 5173 |
| web-admin (Vite) | 5174 |
| core | 8080 |
| PostgreSQL core / analytics | 5432 / 5433 |
| MongoDB | 27017 |
| Redis | 6379 |
| OTLP gRPC / HTTP (otel-collector) | 4317 / 4318 |
| Prometheus | 9090 |
| Tempo | 3200 |
| Grafana (đăng nhập ẩn danh, quyền Admin) | 3100 |

## Build & kiểm thử

`make test` chạy đúng các bước CI (`scripts/ci/run.sh server`: gazelle diff → golangci-lint → `bazel build //...` → `bazel test`; cần Docker cho integration test). `make test-go` là vòng dev nhanh (`go vet` + `go test -race`). Chi tiết từng phần:

```bash
# Server
cd com/tm/server
go test ./...                 # vòng dev nhanh (-short: bỏ integration test cần Docker)
bazel run //:gazelle          # sau khi thêm/xoá file Go hoặc import
bazel test //...              # như CI
bazel run --config=linux-arm64 //services/core/cmd/server:server_docker   # image OCI → Docker (x86: linux-amd64)

# App
cd com/tm/app
pnpm test
pnpm build

# Load test
k6 run loadtest/booking-peak.js
```
