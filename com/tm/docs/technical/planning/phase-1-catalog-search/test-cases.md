# Test cases — Phase 1

| ID | Loại | Kịch bản | Kết quả mong đợi | Requirement / Challenge | Trạng thái |
|---|---|---|---|---|---|
| P1-TC01 | Integration | Tìm Hà Nội → Hải Phòng ngày D, có 3 chuyến | Trả 3 chuyến, sắp theo giờ đi | P1-FR1 | ⬜ |
| P1-TC02 | Integration | Tìm đoạn giữa tuyến (trạm 2 → trạm 4) | Có chuyến; giờ đi = giờ khởi hành + offset trạm 2 | P1-FR1 | ⬜ |
| P1-TC03 | Integration | Tìm ngược chiều tuyến (trạm 4 → trạm 2) | Không trả chuyến đó | P1-FR1 | ⬜ |
| P1-TC04 | Integration | Chuyến có trạng thái `CANCELLED` | Không xuất hiện trong kết quả | P1-FR1 | ⬜ |
| P1-TC05 | Unit | Tuyến có giá STANDARD 200k, VIP 300k | Giá thấp nhất = 200k | P1-FR2 | ⬜ |
| P1-TC06 | Unit | Giá cũ hết hiệu lực hôm qua, giá mới từ hôm nay | Dùng giá mới | P1-FR3 | ⬜ |
| P1-TC07 | Integration | Lấy sơ đồ ghế chuyến 40 ghế, 5 ghế SOLD | 40 ghế, đúng 5 ghế trạng thái SOLD | P1-FR4 | ⬜ |
| P1-TC08 | Integration | Lấy chuyến không tồn tại | 404, `error.code = NOT_FOUND` | P1-FR5, G7 | ⬜ |
| P1-TC09 | Integration | Tham số `date` sai định dạng | 400, `VALIDATION_ERROR`, `details` chỉ rõ field | P1-FR5, G7 | ⬜ |
| P1-TC10 | Unit | Repository trả lỗi PG không mong đợi | 500, không lộ chi tiết SQL ra response; log có đủ ngữ cảnh | G7 | ⬜ |
| P1-TC11 | Integration | Request timeout 50ms, query giả lập `pg_sleep(1)` | Trả lỗi timeout; `pg_stat_activity` không còn query đó | P1-NFR2, G2 | ⬜ |
| P1-TC12 | Integration | Client đóng kết nối giữa chừng | Context bị huỷ, query dừng | G2 | ⬜ |
| P1-TC13 | Benchmark | 10.000 truy vấn search ngẫu nhiên trên dữ liệu seed | p99 < 20ms | P1-NFR1, P7 | ⬜ |
| P1-TC14 | Manual | `EXPLAIN (ANALYZE, BUFFERS)` query search | Dùng index, không Seq Scan trên `trips`/`trip_seats` | P7 | ⬜ |
| P1-TC15 | Integration | Module khác import phần không export / truy vấn bảng của `catalog` | Kiểm tra ranh giới fail | G11 | ⬜ |
| P1-TC16 | Review | Đọc `cmd/server/main.go` | Thấy đủ đồ thị dependency, không dùng DI framework | G12 | ⬜ |
