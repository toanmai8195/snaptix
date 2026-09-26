# Test cases — P0-T02: docker-compose hạ tầng local

> Viết ở bước 1. **Tự duyệt** theo chỉ đạo người dùng (Phase 0 chưa có logic nghiệp vụ). Task: xem [README phase](../../README.md).

File: `deploy/docker-compose.yml`, cấu hình observability trong `deploy/observability/`.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T02-TC01 | Script | `docker compose -f deploy/docker-compose.yml config` | Hợp lệ; có đủ service `postgres-core`, `postgres-analytics`, `mongodb`, `redis`, `otel-collector`, `prometheus`, `grafana`, `tempo`; mọi image pin phiên bản cụ thể (không `latest`) | ✅ |
| P0-T02-TC02 | Script | `docker compose up -d --wait` | Mọi container chạy; service có healthcheck ở trạng thái `healthy` trong ≤ 120s; cổng host đúng theo [local-setup](../../../../local-setup.md): 5432, 5433, 27017, 6379, 4317/4318, 9090, 3100, 3200 | ✅ |
| P0-T02-TC03 | Script | Kết nối từng kho dữ liệu | PG core: DB `core` trả `SELECT 1`; PG analytics: DB `analytics` trả `SELECT 1`; Mongo `ping` → `ok: 1`; Redis `PING` → `PONG` | ✅ |
| P0-T02-TC04 | Script | Kiểm tra observability | Prometheus `/-/ready` OK và target `otel-collector` ở trạng thái `up`; Grafana `/api/health` OK, có datasource **Prometheus** và **Tempo** provision sẵn; gửi 1 span qua OTLP HTTP (`localhost:4318`) → truy vấn được trace đó trong Tempo theo trace ID | ✅ |
| P0-T02-TC05 | Script | Dữ liệu bền qua restart | Tạo bảng trong PG core → `docker compose down` (không `-v`) → `up` lại → bảng còn; `down -v` xoá sạch volume | ✅ |

Test nghiệm thu liên quan: P0-AT01 (một lệnh khởi động hạ tầng — cần thêm `make up` ở P0-T05).
