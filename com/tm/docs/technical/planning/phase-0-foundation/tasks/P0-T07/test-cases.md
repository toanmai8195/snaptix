# Test cases — P0-T07: Skeleton core service

> Viết ở bước 1. **Tự duyệt** theo chỉ đạo người dùng (Phase 0 chưa có logic nghiệp vụ). Task: xem [README phase](../../README.md).

Chương trình: `com/tm/server/services/core/cmd/server`. PG chạy bằng compose project tạm `snaptix-test`.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T07-TC01 | Script | `go build ./...`, `go vet ./...`, gazelle, `bazel build //services/core/...` | Đều pass; gazelle không tạo diff sau khi đã chạy; có target image `server_image` (từ macro `com_tm_go_image`) | ✅ |
| P0-T07-TC02 | Script | Chạy core không đặt biến môi trường | Dùng cấu hình mặc định (`:8080`, PG `localhost:5432/core`, `LOG_LEVEL=info`); log khởi động là **một dòng JSON** có `time`, `level`, `msg`, `addr` | ✅ |
| P0-T07-TC03 | Script | PG đang chạy: `GET /healthz`, `GET /readyz` | Cả hai `200` với body JSON `{"status":"ok"}` | ✅ |
| P0-T07-TC04 | Script | Dừng PG: `GET /readyz`, `GET /healthz` | `/readyz` → `503` trong ≤ 3s, body có `"status":"unavailable"`; `/healthz` vẫn `200`; bật PG lại → `/readyz` về `200` (không cần restart core) | ✅ |
| P0-T07-TC05 | Script | `GET /metrics` | `200`, định dạng Prometheus, có `go_goroutines` | ✅ |
| P0-T07-TC06 | Script | `LOG_LEVEL=verbose` / route không tồn tại | Cấu hình sai → thoát mã khác 0 với lỗi nêu rõ biến sai; route lạ → `404` | ✅ |

Test nghiệm thu liên quan: **P0-AT02**, **P0-AT03** (readyz khi PG chạy / dừng), P0-AT06 (log JSON — cần thêm `trace_id` ở P0-T08/T10).
