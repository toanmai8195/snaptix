# Test cases — P0-T03: goose migration cho PG core và PG analytics

> Viết ở bước 1. **Tự duyệt** theo chỉ đạo người dùng (Phase 0 chưa có logic nghiệp vụ). Task: xem [README phase](../../README.md).

Công cụ: `scripts/migrate.sh <core|analytics> <lệnh goose>` (goose pin phiên bản, chạy bằng `go run`, không cần cài global). DB chạy bằng compose project tạm `snaptix-test`.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T03-TC01 | Script | `scripts/migrate.sh --tool-version` | In đúng phiên bản goose đã pin | ✅ |
| P0-T03-TC02 | Script | DB core mới: `migrate.sh core up` rồi `status` | Exit 0; bảng `goose_db_version` ở DB `core` có version 1; `status` báo migration đầu tiên đã Applied | ✅ |
| P0-T03-TC03 | Script | DB analytics mới: `migrate.sh analytics up` | Exit 0; version 1 ở DB `analytics` (cổng 5433); hai DB theo dõi version **độc lập** (migrate core không tạo bảng version ở analytics và ngược lại) | ✅ |
| P0-T03-TC04 | Script | `up` lần 2 → `down` → `up` | `up` lần 2 không đổi gì (vẫn version 1, exit 0); `down` về version 0; `up` lại về 1 | ✅ |
| P0-T03-TC05 | Script | `scripts/check-migrations.sh` trên thư mục migration | Pass với migration hiện có; fail nếu có file sai quy ước tên `NNNNN_ten.sql`, trùng số thứ tự, hoặc thiếu `-- +goose Up` / `-- +goose Down` | ✅ |

Test nghiệm thu liên quan: P0-AT01 (`make migrate` — cần P0-T05).

> TC01 sửa trong bước 2 (tự duyệt P0): `goose version` là lệnh xem version **của DB** (cần kết nối); phiên bản công cụ lấy bằng cờ `-version`, bọc thành `migrate.sh --tool-version`.
