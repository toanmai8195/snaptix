# Test cases — P0-T11: Skeleton BFF (Fastify + TypeScript)

> Viết ở bước 1. **Tự duyệt** theo chỉ đạo người dùng (Phase 0 chưa có logic nghiệp vụ). Task: xem [README phase](../../README.md).

Package `com/tm/app/apps/bff`: `buildApp(deps)` (test bằng `app.inject`) tách khỏi `server.ts`. MongoDB chạy bằng compose project tạm `snaptix-test`.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T11-TC01 | Script | `pnpm --filter bff typecheck`, `build`; chạy `dev` (tsx) | TypeScript strict pass; `tsup` sinh `dist/server.js` dạng ESM; chạy `node dist/server.js` và `pnpm dev` đều lên được | ✅ |
| P0-T11-TC02 | Script | Chạy BFF không đặt biến môi trường | Lắng nghe `:3000`; log là JSON một dòng, `level` dạng chữ (`info`), có `time`, `msg`, `service: "bff"` | ✅ |
| P0-T11-TC03 | Unit + Script | Mongo đang chạy: `GET /healthz`, `GET /readyz` | Cả hai `200` `{"status":"ok"}` | ✅ |
| P0-T11-TC04 | Unit + Script | Mongo dừng / bật lại | `/readyz` → `503` `{"status":"unavailable","error":"mongodb"}` trong ≤ 3s; `/healthz` vẫn `200`; Mongo bật lại → `/readyz` về `200` không cần restart | ✅ |
| P0-T11-TC05 | Unit | Cấu hình sai / route lạ | `LOG_LEVEL=verbose` hoặc `BFF_PORT=abc` → lỗi nêu đúng tên biến (process thoát mã khác 0); route lạ → `404` JSON | ✅ |
| P0-T11-TC06 | Unit | Request ID | Response có `x-request-id` (32 hex nếu không gửi); header hợp lệ gửi lên được dùng lại; header không hợp lệ bị thay; log request có `reqId` tương ứng | ✅ |

Test nghiệm thu liên quan: **P0-AT02**, **P0-AT03** (phía bff — `/readyz` theo MongoDB); P0-AT06 (log JSON — `trace_id` thêm ở P0-T12).
