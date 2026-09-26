# Test cases — P0-T08: Middleware request ID, recover, access log, OTel HTTP

> Viết ở bước 1. **Tự duyệt** theo chỉ đạo người dùng (Phase 0 chưa có logic nghiệp vụ). Task: xem [README phase](../../README.md).

Middleware trong `services/core/internal/httpx`; log handler gắn ID tương quan trong `pkg/otelx`.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T08-TC01 | Unit | Request ID | Không có header → response `X-Request-ID` sinh mới (32 ký tự hex), mỗi request khác nhau; header hợp lệ gửi lên → dùng lại đúng giá trị; header không hợp lệ (quá dài, ký tự lạ) → thay bằng ID mới | ✅ |
| P0-T08-TC02 | Unit | Handler panic | Trả `500` JSON `{"error":"internal"}`; server vẫn phục vụ request tiếp theo; log mức ERROR có giá trị panic, stack, `request_id` | ✅ |
| P0-T08-TC03 | Unit | Access log | Mỗi request đúng 1 dòng JSON có `method`, `path`, `route` (pattern, vd `/test/{id}`), `status`, `bytes`, `duration_ms`, `request_id`; `/healthz`, `/readyz`, `/metrics` log ở mức DEBUG (ẩn khi level info); 5xx ở mức ERROR | ✅ |
| P0-T08-TC04 | Unit | Trace với TracerProvider test | Mỗi request tạo span server tên `GET /test/{id}`, thuộc tính `http.route`, `http.response.status_code`; request có `traceparent` → span cùng trace ID và có parent là span gửi lên; `trace_id` trong access log = trace ID của span | ✅ |
| P0-T08-TC05 | Unit | Metric với MeterProvider test | Có histogram `http.server.request.duration` với thuộc tính `http.route` và `http.response.status_code` (đúng hợp đồng dashboard RED của P0-T06) | ✅ |
| P0-T08-TC06 | Script | Chạy core thật, gửi request có `traceparent` | Response có `X-Request-ID`; dòng access log có `trace_id` bằng trace ID trong `traceparent` và `request_id` bằng header trả về | ✅ |

Test nghiệm thu liên quan: **P0-AT06** (log JSON có `trace_id`, `request_id`); P0-AT07 (trace xuyên bff → core — cần P0-T10, P0-T12).
