# Test cases — P0-T03: goose migration cho PG core

> Viết ở bước 1, chờ người dùng duyệt trước khi code. Task: xem [README phase](../../README.md).

Cần PG đang chạy (`docker compose -f deploy/docker-compose.yml up -d --wait`). Chỉ PG core — PG analytics thêm ở Phase 6.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T03-TC01 | Manual | `scripts/migrate.sh --tool-version` | In `goose version: v3.26.0`; không có dòng `switching to go...` (không tự tải toolchain Go khác) | ✅ |
| P0-T03-TC02 | Manual | DB mới: `scripts/migrate.sh up` rồi `scripts/migrate.sh status` | Exit 0; `status` báo `00001_init.sql` đã Applied; bảng `goose_db_version` có version 1 | ✅ |
| P0-T03-TC03 | Manual | `up` lần 2 → `down` → `up` | `up` lần 2 không làm gì, exit 0; `down` về version 0; `up` lại về version 1 | ✅ |
| P0-T03-TC04 | Manual | `scripts/migrate.sh create add_demo sql` | Tạo `00002_add_demo.sql` (đánh số tuần tự, không phải timestamp) có sẵn `-- +goose Up` / `-- +goose Down`. Xoá file sau test | ✅ |
| P0-T03-TC05 | Manual | Gọi sai (`scripts/migrate.sh`) và `CORE_DATABASE_URL` trỏ cổng không có DB | Gọi sai → in hướng dẫn, exit 2; sai URL → lỗi kết nối nêu đúng cổng đã cấu hình | ✅ |

Test nghiệm thu liên quan: P0-AT01 (`make migrate` — cần P0-T09).

## Kế hoạch subtask

| # | Làm gì | File | Kiến thức mới |
|---|---|---|---|
| 2.1 | Chạy goose trực tiếp bằng `go run github.com/pressly/goose/v3/cmd/goose@v3.26.0 -version`, không cài global | — (chỉ chạy lệnh) | `go run <module>@<version>`: tải + build đúng phiên bản vào cache, không cần cài, không đụng `go.mod`; chọn phiên bản goose tương thích Go 1.24 (bản mới nhất đòi Go 1.26); `GOTOOLCHAIN=local` để Go không âm thầm tải toolchain khác |
| 2.2 | Viết migration đầu `00001_init.sql`, chạy `goose up` / `status` / `down` thủ công, xem bảng `goose_db_version` trong psql | `com/tm/server/db/core/migrations/00001_init.sql` | Cấu trúc file migration (`-- +goose Up` / `Down`); goose ghi nhớ migration đã chạy trong bảng `goose_db_version`; chuỗi kết nối PG (`postgres://user:pass@host:port/db?sslmode=disable`) |
| 2.3 | Gói lệnh vào `scripts/migrate.sh`: pin phiên bản, URL mặc định theo compose (ghi đè bằng `CORE_DATABASE_URL`), `create` đánh số tuần tự, `--tool-version`, in hướng dẫn khi gọi sai | `scripts/migrate.sh` | Bash script an toàn: `set -euo pipefail`, `${VAR:-default}`, `"$@"`; vì sao đánh số tuần tự (`-s`) thay vì timestamp |
