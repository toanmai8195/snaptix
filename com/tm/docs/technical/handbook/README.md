# Handbook

Ghi chú kỹ thuật **theo task**: kỹ thuật đã áp dụng, lý do, bẫy gặp phải, kèm link tới đoạn code thật. Agent viết sau khi task qua bước 5 (xem [`CLAUDE.md`](../../../../../CLAUDE.md#handbook)).

| | Handbook | Lessons learned |
|---|---|---|
| Cấp độ | Task | Phase |
| Người viết | Agent | Người dùng |
| Nội dung | Kỹ thuật cụ thể + link code | Nhìn lại quá trình, số liệu, điều làm khác |

## Theo phase

| Phase | File |
|---|---|
| 0 — Nền móng | [phase-0.md](phase-0.md) |
| 1 — Catalog & tìm chuyến | [phase-1.md](phase-1.md) |
| 2 — Đăng nhập, BFF, web client | [phase-2.md](phase-2.md) |
| 3 — Ví | [phase-3.md](phase-3.md) |
| 4 — Giữ chỗ & đặt vé | [phase-4.md](phase-4.md) |
| 5 — Admin | [phase-5.md](phase-5.md) |
| 6 — Thống kê | [phase-6.md](phase-6.md) |
| 7 — Chịu tải & tối ưu | [phase-7.md](phase-7.md) |
| 8 — Huỷ/hoàn vé, đối soát, hardening | [phase-8.md](phase-8.md) |
| 9 — Tách wallet (tuỳ chọn) | [phase-9.md](phase-9.md) |

## Chỉ mục theo chủ đề

Tag gợi ý: `go/channel` · `go/errgroup` · `go/context` · `go/generics` · `go/pprof` · `go/testing` · `pg/lock` · `pg/isolation` · `pg/index` · `pg/mvcc` · `pg/partition` · `node/event-loop` · `node/stream` · `node/fastify` · `mongo/index` · `react/state` · `react/memo` · `react/query` · `bazel` · ...

| Chủ đề | Task | Bài học | Link |
|---|---|---|---|
| | | | |

## Mẫu một mục

```markdown
## P4-T03 — Use case hold ghế

### Update có điều kiện thay cho SELECT ... FOR UPDATE
- **Chủ đề**: `pg/lock`, `pg/isolation`
- **Bối cảnh**: 500 request cùng giữ ghế A05.
- **Cách làm & lý do**: `UPDATE ... WHERE status = 'AVAILABLE' RETURNING`, so số dòng trả về; không cần khoá tường minh vì ...
- **Bẫy / lưu ý**: phải sắp `seat_id` trước khi update nhiều ghế, nếu không sẽ deadlock (tái hiện ở P4-TC14).
- **Code**: [booking/store.go#L30-L55 — holdSeats](../../../server/services/core/internal/booking/store.go#L30-L55)
- **Tham khảo**: PostgreSQL docs — Explicit Locking
```
