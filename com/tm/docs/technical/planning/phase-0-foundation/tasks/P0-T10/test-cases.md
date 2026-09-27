# Test cases — P0-T10: GitHub Actions cho Go: gazelle diff, golangci-lint, bazel test target bị ảnh hưởng

> Viết ở bước 1. Người dùng cho phép làm liền phần còn lại của P0 và tự duyệt test case ("hoàn thành luôn P0"). Task: xem [README phase](../../README.md). Challenge: G14.

Theo [project-structure — CI](../../../../project-structure.md#ci): workflow `.github/workflows/ci.yml` **chỉ gọi script** trong `scripts/ci/` — chạy y hệt ở local: `scripts/ci/run.sh server [BASE]`. Job `server` chạy khi đổi `com/tm/server/**`, `.github/**`, `scripts/ci/**`; đổi `com/tm/docs/**` thì không chạy job Go. golangci-lint **v2.6.2** (chạy bằng `go run ...@v2.6.2`, pin như goose — không cần cài). Test bị ảnh hưởng: `scripts/ci/bazel-affected-tests.sh` dựa trên `rdeps`.

Máy local không có `gh` → trạng thái CI đọc qua GitHub API công khai (repo public). Kiểm AT04 bằng **push một nhánh thử** (workflow chạy cho `push` mọi nhánh và `pull_request`) thay cho mở PR, xoá nhánh sau khi kiểm.

**Ngoài phạm vi**: job `repo` (check-structure, check-gitignore, check-migrations — chưa có script) và job `app` (chặng C); kiểm tra link markdown cho `com/tm/docs/**`.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T10-TC01 | Manual | `scripts/ci/run.sh server` ở local | Chạy lần lượt: gazelle diff → golangci-lint → `bazel build //...` → `bazel test` (target bị ảnh hưởng; không có BASE → tất cả); pass, thoát `0` | ✅ |
| P0-T10-TC02 | Manual | Sửa tạm để lint lỗi (vd bỏ qua error không `_`), chạy `run.sh server` | Dừng ở bước lint, thoát khác 0, chỉ rõ file:dòng | ✅ |
| P0-T10-TC03 | Manual | Thêm tạm file Go mới mà không chạy gazelle, chạy `run.sh server` | Dừng ở bước gazelle diff, in diff BUILD cần sửa, thoát khác 0 | ✅ |
| P0-T10-TC04 | Manual | `bazel-affected-tests.sh <BASE>` với các thay đổi giả lập: (a) chỉ sửa `internal/httpx/health.go`; (b) chỉ sửa `db/core/migrations/*.sql`; (c) sửa `go.mod` / `MODULE.bazel`; (d) chỉ sửa docs | (a) `httpx_test` + `integration_test` (phụ thuộc httpx), **không** có `migrations_test`, `server_test` cũng có (server import httpx); (b) `migrations_test` + `integration_test`; (c) mọi test (`//...`); (d) rỗng | ✅ |
| P0-T10-TC05 | CI | Push lên `main` (commit của task này) | Workflow `CI` chạy job `changes` + `server`, xanh | ✅ |
| P0-T10-TC06 | CI | Push nhánh thử có lỗi lint Go | Job `server` fail ở bước lint — P0-AT04 | ✅ |
| P0-T10-TC07 | CI | Push nhánh thử chỉ sửa `com/tm/docs/**` | Job `server` bị **skip** (không chạy Go) — P0-AT05 | ✅ |
| P0-T10-TC08 | CI | Trong job `server` trên Ubuntu | Integration test (testcontainers) chạy thật, không skip — Docker của runner dùng được từ Bazel sandbox Linux | ✅ |
| P0-T10-TC09 | Manual | `make test` | Gọi `scripts/ci/run.sh server` — đúng như [local-setup](../../../../local-setup.md#build--kiểm-thử) ghi | ✅ |

Test nghiệm thu liên quan: P0-AT04 (TC06), P0-AT05 (TC07); DoD "CI xanh trên `main`" (TC05); challenge G14 (tiêu chí "CI chỉ test target bị ảnh hưởng").

## Kế hoạch subtask

| # | Làm gì | File | Kiến thức mới |
|---|---|---|---|
| 2.1 | golangci-lint v2.6.2 (`go run ...@v2.6.2`), cấu hình v2 tối thiểu (bộ `standard`: errcheck, govet, ineffassign, staticcheck, unused), chạy trên code hiện có, sửa lỗi | `com/tm/server/.golangci.yml`, code Go nếu có lỗi | golangci-lint gom nhiều linter, chạy một lần phân tích; định dạng config v2; vì sao pin phiên bản linter (bản mới thêm rule → CI đỏ mà code không đổi) |
| 2.2 | `bazel-affected-tests.sh BASE`: `git diff BASE...HEAD` → file đổi trong `com/tm/server` → label Bazel → `bazel query 'kind(".*_test rule", rdeps(//..., set(...)))'`; đổi file cấu hình chung (`MODULE.bazel`, `go.mod`, `.bazelrc`, `*.bzl`...) → mọi test | `scripts/ci/bazel-affected-tests.sh` | `bazel query`: `rdeps` (ai phụ thuộc vào tôi), `kind`, `set`; source file cũng là target (`//pkg:file.go`); `git diff A...B` (so với điểm rẽ nhánh); vì sao thay đổi cấu hình chung phải test hết |
| 2.3 | `run.sh server [BASE]`: `set -euo pipefail`, từng bước có tiêu đề (`::group::` trên GitHub), gazelle `-mode=diff`, lint, `bazel build //...`, `bazel test` các target từ 2.2; `make test` gọi script | `scripts/ci/run.sh`, `Makefile` | CI "mỏng": workflow chỉ gọi script → chạy lại đúng lỗi CI ở local; `-mode=diff` của gazelle; annotation `::group::` / `::error::` của GitHub Actions |
| 2.4 | Workflow: job `changes` (`dorny/paths-filter`) xuất `server=true/false`; job `server` (`if:`), checkout `fetch-depth: 0`, `setup-go` (theo `go.mod`), `setup-bazel` (cache Bazelisk, disk cache, repository cache), chạy `run.sh server "$BASE"` (BASE = base của PR hoặc `github.event.before`) | `.github/workflows/ci.yml` | Cấu trúc workflow (trigger, job, step, `needs`, `outputs`, `if`); vì sao lọc đường dẫn ở job thay vì `on.paths` (một workflow nhiều job, check bắt buộc vẫn có trạng thái); cache Bazel trên CI; `permissions: contents: read` |

## Ghi chú khi chạy

- TC06 (AT04): nhánh thử `ci-lint` — run push fail ở bước lint với annotation `golangci-lint báo lỗi` + `health.go:21 errcheck`. Người dùng có mở PR #2 từ nhánh này nhưng GitHub **không chạy** workflow `pull_request` khi PR đang conflict (PR #1 được squash-merge → lịch sử các nhánh thử lệch với `main`), nên kết quả lấy từ run push.
- TC07 (AT05): lần đầu **fail** — nhánh `ci-docs` chỉ sửa docs mà job `server` vẫn chạy: với push lên nhánh khác `main`, `dorny/paths-filter` mặc định so với `main`. Sửa: `base: github.event.before` cho push. Kiểm lại: commit chỉ sửa docs `a10fd4c` đẩy lên `main` → job `server` **skipped**.
- TC08: `run.sh` truyền `--test_env=PGTEST_REQUIRE_DOCKER=1` trên CI → `pgtest.Run` thoát lỗi nếu không có Docker; run `ci-try` xanh ⇒ integration test đã chạy thật (log CI cần đăng nhập mới đọc được).
- TC05: run CI của commit `ci: ... [P0-T10][G14]` trên `main` xanh.
