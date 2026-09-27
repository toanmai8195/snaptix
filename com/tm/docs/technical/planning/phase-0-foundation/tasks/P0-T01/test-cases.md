# Test cases — P0-T01: Cấu trúc thư mục com/tm/server, .gitignore cho Go/Bazel

> Viết ở bước 1, chờ người dùng duyệt trước khi code. Task: xem [README phase](../../README.md).

Nguyên tắc: **chỉ tạo thứ đang dùng**. Task này chỉ tạo `com/tm/server/` (kèm README giải thích sẽ chứa gì); các thư mục con (`services/core`, `db/...`, `pkg/`) được tạo ở task dùng tới chúng.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T01-TC01 | Manual | `git ls-files com/tm/server` | Có `com/tm/server/README.md` mô tả vai trò thư mục và cấu trúc dự kiến | ✅ |
| P0-T01-TC02 | Manual | `git check-ignore -v --no-index` với `com/tm/server/bazel-bin`, `com/tm/server/bazel-out`, `com/tm/server/core.test`, `com/tm/server/coverage.out`, `.DS_Store` | Tất cả bị ignore, `-v` chỉ ra đúng dòng pattern trong `.gitignore` | ✅ |
| P0-T01-TC03 | Manual | `git check-ignore --no-index` với `com/tm/server/go.mod`, `go.sum`, `MODULE.bazel`, `BUILD.bazel`, `.bazelrc`, `.bazelversion` | **Không** bị ignore (exit code 1) | ✅ |
| P0-T01-TC04 | Manual | `git clone` repo sang thư mục tạm | `com/tm/server/README.md` có trong bản clone | ✅ |

Test nghiệm thu liên quan: không có.

## Kế hoạch subtask

| # | Làm gì | File | Kiến thức mới |
|---|---|---|---|
| 2.1 | Tạo `com/tm/server/` với `README.md`: thư mục này chứa gì, cấu trúc dự kiến (`services/`, `pkg/`, `db/`), mỗi phần xuất hiện ở task nào | `com/tm/server/README.md` | **Git chỉ lưu file, không lưu thư mục rỗng** — muốn thư mục tồn tại trong repo phải có ít nhất một file bên trong |
| 2.2 | Thêm nhóm pattern Go / Bazel / editor vào `.gitignore`: `bazel-*` (symlink output của Bazel), `*.test` (binary test Go), `*.out` / `*.prof` (coverage, profile), `.DS_Store`, `.idea/`, `.vscode/` | `.gitignore` | Cú pháp `.gitignore`: `*` wildcard, `/` cuối = chỉ thư mục, `!` phủ định; cách kiểm tra một đường dẫn có bị ignore bằng `git check-ignore -v` |
