# Test cases — P0-T04: CI GitHub Actions chạy theo đường dẫn thay đổi

> Viết ở bước 1. **Tự duyệt** theo chỉ đạo người dùng (Phase 0 chưa có logic nghiệp vụ). Task: xem [README phase](../../README.md).

Workflow `.github/workflows/ci.yml` chỉ là lớp vỏ; logic nằm trong `scripts/ci/` để test được ở local.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T04-TC01 | Script | `actionlint` (pin phiên bản) trên `.github/workflows/*.yml` | Không lỗi; workflow chạy trên `pull_request` và `push` vào `main`; checkout `fetch-depth: 0`; job `server`/`app` có điều kiện theo output của job `changes`; job `repo` luôn chạy | ✅ |
| P0-T04-TC02 | Script | `scripts/ci/affected.sh` với danh sách file thay đổi | Chỉ `com/tm/server/**` → `server=true app=false`; chỉ `com/tm/app/**` → `server=false app=true`; chỉ docs/`deploy` → cả hai `false`; `.github/**` hoặc `scripts/ci/**` → cả hai `true` | ✅ |
| P0-T04-TC03 | Script | `scripts/ci/bazel-affected-tests.sh` với package tạm: `a` (có test), `b` import `a` (có test), `c` độc lập (có test) | Đổi file của `a` → test của `a` và `b`, không có `c`; đổi `c` → chỉ test `c`; đổi `MODULE.bazel`/`go.mod`/`.bazelrc` → `//...`; đổi file ngoài `com/tm/server` → rỗng | ✅ |
| P0-T04-TC04 | Script | `scripts/ci/run.sh repo` và `scripts/ci/run.sh server` | `repo`: chạy check-structure, check-gitignore, check-migrations, test script — pass. `server`: gazelle không tạo diff, golangci-lint (bỏ qua khi chưa có package Go), `bazel build //...`, test bị ảnh hưởng — pass; BUILD.bazel lỗi thời (gazelle có diff) → fail | ✅ |
| P0-T04-TC05 | Script | `scripts/ci/run.sh app` | `pnpm install --frozen-lockfile`, lint/test/build — pass; có `BASE` thì chỉ chạy package thay đổi theo `--filter "...[BASE]"`, không lỗi khi không có package nào khớp | ✅ |

Test nghiệm thu liên quan: P0-AT04, P0-AT05, P0-AT11 (cần chạy thật trên GitHub — kiểm chứng sau khi push); challenge G14 (phần "CI chỉ test target bị ảnh hưởng").
