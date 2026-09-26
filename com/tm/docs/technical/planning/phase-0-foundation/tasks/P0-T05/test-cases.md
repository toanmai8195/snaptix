# Test cases — P0-T05: Makefile cho các thao tác thường dùng

> Viết ở bước 1. **Tự duyệt** theo chỉ đạo người dùng (Phase 0 chưa có logic nghiệp vụ). Task: xem [README phase](../../README.md).

`Makefile` ở gốc repo. Biến `PROJECT` (mặc định `snaptix`) chọn compose project để test không đụng stack dev.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T05-TC01 | Script | `make` / `make help` | Liệt kê mọi target kèm mô tả (tối thiểu `up`, `down`, `migrate`, `test`, `lint`, `build`, `gazelle`); `make` không tham số = `help`, không chạy gì khác | ✅ |
| P0-T05-TC02 | Script | `make up PROJECT=snaptix-test` rồi `make ps`, `make down` | `up` trả về khi stack healthy; `ps` liệt kê service; `down` dừng mà giữ volume; `make nuke` xoá cả volume | ✅ |
| P0-T05-TC03 | Script | Sau `make up`: `make migrate` rồi `make migrate-status` | PG core và PG analytics đều ở version 1 | ✅ |
| P0-T05-TC04 | Script | `make test` | Chạy `scripts/ci/run.sh repo`, `server`, `app` — exit 0; có một test Go fail → `make test` exit khác 0 | ✅ |
| P0-T05-TC05 | Script | **P0-AT01**: `git clone` repo sang thư mục tạm, `make up && make migrate` (project riêng) | Hoàn tất trong < 5 phút, cả hai DB version 1 | ✅ |

Test nghiệm thu liên quan: **P0-AT01** (chính là TC05).
