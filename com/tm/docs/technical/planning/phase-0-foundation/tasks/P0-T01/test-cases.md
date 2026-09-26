# Test cases — P0-T01: Khởi tạo cấu trúc thư mục monorepo

> Viết ở bước 1, chờ người dùng duyệt trước khi code. Task: xem [README phase](../../README.md).

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T01-TC01 | Script | Chạy script kiểm tra cấu trúc repo | Tồn tại đủ: `com/tm/server/{api,db/core/migrations,db/analytics/migrations,pkg,services/core,services/stats-worker}`, `com/tm/app/{api,apps/bff,apps/web-client,apps/web-admin,packages/types,packages/api-client,packages/ui,packages/config}`, `com/tm/docs`, `deploy`, `loadtest`; thiếu thư mục nào thì script fail và in tên thư mục đó | ✅ |
| P0-T01-TC02 | Script | `git check-ignore` với các đường dẫn mẫu | **Bị ignore**: `com/tm/server/bazel-out`, `com/tm/server/bazel-bin`, `com/tm/app/node_modules/x`, `com/tm/app/apps/bff/dist/x`, `com/tm/app/apps/web-client/dist/x`, `.env`, `com/tm/server/coverage.out`. **Không bị ignore**: `com/tm/server/go.mod`, `com/tm/server/go.sum`, `.env.example`, `com/tm/app/pnpm-lock.yaml`, `.claude/settings.json` | ✅ |
| P0-T01-TC03 | Script | `git clone` repo sang thư mục tạm, chạy lại script của P0-T01-TC01 trên bản clone | Pass — mọi thư mục khung vẫn tồn tại sau clone (có file giữ chỗ) | ✅ |

Test nghiệm thu liên quan: không có.

**Cách chạy**: `scripts/check-structure.sh` (TC01, TC03 với tham số ROOT), `scripts/check-gitignore.sh` (TC02). TC03 trước commit được mô phỏng bằng index tạm (`GIT_INDEX_FILE` + `git checkout-index`), sau commit chạy lại bằng `git clone` thật.

> **Sửa ở P0-T01a** (người dùng cho phép tự duyệt test case P0): TC02 đổi `bazel-out/x` → `bazel-out`. Khi Bazel đã tạo symlink `bazel-out`, `git check-ignore` báo lỗi *"pathspec is beyond a symbolic link"* với đường dẫn bên trong symlink; kiểm tra chính symlink là đúng ý định (pattern `bazel-*`).
