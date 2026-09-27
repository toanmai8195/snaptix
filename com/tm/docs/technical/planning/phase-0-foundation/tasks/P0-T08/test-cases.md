# Test cases — P0-T08: Khung testcontainers-go cho integration test với PG

> Viết ở bước 1, chờ người dùng duyệt trước khi code. Task: xem [README phase](../../README.md).

Mọi lệnh chạy trong `com/tm/server`, cần Docker đang chạy. Quy ước theo [project-structure](../../../../project-structure.md#thiết-lập-bazel): test cần Docker gắn `tags = ["requires-docker", "requires-network"]`, `size = "large"`; test với PG thật (testcontainers), logic thuần test không cần DB.

**Phiên bản** (Go 1.24.1): `testcontainers-go` **v0.40.0** (v0.41+ đòi Go 1.25), `goose/v3` **v3.26.0** dùng như thư viện (cùng bản với `scripts/migrate.sh`). Image PG trong test = image trong compose (`postgres:17.6-alpine`).

**Vị trí** (quy tắc 3 tầng): helper `services/core/internal/pgtest` — mới chỉ core dùng; khi stats-worker cần thì mới chuyển lên `pkg/`. Integration test của core ở package riêng `services/core/integration` → một `go_test` target riêng để gắn tag Docker mà không ảnh hưởng unit test.

**Ngoài phạm vi**: P0-AT07..AT09 (shutdown) và P0-AT11 (image) — xem câu hỏi cuối file.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T08-TC01 | Integration | Test khởi động PG bằng helper, `SELECT version()` | PG 17.6 (cùng image với compose); container lên và sẵn sàng trước khi test chạy (không lỗi "connection refused" ngẫu nhiên) | ✅ |
| P0-T08-TC02 | Integration | Sau khi helper chuẩn bị DB, đọc `goose_db_version` | Migration đã áp dụng tới bản mới nhất (`00001`), migration đọc từ file embed — không phụ thuộc thư mục chạy test | ✅ |
| P0-T08-TC03 | Integration | Hai test chạy song song (`t.Parallel()`), mỗi test ghi dữ liệu vào DB của mình | Mỗi test có database riêng, không thấy dữ liệu của test kia; cả package chỉ khởi động **một** container | ✅ |
| P0-T08-TC04 | Manual | Chạy xong `go test`, xem `docker ps -a --filter label=org.testcontainers` | Không còn container test nào sót lại (kể cả khi test fail / bị Ctrl-C — nhờ Ryuk) | ✅ |
| P0-T08-TC05 | Integration | Router core (`httpx.NewRouter`) với pool thật tới PG container, gọi `/readyz` | `200` — tương ứng P0-AT02 | ✅ |
| P0-T08-TC06 | Integration | Như TC05, rồi **dừng** container PG, gọi `/readyz` và `/healthz`; khởi động lại container, gọi `/readyz` | Khi dừng: `/readyz` `503`, `/healthz` `200`. Khởi động lại: `/readyz` `200` — tương ứng P0-AT03 | ✅ |
| P0-T08-TC07 | Manual | `go test -short ./...`; và `DOCKER_HOST=unix:///khong-co go test ./...` | Integration test được **skip** (có lý do), không fail; unit test vẫn chạy | ✅ |
| P0-T08-TC08 | Manual | `bazel test //...`; `bazel test --test_tag_filters=-requires-docker //...` | Lần 1: integration target chạy và pass. Lần 2: integration target bị loại, unit test vẫn chạy | ✅ |
| P0-T08-TC09 | Manual | `bazel run //:gazelle` (2 lần), xem BUILD của integration | Lần 2 không đổi; gazelle giữ `tags`, `size` đã thêm tay; `embedsrcs` cho migration được sinh đúng | ✅ |
| P0-T08-TC10 | Manual | `go vet ./...`, `go test -race ./...`, `bazel build //...`, `bazel test //...` | Tất cả pass; dependency mới chỉ khai báo trong `go.mod`, `MODULE.bazel` chỉ thêm `use_repo` | ✅ |

Test nghiệm thu liên quan: P0-AT02, P0-AT03 (TC05, TC06) — đổi ⬜ → ✅ khi integration test pass.

## Kế hoạch subtask

| # | Làm gì | File | Kiến thức mới |
|---|---|---|---|
| 2.1 | Nhúng migration vào code: package `db/core/migrations` với `//go:embed *.sql` → `var FS embed.FS`; gazelle sinh `embedsrcs` | `db/core/migrations/migrations.go`, `BUILD.bazel` | `go:embed` (file thành dữ liệu trong binary); vì sao không đọc theo đường dẫn tương đối (`go test` chạy ở thư mục package, Bazel chạy trong sandbox/runfiles → đường dẫn khác nhau; embed chạy giống nhau ở cả hai); `embed` không lấy được file ở thư mục cha → file Go đặt cạnh migration |
| 2.2 | Integration test đầu tiên, viết thẳng bằng testcontainers-go: package `services/core/integration`, khởi động module `postgres` (`postgres:17.6-alpine`, chờ sẵn sàng), mở `pgxpool`, `SELECT version()`; skip khi `-short` / không có Docker; thêm tay `tags`, `size` vào BUILD; chạy bằng `go test` và `bazel test` | `services/core/integration/pg_test.go`, `BUILD.bazel`, `go.mod`, `MODULE.bazel`, có thể `.bazelrc` | testcontainers-go: container sống theo test, wait strategy (vì sao "port mở" chưa đủ với PG), `Terminate` + `t.Cleanup`, Ryuk dọn container khi process chết; `testing.Short()`; tag Bazel `requires-docker` / `requires-network`, `size = "large"`; Bazel sandbox + Docker socket (cần gì để test trong Bazel gọi được Docker) |
| 2.3 | Tách helper `internal/pgtest`: `TestMain` khởi động **một** container cho cả package, chạy goose (thư viện, `goose.NewProvider` với FS embed) vào database template; `pgtest.NewDB(t)` tạo database riêng cho mỗi test bằng `CREATE DATABASE ... TEMPLATE ...`, tự xoá khi test xong | `services/core/internal/pgtest/pgtest.go`, `services/core/integration/main_test.go` | `TestMain(m *testing.M)`; template database của PG (copy DB có sẵn schema trong vài chục ms thay vì chạy lại migration) → cô lập test mà vẫn nhanh; goose làm thư viện (`Provider`) và `database/sql` + driver `pgx/v5/stdlib`; vì sao helper ở `internal/` chứ chưa lên `pkg/` |
| 2.4 | Integration test cho `/readyz` (P0-AT02, AT03): router core + pool thật; test AT03 dùng container **riêng** (vì phải dừng PG) — `Stop` rồi `Start` lại container | `services/core/integration/readyz_test.go`, `pgtest.go` (nếu cần hàm container riêng) | Test tương tác với hạ tầng thật (dừng/bật PG); vì sao test phá hạ tầng phải có container riêng; `httptest.NewServer` vs gọi `ServeHTTP` trực tiếp |

## Câu hỏi cho người dùng — P0-AT07..AT09, P0-AT11

Bốn test nghiệm thu này ghi loại **Integration** nhưng không cần PG:
- **AT07–AT09** (shutdown): đã được kiểm bằng `main_test.go` của P0-T06 — `serve` chạy `http.Server` thật trên cổng TCP thật với handler chậm, và manual với SIGTERM thật. Đề xuất: đánh ✅ dựa trên các test đó (ghi rõ bằng chứng trong `acceptance-tests.md`), không viết thêm ở task này.
- **AT11** (image + `docker stop`): cần image đã `bazel run ..._docker` trước khi test chạy → `go test` phụ thuộc bước build Bazel. Đề xuất: giữ **manual** (đã pass ở P0-T07-TC05/TC06), tự động hoá ở CI (P0-T10) nếu cần.
