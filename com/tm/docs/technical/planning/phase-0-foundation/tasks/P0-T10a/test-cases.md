# Test cases — P0-T10a: Image OCI cho core (server + worker)

> Viết ở bước 1. **Tự duyệt** theo chỉ đạo người dùng (Phase 0 chưa có logic nghiệp vụ). Task: xem [README phase](../../README.md).

Macro `com_tm_go_image` thêm `image_name` (tag image) và `user` (mặc định non-root 65532). `cmd/worker` là worker tối thiểu (chờ tín hiệu, dừng êm) — job thật thêm ở phase sau.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T10a-TC01 | Script | `bazel query` hai package `cmd/server`, `cmd/worker`; chạy gazelle | Có `server`, `server_image`, `server_docker`, `worker`, `worker_image`, `worker_docker`; tag `com.tm.go.core-server:v1.0.0`, `com.tm.go.core-worker:v1.0.0`; gazelle giữ nguyên `image_name` (không tạo diff) | ✅ |
| P0-T10a-TC02 | Script | `bazel build --config=linux-arm64` và `--config=linux-amd64` cả hai image | Đều pass | ✅ |
| P0-T10a-TC03 | Script | `docker run` image server (env trỏ PG test qua `host.docker.internal`) | `/healthz` 200, `/readyz` 200; log JSON ra stdout | ✅ |
| P0-T10a-TC04 | Script | `docker stop` container server | Thoát mã 0 (dừng êm nhờ SIGTERM), log có `shutting down` → `http server stopped` | ✅ |
| P0-T10a-TC05 | Script | `docker image inspect` image server | User `65532:65532` (non-root), expose `8080/tcp`, kiến trúc arm64, kích thước < 60MB | ✅ |
| P0-T10a-TC06 | Script | `docker run` image worker rồi `docker stop` | Log `worker started`; khi stop log `worker stopped`, thoát mã 0 | ✅ |

Test nghiệm thu liên quan: không có trực tiếp; challenge G14.
