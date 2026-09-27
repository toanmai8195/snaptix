# Test cases — P0-T02: docker-compose chỉ có PostgreSQL core

> Viết ở bước 1, chờ người dùng duyệt trước khi code. Task: xem [README phase](../../README.md).

File: `deploy/docker-compose.yml`. Chỉ một service `postgres-core` — Mongo, Redis, PG analytics, observability thêm ở phase dùng tới.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T02-TC01 | Manual | `docker compose -f deploy/docker-compose.yml config` | Hợp lệ; đúng một service `postgres-core`; image pin phiên bản cụ thể (`postgres:17.6-alpine`, không `latest`) | ✅ |
| P0-T02-TC02 | Manual | `docker compose -f deploy/docker-compose.yml up -d --wait` | Lệnh chỉ trả về khi container ở trạng thái `healthy` (≤ 60s); cổng host 5432 đang lắng nghe | ✅ |
| P0-T02-TC03 | Manual | `docker compose ... exec postgres-core psql -U snaptix -d core -c 'SELECT current_user, current_database()'` | Trả `snaptix | core` | ✅ |
| P0-T02-TC04 | Manual | Tạo bảng thử → `down` → `up --wait` → kiểm tra bảng → `down -v` → `up --wait` → kiểm tra lại | Sau `down` + `up`: bảng **còn**; sau `down -v` + `up`: bảng **mất** | ✅ |

Test nghiệm thu liên quan: P0-AT01 (một lệnh bật hạ tầng — cần `make up` ở P0-T09).

## Kế hoạch subtask

| # | Làm gì | File | Kiến thức mới |
|---|---|---|---|
| 2.1 | Compose tối thiểu: service `postgres-core`, image `postgres:17.6-alpine`, biến `POSTGRES_USER/PASSWORD/DB`, map cổng `5432:5432`. Chạy `up -d`, vào `psql` thử | `deploy/docker-compose.yml` | Cấu trúc file compose (service, image, environment, ports); image PG chính thức tự tạo user + DB từ biến môi trường ở lần khởi động đầu; vì sao pin tag cụ thể thay vì `latest` |
| 2.2 | Thêm named volume `pg-core` gắn vào `/var/lib/postgresql/data` | `deploy/docker-compose.yml` | Container là tạm thời — dữ liệu nằm trong container sẽ mất khi container bị xoá; named volume vs bind mount; `down` (giữ volume) vs `down -v` (xoá volume) |
| 2.3 | Thêm healthcheck `pg_isready` + đặt tên project `snaptix`; dùng `up -d --wait` | `deploy/docker-compose.yml` | Container "đang chạy" ≠ PG "sẵn sàng nhận kết nối"; healthcheck + `--wait` thay cho `sleep`; `$$` để compose không tự thay biến |
