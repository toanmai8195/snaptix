# Handbook

Ghi chú kỹ thuật **theo task**: kỹ thuật đã áp dụng, lý do, bẫy gặp phải, kèm link tới đoạn code thật. Agent viết sau khi task qua bước 5 (xem [`CLAUDE.md`](../../../../../CLAUDE.md#handbook)).

| | Handbook | Lessons learned |
|---|---|---|
| Cấp độ | Task | Phase |
| Người viết | Agent | Người dùng |
| Nội dung | Kỹ thuật cụ thể + link code | Nhìn lại quá trình, số liệu, điều làm khác |

## Cấu trúc

Mỗi task một file, nhóm theo phase:

```
handbook/
├── README.md              # file này: chỉ mục theo task và theo chủ đề
├── phase-0/
│   └── P0-T01.md    # tạo ở bước 5 của task
├── phase-1/
│   └── ...
└── phase-13/
```

## Theo task

| Task | Tên | File |
|---|---|---|
| P0-T01 | Cấu trúc thư mục com/tm/server, .gitignore cho Go/Bazel | [phase-0/P0-T01.md](phase-0/P0-T01.md) |
| P0-T02 | docker-compose chỉ có PostgreSQL core | [phase-0/P0-T02.md](phase-0/P0-T02.md) |
| P0-T03 | goose migration cho PG core | [phase-0/P0-T03.md](phase-0/P0-T03.md) |
| P0-T04 | Bazel + Gazelle + một go.mod | [phase-0/P0-T04.md](phase-0/P0-T04.md) |
| P0-T05 | Skeleton core: HTTP server, chi, config, slog, pgxpool, /readyz | [phase-0/P0-T05.md](phase-0/P0-T05.md) |
| P0-T06 | Graceful shutdown: SIGTERM, Shutdown có timeout, đóng pool sau cùng | [phase-0/P0-T06.md](phase-0/P0-T06.md) |

## Chỉ mục theo chủ đề

Tag gợi ý: `go/channel` · `go/errgroup` · `go/context` · `go/generics` · `go/pprof` · `go/testing` · `pg/lock` · `pg/isolation` · `pg/index` · `pg/mvcc` · `pg/partition` · `node/event-loop` · `node/stream` · `node/fastify` · `mongo/index` · `react/state` · `react/memo` · `react/query` · `bazel` · ...

| Chủ đề | Task | Bài học | Link |
|---|---|---|---|
| `git` | P0-T01 | Git không lưu thư mục rỗng; README giữ thư mục | [P0-T01](phase-0/P0-T01.md#git-lưu-file-không-lưu-thư-mục) |
| `git` | P0-T01 | Cú pháp `.gitignore`; `git check-ignore -v --no-index` | [P0-T01](phase-0/P0-T01.md#cú-pháp-gitignore-và-git-check-ignore) |
| `docker` `pg` | P0-T02 | Container tạm thời; named volume, `down` vs `down -v` | [P0-T02](phase-0/P0-T02.md#container-là-tạm-thời--named-volume-giữ-dữ-liệu) |
| `docker/compose` | P0-T02 | Healthcheck + `up --wait`; `$$` trong healthcheck | [P0-T02](phase-0/P0-T02.md#healthcheck--up---wait-đang-chạy--sẵn-sàng) |
| `pg/migration` | P0-T03 | Migration vs initial script; cơ chế bảng version + transaction DDL của PG | [P0-T03](phase-0/P0-T03.md#cơ-chế-goose-bảng-version--transaction-của-pg) |
| `go/tooling` | P0-T03 | `go run pkg@version`, build tag, `GOTOOLCHAIN=local` | [P0-T03](phase-0/P0-T03.md#pin-công-cụ-go-bằng-go-run-pkgversion) |
| `bash` | P0-T03 | `set -euo pipefail`, `${VAR:-default}`, `"$@"`, ROOT theo vị trí script | [P0-T03](phase-0/P0-T03.md#bash-script-an-toàn) |
| `go/modules` | P0-T04 | Module path = tiền tố import; một go.mod cho cả server | [P0-T04](phase-0/P0-T04.md#go-module-và-module-path) |
| `bazel` | P0-T04 | Hermetic, cache theo nội dung, bzlmod, Go SDK riêng | [P0-T04](phase-0/P0-T04.md#bazel-hermetic-cache-theo-nội-dung-bzlmod) |
| `bazel` `gazelle` | P0-T04 | Package / target / label; gazelle prefix + naming import | [P0-T04](phase-0/P0-T04.md#gazelle-package-target-label) |
| `go/modules` `bazel` | P0-T04 | go get → go mod tidy → bazel mod tidy → gazelle; go.sum | [P0-T04](phase-0/P0-T04.md#thêm-thư-viện-ngoài-gomod-là-nguồn-sự-thật) |
| `go/basics` `bazel` | P0-T05 | Package main, `cmd/<tên>`; `go run` để lại process con mồ côi | [P0-T05](phase-0/P0-T05.md#package-main-và-thư-mục-cmd) |
| `go/net-http` | P0-T05 | Handler / HandlerFunc; `http.Server` + `ReadHeaderTimeout` | [P0-T05](phase-0/P0-T05.md#nethttp-handler-handlerfunc-httpserver-có-timeout) |
| `go/interface` `go/pointer` | P0-T05 | ResponseWriter là interface, `*Request` là con trỏ struct | [P0-T05](phase-0/P0-T05.md#responsewriter-là-interface-request-là-con-trỏ-tới-struct) |
| `go/chi` `go/package` | P0-T05 | chi tương thích net/http; `internal/`; export bằng chữ hoa | [P0-T05](phase-0/P0-T05.md#chi-router-và-internal) |
| `go/config` `go/errors` | P0-T05 | `loadConfig(getenv)`; `fmt.Errorf` `%q` `%w` | [P0-T05](phase-0/P0-T05.md#config-từ-env-truyền-getenv-như-tham-số) |
| `go/slog` `go/middleware` | P0-T05 | slog JSON, logger tường minh; middleware bọc ResponseWriter | [P0-T05](phase-0/P0-T05.md#slog-json-và-middleware-log-request) |
| `pg/pgx` `go/context` `k8s/probe` | P0-T05 | pgxpool lười; `/readyz` + `WithTimeout`; interface phía dùng; liveness vs readiness | [P0-T05](phase-0/P0-T05.md#pgxpool-lười-readyz-với-context-timeout-interface-phía-dùng) |
| `go/defer` `go/errors` | P0-T06 | `os.Exit` bỏ qua defer → pattern `run() error` | [P0-T06](phase-0/P0-T06.md#osexit-bỏ-qua-defer--pattern-run-error) |
| `go/signal` `go/goroutine` `go/channel` | P0-T06 | SIGTERM/SIGINT/SIGKILL; goroutine + channel buffer 1 + select; signal lần 2 | [P0-T06](phase-0/P0-T06.md#signal-goroutine-channel-select) |
| `go/net-http` `go/context` | P0-T06 | `Shutdown` có timeout, `Close` khi quá hạn; timeout < grace period | [P0-T06](phase-0/P0-T06.md#httpservershutdown-có-timeout-close-khi-quá-hạn) |
| `go/testing` `go/net` | P0-T06 | `net.Listen` + cổng `:0`; nhận ctx thay cho signal để test được | [P0-T06](phase-0/P0-T06.md#thiết-kế-để-test-được-netlisten-cổng-0-ctx-thay-cho-signal) |

## Mẫu một file

```markdown
# Handbook — P4-T03: Use case hold ghế

## Update có điều kiện thay cho SELECT ... FOR UPDATE
- **Chủ đề**: `pg/lock`, `pg/isolation`
- **Bối cảnh**: 500 request cùng giữ ghế A05.
- **Cách làm & lý do**: `UPDATE ... WHERE status = 'AVAILABLE' RETURNING`, so số dòng trả về; không cần khoá tường minh vì ...
- **Bẫy / lưu ý**: phải sắp `seat_id` trước khi update nhiều ghế, nếu không sẽ deadlock (tái hiện ở P4-AT14).
- **Code**: [booking/store.go#L30-L55 — holdSeats](../../../../server/services/core/internal/booking/store.go#L30-L55)
- **Tham khảo**: PostgreSQL docs — Explicit Locking
```
