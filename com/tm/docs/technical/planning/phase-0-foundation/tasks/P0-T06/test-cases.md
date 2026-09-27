# Test cases — P0-T06: Graceful shutdown (SIGTERM, http.Server.Shutdown có timeout, đóng pool sau cùng)

> Viết ở bước 1, chờ người dùng duyệt trước khi code. Task: xem [README phase](../../README.md). Challenge: G3.

Mọi lệnh `go`/`bazel` chạy trong `com/tm/server`. Test manual chạy **binary đã build** (`go build -o /tmp/core ./services/core/cmd/server`) thay vì `go run` — `go run` là process cha, signal gửi vào nó không đi thẳng tới server (bài học P0-T05). Biến mới: `CORE_SHUTDOWN_TIMEOUT` (mặc định `15s`, theo [local-setup](../../../../local-setup.md)).

**Request chậm để test manual**: chưa có endpoint chậm thật, nên dùng `/readyz` khi PG bị `docker pause` — request chạy đúng 2s (timeout ping) rồi trả 503. Bản integration với handler chậm 3s trả 200 (P0-AT07) làm ở P0-T08; ở task này kiểm chứng bằng unit test với handler chậm giả.

**Ngoài phạm vi**: rút pod khỏi load balancer trước khi dừng (readiness trả 503 + `preStop`), rolling deploy dưới tải — tiêu chí hoàn thành G3, làm khi có K8s/load test. Task này đưa G3 lên 🟨.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T06-TC01 | Manual | Core đang chạy, không có request; `kill -TERM <pid>` | Log theo thứ tự: `shutting down` (kèm tên signal) → `http server stopped` → `db pool closed`; thoát mã `0`, gần như ngay lập tức | ✅ |
| P0-T06-TC02 | Manual | Như TC01 nhưng nhấn Ctrl-C (SIGINT) | Giống TC01, thoát mã `0` | ✅ |
| P0-T06-TC03 | Unit + Manual | Request chậm đang chạy, gửi tín hiệu dừng sau 1s. Unit: handler giả ngủ 3s. Manual: `docker pause` PG, gọi `/readyz`, `kill -TERM` sau 1s | Request **hoàn thành** (unit: 200 sau ~3s; manual: nhận đủ response 503 sau ~2s, không bị reset kết nối); sau đó process thoát mã `0` | ✅ |
| P0-T06-TC04 | Unit | Sau khi bắt đầu shutdown (đang chờ request chậm), mở kết nối mới | Kết nối mới bị từ chối (listener đã đóng); request cũ vẫn hoàn thành | ✅ |
| P0-T06-TC05 | Unit + Manual | Request chạy lâu hơn timeout. Unit: timeout 500ms, handler 3s. Manual: `CORE_SHUTDOWN_TIMEOUT=500ms`, PG pause, gọi `/readyz`, `kill -TERM` | Dừng sau ~timeout (không chờ hết request), log `WARN` nêu quá shutdown timeout, pool vẫn đóng, thoát mã `1` | ✅ |
| P0-T06-TC06 | Manual | Đang chờ request chậm (như TC03, `CORE_SHUTDOWN_TIMEOUT=30s`), gửi tín hiệu lần 2 | Process thoát **ngay** (mặc định của Go khi nhận signal lần 2), không chờ hết timeout | ✅ |
| P0-T06-TC07 | Unit | `CORE_SHUTDOWN_TIMEOUT`: không đặt / `5s` / `abc` / `-1s` | `15s` / `5s` / lỗi / lỗi (timeout phải > 0); lỗi → thoát mã khác 0 lúc khởi động | ✅ |
| P0-T06-TC08 | Manual | Cổng đã bị chiếm (chạy 2 bản cùng cổng) | Bản thứ 2 log lỗi listen, `db pool closed` vẫn được log (defer chạy), thoát mã `1` | ✅ |
| P0-T06-TC09 | Manual | Hồi quy P0-T05: `/healthz`, `/readyz` (PG chạy / dừng), log JSON | Như P0-T05: 200 / 200 / 503, mọi dòng log là JSON | ✅ |
| P0-T06-TC10 | Manual | `bazel run //:gazelle` (2 lần), `go vet ./...`, `go test -race ./...`, `bazel test //...` | Gazelle lần 2 không đổi; tất cả pass; không thêm dependency mới | ✅ |

Test nghiệm thu liên quan: P0-AT07 (request chậm + SIGTERM → 200, thoát 0), P0-AT08 (sau SIGTERM request mới bị từ chối), P0-AT09 (quá timeout vẫn thoát, log cảnh báo) — task này làm cho chúng pass ở mức unit/manual; bản integration tự động ở P0-T08. Challenge G3.

## Kế hoạch subtask

| # | Làm gì | File | Kiến thức mới |
|---|---|---|---|
| 2.1 | Tách `main` → `run() error`: `main` chỉ gọi `run`, lỗi thì log + `os.Exit(1)`; mọi `defer` (vd `pool.Close()`) nằm trong `run` và log `db pool closed`. Chưa đổi hành vi | `cmd/server/main.go` | Vì sao `os.Exit` bỏ qua `defer` (thoát ngay, không unwind stack); pattern `run() error` phổ biến trong Go — một chỗ duy nhất gọi `os.Exit`, `defer` luôn chạy |
| 2.2 | Bắt signal: `signal.NotifyContext(ctx, SIGINT, SIGTERM)`; chạy server trong goroutine, gửi lỗi vào channel; `select` chờ **hoặc** signal **hoặc** server lỗi. Có signal → log `shutting down` rồi `srv.Close()` (dừng thô, chưa chờ request) | `main.go` | SIGTERM vs SIGINT vs SIGKILL (K8s/`docker stop` gửi SIGTERM, sau grace period gửi SIGKILL không bắt được); goroutine + channel + `select`; channel có buffer 1 để goroutine không bị kẹt; `ctx.Done()`; `stop()` khôi phục hành vi mặc định → signal lần 2 giết process; `http.ErrServerClosed` không phải lỗi thật |
| 2.3 | Dừng êm: thay `srv.Close()` bằng `srv.Shutdown(ctx)` với `context.WithTimeout(context.Background(), cfg.ShutdownTimeout)`; quá timeout → log `WARN`, `srv.Close()`, trả lỗi (thoát 1). Thêm `ShutdownTimeout` vào config (`time.ParseDuration`, mặc định 15s, > 0). Thứ tự: HTTP dừng → pool đóng | `main.go`, `config.go` | `Shutdown` làm gì: đóng listener (request mới bị từ chối) → đóng kết nối idle → chờ request đang chạy; vì sao context của Shutdown **không** kế thừa ctx signal (đã bị huỷ); `Shutdown` vs `Close`; vì sao đóng pool **sau** HTTP (request đang chạy còn cần DB); `time.Duration`, `ParseDuration` |
| 2.4 | Tách vòng đời server thành `serve(ctx, srv, ln net.Listener, timeout, logger) error`; `run` tự `net.Listen` rồi gọi `serve` → lỗi bind trả về ngay, test dùng listener cổng `:0` với handler chậm giả (unit test ở bước 3 chạy TC03–TC05 mà không cần signal thật) | `main.go` | `net.Listen` + `srv.Serve(ln)` thay cho `ListenAndServe`; cổng `0` = OS chọn cổng trống (test chạy song song không đụng nhau); thiết kế để test được: nhận `context` thay vì tự bắt signal bên trong |
