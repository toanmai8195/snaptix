# Test cases — Phase 0

| ID | Loại | Kịch bản | Kết quả mong đợi | Requirement / Challenge | Trạng thái |
|---|---|---|---|---|---|
| P0-TC01 | Manual | Máy sạch, clone repo, chạy `make up && make migrate` | Tất cả container healthy, migration áp dụng thành công | P0-FR1, P0-FR2 | ⬜ |
| P0-TC02 | Integration | Gọi `core /readyz` khi PG đang chạy | 200 | P0-FR3 | ⬜ |
| P0-TC03 | Integration | Dừng PG, gọi `core /readyz` | 503; `/healthz` vẫn 200 | P0-FR3 | ⬜ |
| P0-TC04 | CI | Mở PR có lỗi lint Go | CI fail ở bước lint | P0-FR4 | ⬜ |
| P0-TC05 | CI | Mở PR có test TS fail | CI fail ở bước test | P0-FR4 | ⬜ |
| P0-TC06 | Manual | Gọi 1 request bất kỳ, xem log core | Log JSON có `trace_id`, `request_id`, `status`, `latency_ms` | P0-NFR1 | ⬜ |
| P0-TC07 | Manual | Gọi `bff /healthz?deep=1`, mở Grafana | Một trace chứa span bff → core → PG | P0-NFR2, G10 | ⬜ |
| P0-TC08 | Integration | Gửi request có handler sleep 3s, gửi SIGTERM sau 1s | Request trả 200, process thoát code 0 | P0-NFR3, G3 | ⬜ |
| P0-TC09 | Integration | Gửi SIGTERM rồi gửi request mới | Request mới bị từ chối (connection refused / 503) | P0-NFR3, G3 | ⬜ |
| P0-TC10 | Integration | Handler chạy quá shutdown timeout | Service vẫn thoát sau timeout, log cảnh báo | G3 | ⬜ |
| P0-TC11 | CI | PR chỉ sửa `com/tm/app/**` | Không chạy job Bazel | G14 | ⬜ |
| P0-TC12 | Manual | Thêm package Go mới, chạy `bazel run //:gazelle` | BUILD.bazel sinh đúng; `go build ./...` và `bazel build //...` đều pass | G14 | ⬜ |
| P0-TC13 | Integration | Package ở `services/stats-worker` import `services/core/internal/...` | Build fail (visibility / `internal`) | G14 | ⬜ |
