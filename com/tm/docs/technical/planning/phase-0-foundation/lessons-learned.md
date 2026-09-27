# Lessons learned — Phase 0 — Nền móng

> Viết **sau khi** đóng phase. Trung thực, cụ thể, có số liệu. Mục đích là để chính mình của 6 tháng sau đọc lại vẫn hiểu.
>
> **Bản nháp do agent viết theo yêu cầu người dùng ("hoàn thành P0 đi")** — số liệu và sự cố lấy từ phiên làm việc, commit và test case. Phần cảm nhận / điều muốn làm khác nên được người dùng đọc lại và sửa theo trải nghiệm của mình.

## Tổng quan

| | Dự kiến | Thực tế |
|---|---|---|
| Bắt đầu | — | 2026-09-27 11:47 (`a7ba5a8` reset, làm lại P0 tối giản theo quy trình subtask; lần làm đầu 2026-09-26 bị bỏ vì chạy quá nhanh, nhiều task liền) |
| Kết thúc | — | 2026-09-27 22:26 (`124c0ff`, mọi test nghiệm thu ✅) |
| Số task | 10 | 10 (P0-T01..T10), 11 commit có ID task |
| Task phát sinh | — | 0 task mới; phát sinh việc: sửa AT05 (paths-filter), bắt buộc Docker trên CI (`PGTEST_REQUIRE_DOCKER`), pin `grpc-gateway/v2` |

Cách làm: T01–T08.2.2 làm từng subtask, dừng cho người dùng đọc code; từ T08.2.3 đến hết làm liền theo yêu cầu người dùng (tự duyệt test case, tự commit + push).

## Challenge

| # | Kết quả | Cách đã giải | Link ADR / PR |
|---|---|---|---|
| G3 | 🟨 | `signal.Notify` SIGINT/SIGTERM → `http.Server.Shutdown` với `CORE_SHUTDOWN_TIMEOUT` (mặc định 15s), quá hạn → `Close` + exit 1; pool đóng sau cùng (`defer` LIFO trong `run() error`); signal lần 2 giết ngay. Chưa đạt tiêu chí "rolling deploy dưới tải" (cần K8s + readiness 503 + `preStop` + load test) | [handbook P0-T06](../../handbook/phase-0/P0-T06.md) |
| G14 | ✅ | Một `go.mod` + Bazel 8.7 bzlmod + Gazelle (`map_kind` → `com_tm_go_image`); image rules_oci distroless static nonroot, pin digest; CI `bazel test` chỉ target từ `rdeps` của file đổi | [ADR-0002](../../adr/0002-bazel-go-pnpm-ts.md), PR #1, [handbook P0-T07](../../handbook/phase-0/P0-T07.md), [P0-T10](../../handbook/phase-0/P0-T10.md) |

## Điều làm tốt

- Mỗi task có test case trước khi code, test nghiệm thu gắn bằng chứng cụ thể (file test / commit) — 11/11 AT ✅, 0 test case ⬜.
- Kiểm chứng bằng hạ tầng thật chứ không giả: PG `docker pause` để thử timeout `/readyz` (503 sau đúng 2.003s), dừng/bật PG thật trong integration test, `kill -9` để thấy Ryuk dọn container, tắt hẳn Docker Desktop để thấy test skip.
- CI "mỏng": workflow chỉ gọi `scripts/ci/run.sh` → mọi lỗi CI tái hiện được ở local bằng `make test`.
- Thiết kế để test được ngay từ đầu: `loadConfig(getenv)`, `serve(ctx, srv, ln, ...)` nhận context + listener, interface `pinger` ở phía dùng.

## Điều chưa tốt

- Lần làm P0 đầu tiên (2026-09-26) chạy 14 task trong ~1 giờ → không kịp hiểu, phải reset toàn bộ. Tốc độ ≠ tiến độ học.
- Từ T08.2.3 chuyển sang làm liền: nhanh hơn nhưng các điểm dừng đọc code (pgtest, Makefile, CI) bị bỏ — nên đọc lại handbook P0-T08..T10 để bù.
- Nhiều lần kết quả đo sai vì **môi trường**, không phải vì code (process cũ giữ cổng, script shell sai) — mất thời gian chẩn đoán.

## Kiến thức mới theo công nghệ

| Công nghệ | Điều học được | Tài liệu tham khảo |
|---|---|---|
| Go | `package main` / `cmd/`; `net/http` Handler/HandlerFunc, `http.Server` + timeout; interface vs con trỏ struct (`ResponseWriter` vs `*Request`); chi; `internal/`; `log/slog`; `context` (timeout, huỷ lan truyền); goroutine + channel buffer + `select`; `os.Exit` bỏ qua `defer` → `run() error`; `go:embed`; `TestMain`, `testing.Short()`, table-driven test, cổng `:0` | handbook [P0-T05](../../handbook/phase-0/P0-T05.md), [P0-T06](../../handbook/phase-0/P0-T06.md), [P0-T08](../../handbook/phase-0/P0-T08.md) |
| PostgreSQL | Chạy bằng Docker + healthcheck; goose (SQL thuần, DDL trong transaction); pgxpool kết nối lười; `CREATE DATABASE ... TEMPLATE` để cô lập test; không dùng DB làm template khi còn kết nối; `DROP DATABASE ... WITH (FORCE)` | [ADR-0003](../../adr/0003-goose-migrations.md), handbook [P0-T03](../../handbook/phase-0/P0-T03.md), [P0-T08](../../handbook/phase-0/P0-T08.md) |
| Bazel / CI | Platform + build chéo; rules_oci (image = config + layer tar); macro Starlark; Gazelle `map_kind`; `bazel query rdeps`; tag `requires-docker`; GitHub Actions: lọc đường dẫn ở job, `paths-filter` base khi push | handbook [P0-T07](../../handbook/phase-0/P0-T07.md), [P0-T10](../../handbook/phase-0/P0-T10.md) |
| Node.js | — (chặng B) | |
| MongoDB | — (chặng B) | |
| React | — (chặng C) | |

