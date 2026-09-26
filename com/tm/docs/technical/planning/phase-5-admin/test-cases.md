# Test cases — Phase 5

| ID | Loại | Kịch bản | Kết quả mong đợi | Requirement / Challenge | Trạng thái |
|---|---|---|---|---|---|
| P5-TC01 | Integration | Tạo tuyến 4 trạm, offset tăng dần | Lưu thành công, thứ tự đúng | P5-FR1 | ⬜ |
| P5-TC02 | Integration | Tạo tuyến có offset giảm dần | 400 `VALIDATION_ERROR` | P5-FR1, R4 | ⬜ |
| P5-TC03 | Integration | Sửa sơ đồ ghế đã có vé bán | Từ chối; gợi ý tạo phiên bản mới | P5-FR1 | ⬜ |
| P5-TC04 | Integration | Đổi giá khi đã có vé bán | Vé cũ giữ giá; vé mới dùng giá mới; bản ghi giá cũ còn nguyên | P5-FR2 | ⬜ |
| P5-TC05 | Integration | Preview lịch hằng ngày T2–T6 trong 1 tháng | Đúng số slot, đúng ngày | P5-FR3 | ⬜ |
| P5-TC06 | Integration | Xác nhận sinh slot 2 lần | Không tạo slot trùng | P5-FR3 | ⬜ |
| P5-TC07 | Performance | Sinh 1 năm slot, xe 40 ghế | < 5s | P5-NFR1 | ⬜ |
| P5-TC08 | Integration | Đổi sang xe có sơ đồ thiếu ghế đã bán | Từ chối, liệt kê ghế xung đột | P5-FR4 | ⬜ |
| P5-TC09 | Integration | Tạm dừng bán slot | Hold mới bị từ chối; slot ẩn khỏi tìm kiếm | P5-FR4 | ⬜ |
| P5-TC10 | Integration | Khoá ghế A01 | A01 trạng thái BLOCKED, không giữ được | P5-FR4 | ⬜ |
| P5-TC11 | Integration | Ma trận: 4 vai trò × toàn bộ route admin | Chỉ vai trò được phép nhận 2xx; còn lại 403 | P5-FR6, N4 | ⬜ |
| P5-TC12 | Integration | Route admin mới chưa khai báo quyền | Mặc định 403 | N4 | ⬜ |
| P5-TC13 | Integration | Analyst gọi API ghi | 403, có audit log lần từ chối | N4 | ⬜ |
| P5-TC14 | Integration | Operator sửa giá | Audit log có actor, before, after, ip, thời gian | P5-FR7, M4 | ⬜ |
| P5-TC15 | Integration | Dùng user DB của BFF chạy `updateOne` / `deleteOne` trên `audit_logs` | Bị từ chối quyền | M4 | ⬜ |
| P5-TC16 | E2E | Form lịch chạy nhập ngày kết thúc trước ngày bắt đầu | Báo lỗi ở client; gửi thẳng API cũng bị 400 với cùng thông điệp | R4 | ⬜ |
| P5-TC17 | E2E | Vẽ sơ đồ 2 tầng, 30 giường bằng kéo thả, lưu, mở lại | Sơ đồ giữ nguyên | R4 | ⬜ |
| P5-TC18 | E2E | Bảng đơn hàng 100.000 dòng, cuộn nhanh, lọc theo chuyến | Cuộn mượt (≥ 50 fps), lọc < 500ms | P5-NFR2, R5 | ⬜ |
| P5-TC19 | E2E | Setup tuyến → sinh slot → đặt vé trên web client | Thành công | P5-FR1..3 | ⬜ |
