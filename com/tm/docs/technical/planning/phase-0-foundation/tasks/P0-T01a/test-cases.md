# Test cases — P0-T01a: Thiết lập Bazel + Go module + macro build image cho com/tm/server

> Viết ở bước 1, chờ người dùng duyệt trước khi code. Task: xem [README phase](../../README.md).

Mọi lệnh chạy trong `com/tm/server`.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T01a-TC01 | Script | Chạy `bazel version` | Phiên bản Bazel đang chạy đúng bằng giá trị trong `.bazelversion` (Bazelisk đọc file này) | ✅ |
| P0-T01a-TC02 | Script | Workspace mới, chưa có code Go: `bazel query //...` và `bazel build //...` | Cả hai thành công; `bazel query` liệt kê được target `//:gazelle`; `MODULE.bazel` resolve được rules_go, gazelle, rules_oci | ✅ |
| P0-T01a-TC03 | Script | Chạy `bazel run //:gazelle` hai lần liên tiếp | Lần 1 thành công; lần 2 không tạo thay đổi nào (`git status` sạch sau lần 2) — gazelle idempotent | ✅ |
| P0-T01a-TC04 | Script | Tạo package Go tạm `pkg/probe` (1 hàm + 1 test), chạy gazelle | Gazelle sinh `pkg/probe/BUILD.bazel` với target tên `probe` (không phải `go_default_library`) và `importpath = "github.com/toanmai8195/snaptix/com/tm/server/pkg/probe"`; `go build ./... && go test ./...` và `bazel build //... && bazel test //...` đều pass. Xoá package tạm sau test | ✅ |
| P0-T01a-TC05 | Script | Package tạm import thư viện ngoài (`github.com/google/uuid`): `go get` → `go mod tidy` → `bazel mod tidy` → gazelle | `go.mod`/`go.sum` có dependency; Bazel resolve dependency **từ `go.mod`** (qua `go_deps.from_file`), không khai báo tay trong `MODULE.bazel`; `bazel build //...` pass. Khôi phục `go.mod`/`go.sum`/`MODULE.bazel` sau test | ✅ |
| P0-T01a-TC06 | Script | Tạo chương trình tạm `services/probe/cmd/probe` (`package main`, in ra `probe ok`), chạy gazelle, rồi `bazel query //services/probe/...` | Gazelle sinh `com_tm_go_image(...)` thay cho `go_binary` (nhờ `map_kind`); có đủ target `probe`, `probe_image`, `probe_docker`; **không** có `probe_push` khi không truyền `repository`; `bazel run //services/probe/cmd/probe:probe` in `probe ok` | ✅ |
| P0-T01a-TC07 | Script | `bazel run --config=linux-arm64 //services/probe/cmd/probe:probe_docker` rồi `docker run --rm com.tm.go.probe:v1.0.0` | Image load vào Docker với tag `com.tm.go.probe:v1.0.0`, base distroless, kiến trúc `arm64`; container in `probe ok` và exit 0. `bazel build --config=linux-amd64 ...:probe_image` cũng pass (cross-build). Xoá chương trình tạm và image sau test | ✅ |

Test nghiệm thu liên quan: P0-AT12 (thêm package Go + gazelle, build bằng cả `go` và Bazel) — một phần; challenge G14.

**Cách chạy**: `scripts/test/bazel-setup.test.sh` (tự dọn package tạm và khôi phục `go.mod`/`MODULE.bazel`; `SKIP_DOCKER=1` để bỏ TC07). Kết quả nghiệm thu: 33 kiểm tra pass.