## Sai lầm & sự cố

| Vấn đề | Nguyên nhân gốc | Cách phát hiện | Cách khắc phục | Phòng tránh về sau |
|---|---|---|---|---|
| `POST /healthz` trả 200 thay vì 405; bind cổng 8080 lỗi | Process cũ (`bazel run` / con của `go run`) vẫn giữ cổng; `pkill` chỉ giết `go run` cha | `lsof -iTCP:8080 -sTCP:LISTEN`, PPID = 1 | Kill process mồ côi, test bằng binary `go build -o` | Test thủ công bằng binary đã build; kiểm cổng trước khi đo |
| "Dừng PG" mà `/readyz` vẫn 200 | `$C stop ...` với `C="docker compose -f ..."`: zsh không tách từ khi mở rộng biến | Các request cách nhau vài ms (stop thật mất vài giây) | Viết lệnh đầy đủ | Không nhét lệnh nhiều từ vào biến trong zsh |
| AT05 fail: push chỉ sửa docs vẫn chạy job Go | `dorny/paths-filter` với push lên nhánh khác `main` so với `main`; lịch sử lệch sau squash-merge PR #1 | Run CI nhánh `ci-docs` có job `server` success thay vì skipped | `base: github.event.before` cho push | Kiểm cả nhánh "không được chạy", không chỉ nhánh "phải chạy" |
| PR #2 conflict, workflow `pull_request` không chạy | Squash-merge PR #1 → nhánh thử tách từ `ci-try` không chung lịch sử với `main`; GitHub không chạy `pull_request` khi PR conflict | Người dùng báo conflict | Hoàn tất task trực tiếp trên `main`, xoá nhánh thử | Nhánh thử CI: tách từ `main` mới nhất, không merge |
| `go mod tidy` fail khi thêm testcontainers | Test của một dependency otel import `grpc-gateway/v2`; tidy tự tìm bản mới nhất đòi Go 1.26 | Lỗi `requires go >= 1.26.0` | Pin `grpc-gateway/v2 v2.16.0` (indirect) | Khi pin Go cũ, chọn bản thư viện theo `go` directive (testcontainers v0.40.0, pgx v5.8.0, goose v3.26.0) |
| Cleanup test lỗi `57P01` sau khi restart PG | Kết nối cũ trong pool admin đã chết | Test logic pass nhưng cleanup fail | Thử lại `DROP DATABASE` vài lần | Sau khi hạ tầng restart, coi mọi kết nối cũ là đã chết |
| Không chứng minh được integration test chạy trên CI | Log CI cần đăng nhập (API 403); Bazel không in `-v` nên test skip vẫn `PASSED` | Đọc lại cách Bazel báo kết quả | `PGTEST_REQUIRE_DOCKER=1` trên CI → thiếu Docker là fail | Test "không được skip" trong CI phải fail thay vì skip |

## Số liệu

| Chỉ số | Trước | Sau | Ghi chú |
|---|---|---|---|
| `make up && make migrate` trên clone mới | — | 3.8s | image PG + module goose đã cache (P0-AT01) |
| Image `com.tm.go.core-server:v1.0.0` | — | 17.4 MB | distroless static nonroot, arm64 |
| Build image tái lập | — | digest giống hệt sau `bazel clean` | `sha256:aaf0a59b…` |
| `docker stop` core | — | 0.26s, exit 0 | binary là PID 1, nhận SIGTERM trực tiếp |
| `/readyz` khi PG treo | — | 503 sau 2.003s | `context.WithTimeout` 2s |
| Integration test (package) | — | ~4.5s (lần đầu ~8s tải image) | một container + DB template |
| Job `server` trên CI | — | ~195s (cache trống) | setup Go + Bazel + lint + build + test |
| Test chọn theo `rdeps` | 4/4 target mỗi lần | sửa `httpx` → 3/4; sửa migration → 2/4; sửa docs → 0 | `bazel-affected-tests.sh` |

## Nếu làm lại

- Giữ nhịp subtask + dừng đọc code cho cả task nhiều khái niệm mới (pgtest, CI), chỉ làm liền với task cấu hình đơn giản (Makefile).
- Chuẩn hoá cách chạy thử thủ công ngay từ đầu: build binary, kiểm cổng, không dùng `go run` khi test signal.
- Thử CI trên nhánh tách từ `main` mới nhất và không merge nhánh thử.

## Hành động cho phase sau

- [ ] G3: khi có K8s / load test — readiness trả 503 khi bắt đầu shutdown + `preStop`, đo rolling deploy dưới tải.
- [ ] Đọc lại handbook P0-T08 (pgtest), P0-T09, P0-T10 — phần làm liền chưa được dừng đọc code.
- [ ] Phase 1: migration đầu tiên có bảng thật → kiểm `pgtest` + `migrations_test` với nhiều file; cân nhắc `-- +goose NO TRANSACTION` cho `CREATE INDEX CONCURRENTLY`.
- [ ] Theo dõi thời gian job CI khi cache ấm; nâng rules_oci khi hỗ trợ Bazel 9.
