# Test cases — Phase 3

| ID | Loại | Kịch bản | Kết quả mong đợi | Requirement / Challenge | Trạng thái |
|---|---|---|---|---|---|
| P3-TC01 | E2E | Nạp 500.000đ, mock provider báo thành công | Số dư +500.000đ; lịch sử có 1 dòng TOPUP | P3-FR1, P3-FR2 | ⬜ |
| P3-TC02 | Integration | Webhook thành công gửi 10 lần | Chỉ 1 `ledger_transaction`; số dư +1 lần | P3-FR2, P6 | ⬜ |
| P3-TC03 | Integration | Webhook sai chữ ký HMAC | 401, không đổi số dư | P3-FR2 | ⬜ |
| P3-TC04 | Integration | Webhook đến sau khi topup đã EXPIRED | Không cộng tiền; ghi log cảnh báo để đối soát | P3-FR4 | ⬜ |
| P3-TC05 | Integration | Topup PENDING 31 phút | Worker chuyển EXPIRED | P3-FR4 | ⬜ |
| P3-TC06 | Integration | 100 request tạo topup song song cùng `Idempotency-Key` | 1 topup; 100 response giống hệt nhau | P3-FR5, P6 | ⬜ |
| P3-TC07 | Integration | Cùng key, body khác | 409 `IDEMPOTENCY_CONFLICT` | P3-FR5, P6 | ⬜ |
| P3-TC08 | Integration | Lịch sử 55 giao dịch, `limit=20` | 3 trang, không trùng không thiếu | P3-FR3 | ⬜ |
| P3-TC09 | Unit | `Ledger.Post` với entries tổng ≠ 0 | Trả lỗi, không ghi gì | P3-NFR1, P5 | ⬜ |
| P3-TC10 | Integration | Ghi thẳng SQL làm số dư ví user âm | Constraint từ chối | P3-NFR2, P5 | ⬜ |
| P3-TC11 | Integration | `UPDATE ledger_entries` bằng user ứng dụng | Bị từ chối | P5 | ⬜ |
| P3-TC12 | Integration | Hai transaction trừ cùng ví song song ở READ COMMITTED, không khoá | **Tái hiện** lost update (test đỏ ở nhánh chưa sửa) | P4 | ⬜ |
| P3-TC13 | Integration | Như TC12 sau khi áp dụng giải pháp | Số dư đúng | P4 | ⬜ |
| P3-TC14 | Integration | Chạy SERIALIZABLE, gây xung đột | Nhận `40001`, retry thành công, kết quả đúng | P4 | ⬜ |
| P3-TC15 | Integration | 200 goroutine cùng cộng/trừ một ví (`-race`) | Số dư cuối = tổng kỳ vọng; không data race | G9 | ⬜ |
| P3-TC16 | Property | Cộng/trừ `Money` ngẫu nhiên | Không overflow âm thầm; phép toán đúng | G6 | ⬜ |
| P3-TC17 | Lint | Thêm `float64` vào package wallet | Lint fail | P3-NFR3, G6 | ⬜ |
| P3-TC18 | Unit | Hàm thuần tạo bút toán cho topup/payment/refund | Tổng = 0, đúng tài khoản; test không cần DB | G5 | ⬜ |
| P3-TC19 | Integration | Use case lỗi giữa chừng (sau khi ghi entry) | Rollback toàn bộ: không entry, không đổi balance, không idempotency record | G5 | ⬜ |
| P3-TC20 | Script | Chạy script bất biến sau toàn bộ test | Tổng entries = 0; balance khớp | P3-NFR1 | ⬜ |
