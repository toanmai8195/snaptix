# Test cases — P0-T04: Bazel + Gazelle + một go.mod cho com/tm/server

> Viết ở bước 1, chờ người dùng duyệt trước khi code. Task: xem [README phase](../../README.md).

Mọi lệnh chạy trong `com/tm/server`. Task này **chưa có code Go thật** (core ở P0-T05) — kiểm chứng bằng package tạm `pkg/probe`, xoá sau test.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T04-TC01 | Manual | `bazel version` | Build label đúng bằng `.bazelversion` (`8.7.0`) — Bazelisk đọc file này | ✅ |
| P0-T04-TC02 | Manual | Workspace chưa có code Go: `bazel build //...`, `bazel query //...` | Build thành công; query liệt kê `//:gazelle`; `MODULE.bazel` resolve được `rules_go`, `gazelle` | ✅ |
| P0-T04-TC03 | Manual | `bazel run //:gazelle` hai lần | Lần 2 không sinh thay đổi nào (`git status` giống sau lần 1) | ✅ |
| P0-T04-TC04 | Manual | Tạo `pkg/probe` (1 hàm + 1 test), chạy gazelle | Sinh `pkg/probe/BUILD.bazel`, target tên `probe` (không phải `go_default_library`), `importpath` = `github.com/toanmai8195/snaptix/com/tm/server/pkg/probe`; `go test ./...` và `bazel test //...` đều pass | ✅ |
| P0-T04-TC05 | Manual | `pkg/probe` import `github.com/google/uuid`: `go get` → `go mod tidy` → `bazel mod tidy` → gazelle | Dependency khai báo **chỉ trong `go.mod`**; `bazel mod tidy` tự thêm `use_repo(go_deps, "com_github_google_uuid")`; BUILD có `@com_github_google_uuid//:uuid`; `bazel build //...` pass. Khôi phục `go.mod`/`MODULE.bazel` và xoá `pkg/probe` sau test | ✅ |

Test nghiệm thu liên quan: P0-AT10 (thêm package Go + gazelle, build bằng cả `go` và Bazel); challenge G14.

## Kế hoạch subtask

| # | Làm gì | File | Kiến thức mới |
|---|---|---|---|
| 2.1 | `go mod init github.com/toanmai8195/snaptix/com/tm/server`; thử một chương trình tạm để thấy module path được dùng thế nào (xoá sau) | `com/tm/server/go.mod` | Go module là gì; module path = tiền tố import của mọi package; dòng `go 1.24.1`; một `go.mod` cho cả `com/tm/server` (mọi service dùng chung phiên bản dependency) |
| 2.2 | Bazel tối thiểu: `.bazelversion`, `MODULE.bazel` (rules_go, gazelle, Go SDK, `go_deps.from_file`), `.bazelrc`; `bazel build //...` trên workspace rỗng | `.bazelversion`, `MODULE.bazel`, `.bazelrc`, `BUILD.bazel` rỗng | Bazel khác `go build` ở đâu (hermetic, cache, build nhiều ngôn ngữ); Bazelisk chọn phiên bản; bzlmod (`bazel_dep`); Bazel tự tải Go SDK riêng; lockfile `MODULE.bazel.lock` |
| 2.3 | Target `//:gazelle` + directive `prefix`, `go_naming_convention import`; thử với `pkg/probe` tạm → gazelle sinh `BUILD.bazel` → `go test` và `bazel test` đều chạy | `BUILD.bazel` | Package Bazel, target, label (`//pkg/probe:probe`); `go_library` / `go_test`; Gazelle sinh BUILD từ import của code Go; vì sao code phải build được bằng cả `go` và Bazel |
| 2.4 | Luồng thêm thư viện ngoài: `go get` → `go mod tidy` → `bazel mod tidy` → gazelle, với `github.com/google/uuid` trong `pkg/probe` (khôi phục sau) | `go.mod`, `MODULE.bazel` (tạm) | `go.mod` là nguồn sự thật duy nhất cho dependency; `go.sum` là gì; tên repo Bazel `com_github_google_uuid`; `use_repo` |
