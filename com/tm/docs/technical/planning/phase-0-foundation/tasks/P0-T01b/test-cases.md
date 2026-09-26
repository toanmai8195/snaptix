# Test cases — P0-T01b: pnpm workspace cho com/tm/app

> Viết ở bước 1. **Tự duyệt** theo chỉ đạo người dùng (Phase 0 chưa có logic nghiệp vụ). Task: xem [README phase](../../README.md).

Mọi lệnh chạy trong `com/tm/app`.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T01b-TC01 | Script | `pnpm install` trên workspace mới | Thành công; tạo `pnpm-lock.yaml`; `node_modules` bị git ignore; `package.json` gốc có `packageManager` pin phiên bản pnpm và `engines.node >= 22` | ✅ |
| P0-T01b-TC02 | Script | Workspace chưa có package nào: `pnpm build`, `pnpm test`, `pnpm lint` | Exit 0 (không lỗi khi chưa có package) | ✅ |
| P0-T01b-TC03 | Script | Tạo package tạm `apps/probe` và `packages/probe-lib` (mỗi cái có script `build`/`test`/`lint`), chạy `pnpm build`, `pnpm test` | Chạy đúng script của **cả hai** package (workspace nhận `apps/*` và `packages/*`); `pnpm --filter probe build` chỉ chạy `apps/probe` | ✅ |
| P0-T01b-TC04 | Script | Với package tạm chưa commit, chạy `pnpm --filter "...[HEAD]" build` | Chỉ chạy package có thay đổi so với `HEAD` (và package phụ thuộc vào nó) — cơ chế CI dùng ở P0-T04. Xoá package tạm sau test | ✅ |

Test nghiệm thu liên quan: P0-AT05 (CI fail khi test TS fail) — cần P0-T04.
