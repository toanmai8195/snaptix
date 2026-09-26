# Test cases — Phase 7

| ID | Loại | Kịch bản | Kết quả mong đợi | Requirement / Challenge | Trạng thái |
|---|---|---|---|---|---|
| P7-TC01 | Load | k6 search 5.000 req/s, 15 phút | p99 < 150ms, lỗi < 0,1% | P7-NFR1 | ⬜ |
| P7-TC02 | Load | k6 booking 1.000/s, 15 phút | p99 < 300ms; 0 vé trùng | P7-NFR2 | ⬜ |
| P7-TC03 | Profile | pprof CPU dưới tải trước/sau tối ưu | Điểm nóng chính giảm ≥ 30% hoặc có giải thích | G8 | ⬜ |
| P7-TC04 | Profile | pprof mutex/block | Không có lock tranh chấp bất thường trong code Go | G8 | ⬜ |
| P7-TC05 | Load | Scale core 10 instance + worker | Không lỗi `too many connections`; số connection PG < giới hạn | P7-NFR3, P9 | ⬜ |
| P7-TC06 | Integration | pgx qua PgBouncer transaction mode | Không lỗi prepared statement | P9 | ⬜ |
| P7-TC07 | Integration | Đặt vé rồi đọc ngay "vé của tôi" khi replica trễ 2s | Vẫn thấy vé mới (đọc từ primary) | P11 | ⬜ |
| P7-TC08 | Load | Tìm chuyến dưới tải | ≥ 80% truy vấn đọc đi vào replica | P11 | ⬜ |
| P7-TC09 | Soak | 24h tải đều | Dead tuples `trip_seats` ổn định; kích thước bảng không tăng liên tục | P7-NFR4, P12 | ⬜ |
| P7-TC10 | Soak | 24h tải đều | RSS core và bff ổn định | P7-NFR4, N9 | ⬜ |
| P7-TC11 | Load | Xuất CSV lớn trong khi phục vụ search | Event loop lag p99 < 50ms | N1 | ⬜ |
| P7-TC12 | Integration | 3 instance BFF, gửi 40 request tìm chuyến / 10s từ 1 IP xoay vòng qua các instance | Request thứ 31 trở đi nhận 429 | P7-FR1, N5 | ⬜ |
| P7-TC13 | Load | 5.000 request cùng lúc vào 1 key cache vừa hết hạn | Chỉ 1 request đi xuống core | N7 | ⬜ |
| P7-TC14 | Load | Search có cache | Request tới core giảm ≥ 70% | N7 | ⬜ |
| P7-TC15 | Load | Mở bán với Redis pre-check | Số transaction PG thất bại do tranh chấp giảm; dữ liệu Redis khớp PG sau test | D1 | ⬜ |
| P7-TC16 | Load | Traffic mở bán gấp 10 lần, bật waiting room | p99 người đã vào < 300ms; người chờ thấy vị trí hàng đợi | P7-NFR5, D2 | ⬜ |
| P7-TC17 | Chaos | Kill Redis giữa load test booking | Latency tăng; 0 vé trùng; script bất biến ledger = 0 | P7-NFR6, D3 | ⬜ |
| P7-TC18 | Chaos | Kill 1 instance core giữa load test | Load balancer chuyển tải; request idempotent retry thành công; không mất tiền | P7-NFR2 | ⬜ |
| P7-TC19 | Chaos | Tiêm độ trễ 200ms vào PG (toxiproxy) | Circuit breaker BFF hoạt động; không treo | P7-NFR2 | ⬜ |
| P7-TC20 | Load | Core giới hạn 2 CPU / 1GB, tải booking 15 phút | Không CPU throttling đáng kể, không OOMKill; GC pause p99 có số liệu | G13 | ⬜ |
