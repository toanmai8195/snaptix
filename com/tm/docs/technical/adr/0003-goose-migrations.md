# ADR-0003: goose cho migration PostgreSQL

- **Trạng thái**: Chấp nhận
- **Ngày**: 2026-09-27

## Bối cảnh

- Schema PG core (và analytics từ Phase 6) phải thay đổi có kiểm soát, chạy được ở máy dev, trong integration test và khi deploy.
- Dự án học PostgreSQL → migration nên là **SQL thuần** để đọc/viết SQL thật, không qua DSL.
- Không muốn bắt cài công cụ thủ công; phiên bản phải cố định.

## Các phương án

1. **golang-migrate** — Ưu: phổ biến. Nhược: file up/down tách riêng; dùng làm thư viện kém gọn hơn.
2. **Atlas** (khai báo schema, tự sinh diff) — Ưu: mạnh. Nhược: che mất việc tự viết DDL — trái mục tiêu học PG; thêm khái niệm.
3. **goose** — Ưu: SQL thuần, up/down trong một file (`-- +goose Up/Down`), mỗi migration chạy trong transaction (PG hỗ trợ DDL trong transaction), đánh số tuần tự (`-s`), dùng được cả CLI lẫn thư viện (`goose.NewProvider` với `fs.FS`). Nhược: ít tính năng "tự động" hơn Atlas.

## Quyết định

Chọn **goose v3.26.0** (bản mới hơn đòi Go ≥ 1.26):
- CLI: `scripts/migrate.sh` chạy goose bằng `go run ...@v3.26.0` (pin, không cần cài); `make migrate` gọi script.
- File: `com/tm/server/db/core/migrations/NNNNN_ten.sql`, chỉ thêm mới, không sửa migration đã merge.
- Test: migration nhúng vào code (`//go:embed *.sql`, package `db/core/migrations`); `pgtest` chạy goose như thư viện vào database template, mỗi test copy bằng `CREATE DATABASE ... TEMPLATE`.

## Hệ quả

- Dễ hơn: một nguồn migration dùng cho dev, test, deploy; unit test kiểm quy ước tên và đánh số liên tục.
- Khó hơn: đổi schema phải tự viết cả `Down`; thêm migration phải chạy gazelle (file mới vào `embedsrcs`).
- Theo dõi: khi có migration dài/khoá bảng lớn (Phase 1+), cân nhắc `-- +goose NO TRANSACTION` cho `CREATE INDEX CONCURRENTLY`.
