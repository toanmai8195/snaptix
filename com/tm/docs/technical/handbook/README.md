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
│   └── P0-T01.md
├── phase-1/
│   └── ...
└── phase-9/
```

## Theo task

| Task | Tên | File |
|---|---|---|
| P0-T01 | Khởi tạo cấu trúc thư mục monorepo | [phase-0/P0-T01.md](phase-0/P0-T01.md) |
| P0-T01a | Bazel + Go module + macro build image | [phase-0/P0-T01a.md](phase-0/P0-T01a.md) |
| P0-T01b | pnpm workspace cho com/tm/app | [phase-0/P0-T01b.md](phase-0/P0-T01b.md) |
| P0-T02 | docker-compose hạ tầng local | [phase-0/P0-T02.md](phase-0/P0-T02.md) |
| P0-T03 | goose migration cho PG core và analytics | [phase-0/P0-T03.md](phase-0/P0-T03.md) |
| P0-T04 | CI GitHub Actions theo đường dẫn thay đổi | [phase-0/P0-T04.md](phase-0/P0-T04.md) |
| P0-T05 | Makefile cho các thao tác thường dùng | [phase-0/P0-T05.md](phase-0/P0-T05.md) |
| P0-T06 | Dashboard Grafana RED metrics | [phase-0/P0-T06.md](phase-0/P0-T06.md) |
| P0-T07 | Skeleton core service | [phase-0/P0-T07.md](phase-0/P0-T07.md) |
| P0-T08 | Middleware request ID, recover, access log, OTel HTTP | [phase-0/P0-T08.md](phase-0/P0-T08.md) |

## Chỉ mục theo chủ đề

Tag gợi ý: `go/channel` · `go/errgroup` · `go/context` · `go/generics` · `go/pprof` · `go/testing` · `pg/lock` · `pg/isolation` · `pg/index` · `pg/mvcc` · `pg/partition` · `node/event-loop` · `node/stream` · `node/fastify` · `mongo/index` · `react/state` · `react/memo` · `react/query` · `bazel` · ...

| Chủ đề | Task | Bài học | Link |
|---|---|---|---|
| `git` | P0-T01 | Git không track thư mục rỗng; mô phỏng clone trước commit bằng index tạm | [phase-0](phase-0/P0-T01.md#git-chỉ-theo-dõi-file-không-theo-dõi-thư-mục) |
| `git` | P0-T01 | Thứ tự pattern phủ định trong `.gitignore`; `git check-ignore --no-index` | [phase-0](phase-0/P0-T01.md#thứ-tự-pattern-trong-gitignore-và-git-check-ignore) |
| `bash` | P0-T01 | Tương thích bash 3.2 (không `mapfile`); `$pipestatus` trong zsh | [phase-0](phase-0/P0-T01.md#shell-script-chạy-được-trên-bash-32-của-macos) |
| `bazel` `go/modules` | P0-T01a | Dependency Go từ go.mod qua go_deps; thứ tự go get → bazel mod tidy → gazelle | [P0-T01a](phase-0/P0-T01a.md#bzlmod-dependency-go-lấy-thẳng-từ-gomod) |
| `bazel` `gazelle` | P0-T01a | Naming convention import; map_kind go_binary → macro | [P0-T01a](phase-0/P0-T01a.md#gazelle-prefix-naming-convention-và-map_kind) |
| `bazel` `oci` | P0-T01a | Binary tĩnh + distroless, pin digest, cross-build theo --config | [P0-T01a](phase-0/P0-T01a.md#macro-com_tm_go_image-binary-tĩnh--oci-image-distroless) |
| `git` `bazel` | P0-T01a | git check-ignore không đi xuyên symlink bazel-out | [P0-T01a](phase-0/P0-T01a.md#git-check-ignore-và-symlink-của-bazel) |
| `node/pnpm` | P0-T01b | Script gốc `--recursive --if-present`; pin packageManager/engines | [P0-T01b](phase-0/P0-T01b.md#pnpm-workspace-script-gốc-chạy-đệ-quy) |
| `node/pnpm` `git` | P0-T01b | `--filter "...[ref]"` chỉ thấy file đã track | [P0-T01b](phase-0/P0-T01b.md#--filter-ref-chỉ-thấy-thay-đổi-đã-track) |
| `docker/compose` | P0-T02 | Healthcheck + `up --wait`; `$${VAR}` trong healthcheck | [P0-T02](phase-0/P0-T02.md#healthcheck--up---wait-thay-cho-sleep) |
| `observability` `otel` | P0-T02 | OTLP → collector → Tempo/Prometheus → Grafana; test bằng span tự gửi | [P0-T02](phase-0/P0-T02.md#luồng-observability-otlp--collector--tempo--prometheus--grafana) |
| `docker/compose` `testing` | P0-T02 | Test compose với project riêng để không phá dữ liệu dev | [P0-T02](phase-0/P0-T02.md#test-compose-không-phá-dữ-liệu-dev) |
| `go/tooling` | P0-T03 | Pin công cụ bằng `go run pkg@version` + build tag; GOTOOLCHAIN=local chặn tự tải toolchain | [P0-T03](phase-0/P0-T03.md#bẫy-dependency-yêu-cầu-go-mới-hơn--tự-tải-toolchain) |
| `pg/migration` | P0-T03 | goose: version của DB vs công cụ; migration tuần tự, luôn có Down | [P0-T03](phase-0/P0-T03.md#quy-ước-migration-đánh-số-tuần-tự-luôn-có-down) |
| `ci` `github-actions` | P0-T04 | Workflow mỏng, logic trong script test được; base PR vs push; tránh script injection | [P0-T04](phase-0/P0-T04.md#workflow-mỏng-logic-trong-script) |
| `bazel` `ci` | P0-T04 | Test bị ảnh hưởng bằng `rdeps`; `bazel test` exit 4 khi không có test | [P0-T04](phase-0/P0-T04.md#bazel-chỉ-test-target-bị-ảnh-hưởng-bằng-rdeps) |
| `bazel` `gazelle` `ci` | P0-T04 | `gazelle -mode=diff` chặn BUILD.bazel lỗi thời | [P0-T04](phase-0/P0-T04.md#gazelle--modediff-giữ-buildbazel-luôn-cập-nhật) |
| `make` | P0-T05 | `make help` từ comment `##`; `$$` trong recipe; biến `PROJECT ?=` | [P0-T05](phase-0/P0-T05.md#make-help-tự-sinh-từ-comment-) |
| `observability` `prometheus` | P0-T06 | RED từ metric OTel `http.server.request.duration`; tên sau collector; clamp_min, sum by le | [P0-T06](phase-0/P0-T06.md#red-và-hợp-đồng-metric-opentelemetry) |
| `otel` `testing` | P0-T06 | Test dashboard bằng histogram OTLP giả lập; rate cần ≥ 2 mẫu | [P0-T06](phase-0/P0-T06.md#test-dashboard-bằng-metric-otlp-giả-lập) |
| `go/interface` | P0-T07 | Interface ở phía dùng (`httpx.Pinger`); fake bằng func type | [P0-T07](phase-0/P0-T07.md#interface-khai-báo-ở-phía-dùng-httpxpinger) |
| `go/structure` | P0-T07 | Wiring tay, `run()` trả error để defer chạy; `pkg/` vs `internal/` | [P0-T07](phase-0/P0-T07.md#wiring-tay-trong-maingo-run-trả-error) |
| `go/context` | P0-T07 | healthz vs readyz; timeout bằng context; pgxpool mở kết nối lười | [P0-T07](phase-0/P0-T07.md#healthz-vs-readyz-timeout-bằng-context) |
| `go/http` | P0-T08 | Thứ tự middleware; route pattern chỉ có sau routing; recover + ErrAbortHandler | [P0-T08](phase-0/P0-T08.md#route-pattern-chỉ-có-sau-khi-chi-định-tuyến) |
| `go/slog` `otel` | P0-T08 | Handler gắn trace_id/request_id từ context; bọc cả WithAttrs/WithGroup | [P0-T08](phase-0/P0-T08.md#log-handler-gắn-trace_id--request_id-từ-context) |
| `go/http` `go/testing` | P0-T08 | Khoá header được chuẩn hoá — dùng Header.Add; test OTel bằng SDK in-memory | [P0-T08](phase-0/P0-T08.md#header-go-được-chuẩn-hoá-khoá) |

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
