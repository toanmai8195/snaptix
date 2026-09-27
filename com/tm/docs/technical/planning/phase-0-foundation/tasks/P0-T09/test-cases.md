# Test cases — P0-T09: Makefile: up, down, migrate, test, gazelle

> Viết ở bước 1. Người dùng cho phép làm liền phần còn lại của P0 và tự duyệt test case ("hoàn thành luôn P0"). Task: xem [README phase](../../README.md).

`Makefile` ở **gốc repo** (theo [local-setup](../../../../local-setup.md): clone → `make up && make migrate`). Mỗi target chỉ gọi công cụ đã có — không lặp lại logic: `docker compose -f deploy/docker-compose.yml`, `scripts/migrate.sh`, `go`/`bazel` trong `com/tm/server`. `make test` ở task này = vet + unit/integration test của server; P0-T10 đổi sang gọi `scripts/ci/run.sh` (chạy đúng các bước CI).

GNU Make trên macOS là 3.81 → chỉ dùng cú pháp có từ 3.81.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T09-TC01 | Manual | `make` (không tham số) và `make help` | In danh sách target kèm mô tả (`help`, `up`, `down`, `migrate`, `test`, `gazelle`), thoát `0`; không chạy lệnh nào khác | ✅ |
| P0-T09-TC02 | Manual | `make down` rồi `make up`; chạy `make up` lần 2 | `up` chỉ trả về khi PG **healthy** (`--wait`); lần 2 không lỗi, không tạo lại container | ✅ |
| P0-T09-TC03 | Manual | `make migrate` hai lần | Lần 1 áp dụng migration (hoặc báo đã mới nhất); lần 2 không lỗi — `goose up` idempotent | ✅ |
| P0-T09-TC04 | Manual | `make test`; và `make migrate` khi PG đang dừng | `make test` chạy `go vet` + `go test -race ./...` trong `com/tm/server`, pass. Lệnh con lỗi → `make` thoát mã khác 0 (không nuốt lỗi) | ✅ |
| P0-T09-TC05 | Manual | `make gazelle` | Chạy `bazel run //:gazelle`, không sinh thay đổi (`git status` sạch) | ✅ |
| P0-T09-TC06 | Manual | Từ thư mục khác: `make -C <repo> help`, `cd com/tm/server && make -C ../../.. migrate` | Chạy đúng — đường dẫn trong Makefile tính theo vị trí Makefile, không theo thư mục đang đứng | ✅ |
| P0-T09-TC07 | Manual | P0-AT01: clone repo ra thư mục tạm, compose project riêng (`COMPOSE_PROJECT_NAME`) để không đụng DB đang dùng, đo `make up && make migrate` | Xong < 5 phút; PG healthy, migration áp dụng. Ghi rõ điều kiện đo (image PG / module Go đã có trong cache) | ✅ |

Test nghiệm thu liên quan: P0-AT01 (TC07).

## Kế hoạch subtask

| # | Làm gì | File | Kiến thức mới |
|---|---|---|---|
| 2.1 | `Makefile` gốc: `help` là target mặc định (đọc comment `##` sau tên target), `up`, `down`; `.PHONY` | `Makefile` | Cú pháp Make: target, recipe (phải thụt bằng **tab**), `.PHONY` (target không phải file), `.DEFAULT_GOAL`; biến `$(MAKEFILE_LIST)`; `@` để không in lệnh; vì sao Makefile chỉ là "mục lục lệnh" gọi script có sẵn |
| 2.2 | `migrate` (→ `scripts/migrate.sh up`), `test` (→ `go vet` + `go test -race ./...`), `gazelle` (→ `bazel run //:gazelle`); đường dẫn tuyệt đối theo vị trí Makefile | `Makefile`, `local-setup.md` | Mỗi dòng recipe chạy trong **shell riêng** → `cd dir && lệnh` cùng dòng; lệnh lỗi → make dừng và trả mã lỗi; `$(dir $(abspath $(lastword $(MAKEFILE_LIST))))` để `make -C` từ nơi khác vẫn đúng |
