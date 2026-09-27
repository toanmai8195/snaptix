# Test nghiệm thu — Phase 0

> Kiểm chứng requirement và challenge của cả phase; dùng cho **DoD** khi đóng phase. Một test nghiệm thu có thể pass nhờ nhiều task — đánh ✅ khi nó thực sự pass.

| ID | Loại | Kịch bản | Kết quả mong đợi | Requirement / Challenge | Trạng thái |
|---|---|---|---|---|---|
| P0-AT01 | Manual | Máy sạch, clone repo, chạy `make up && make migrate` | PG healthy, migration áp dụng, xong trong < 5 phút | P0-FR1, P0-FR2 | ✅ 3.8s (clone mới, project compose riêng; image PG + module goose đã cache) — P0-T09-TC07 |
| P0-AT02 | Integration | Gọi `core /readyz` khi PG đang chạy | 200 | P0-FR3 | ✅ `integration/readyz_test.go` — TestReadyzWithPostgres |
| P0-AT03 | Integration | Dừng PG, gọi `core /readyz` | 503; `/healthz` vẫn 200 | P0-FR3 | ✅ `integration/readyz_test.go` — TestReadyzWhenPostgresStops |
| P0-AT04 | CI | Mở PR có lỗi lint Go | CI fail ở bước lint | P0-FR4 | ⬜ |
| P0-AT05 | CI | PR chỉ sửa `com/tm/docs/**` | Không chạy job Go | G14 | ⬜ |
| P0-AT06 | Manual | Gọi 1 request, xem log core | Mỗi dòng log là JSON có `time`, `level`, `msg` | P0-NFR1 | ✅ |
| P0-AT07 | Integration | Request có handler chậm 3s, gửi SIGTERM sau 1s | Request trả 200, process thoát mã 0 | P0-NFR2, G3 | ✅ `cmd/server/main_test.go` — TestServeWaitsForInFlightRequest (http.Server thật, handler chậm) + manual SIGTERM thật (P0-T06-TC03) |
| P0-AT08 | Integration | Gửi SIGTERM rồi gửi request mới | Request mới bị từ chối | P0-NFR2, G3 | ✅ `cmd/server/main_test.go` — TestServeRejectsNewConnectionsDuringShutdown + manual (P0-T06-TC03) |
| P0-AT09 | Integration | Handler chạy quá shutdown timeout | Service vẫn thoát sau timeout, log cảnh báo | G3 | ✅ `cmd/server/main_test.go` — TestServeShutdownTimeout + manual (P0-T06-TC05) |
| P0-AT10 | Manual | Thêm package Go mới, chạy gazelle | Build/test pass bằng cả `go` và Bazel | G14 | ✅ |
| P0-AT11 | Integration | Chạy image core bằng Docker, `docker stop` | `/healthz` 200 trong container; thoát mã 0 | G14 | ✅ manual manual ở P0-T07-TC05/TC06 (image cần `bazel run ..._docker` trước — chưa tự động) |
