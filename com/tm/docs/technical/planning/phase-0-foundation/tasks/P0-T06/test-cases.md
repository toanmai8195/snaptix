# Test cases — P0-T06: Dashboard Grafana RED metrics

> Viết ở bước 1. **Tự duyệt** theo chỉ đạo người dùng (Phase 0 chưa có logic nghiệp vụ). Task: xem [README phase](../../README.md).

Hợp đồng metric: service phát metric OTel chuẩn `http.server.request.duration` (histogram, giây) với thuộc tính `http.response.status_code`, `http.route`; qua collector thành `http_server_request_duration_seconds_*` có label `service_name`.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T06-TC01 | Script | Bật stack, gọi Grafana `/api/search` | Dashboard **snaptix — RED** được provision tự động (không tạo tay), có biến `service` và 3 nhóm panel: Rate (req/s), Errors (tỉ lệ 5xx), Duration (p50/p95/p99) | ✅ |
| P0-T06-TC02 | Script | Lấy mọi biểu thức PromQL của dashboard (thay `$service` → `.+`), gửi tới Prometheus `/api/v1/query` | Mọi biểu thức hợp lệ (`status: success`) | ✅ |
| P0-T06-TC03 | Script | Gửi metric OTLP giả lập trong ~30s cho service `red-probe`: request 200 và 500, thời gian khác nhau | Prometheus có `http_server_request_duration_seconds_count{service_name="red-probe"}`; truy vấn của panel Rate, Errors, Duration trả dữ liệu khác rỗng; tỉ lệ lỗi ≈ tỉ lệ 500 đã gửi | ✅ |
| P0-T06-TC04 | Script | Truy vấn giá trị của biến `service` | Danh sách service chứa `red-probe` | ✅ |

Test nghiệm thu liên quan: không có trực tiếp (dữ liệu thật đến khi core/bff phát metric ở P0-T08, P0-T10, P0-T12).
