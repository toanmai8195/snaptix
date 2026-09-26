# Test cases — P0-T09: Graceful shutdown cho core

> Viết ở bước 1. **Tự duyệt** theo chỉ đạo người dùng (Phase 0 chưa có logic nghiệp vụ). Task: xem [README phase](../../README.md).

Logic dừng: `httpx.Serve(ctx, srv, listener, log, timeout)`; `main` nối với `signal.NotifyContext(SIGINT, SIGTERM)`. Unit test chạy `http.Server` thật qua TCP (cổng `:0`).

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T09-TC01 | Unit | Huỷ context khi không có request nào | `Serve` trả `nil` nhanh (< 1s); log `shutting down` rồi `http server stopped` | ✅ |
| P0-T09-TC02 | Unit | Request chậm 500ms đang chạy, huỷ context sau 100ms | Request vẫn hoàn thành `200` đầy đủ body; `Serve` chỉ trả về sau khi request xong (P0-AT08) | ✅ |
| P0-T09-TC03 | Unit | Sau khi bắt đầu dừng, mở kết nối mới | Kết nối mới bị từ chối (P0-AT09) | ✅ |
| P0-T09-TC04 | Unit | Handler chạy lâu hơn timeout (timeout 200ms, handler 5s) | `Serve` trả về sau ≈ timeout (< 1s) với lỗi bọc `context.DeadlineExceeded`; log WARN; kết nối bị đóng cưỡng bức (P0-AT10) | ✅ |
| P0-T09-TC05 | Script | Binary core: gửi SIGTERM, rồi thử với SIGINT | Thoát mã 0 trong < 2s; log theo thứ tự `shutting down` → `http server stopped` → `database pool closed` (pool đóng **sau** server) | ✅ |
| P0-T09-TC06 | Unit | Cấu hình `CORE_SHUTDOWN_TIMEOUT` | Mặc định `15s`; `30s` → 30 giây; giá trị sai (`abc`, `-1s`) → lỗi nêu tên biến | ✅ |

Test nghiệm thu liên quan: **P0-AT08**, **P0-AT09**, **P0-AT10**; challenge **G3** (phần rolling deploy dưới tải kiểm ở Phase 7).
