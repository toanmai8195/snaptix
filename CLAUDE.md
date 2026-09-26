# snaptix — Quy tắc cho Claude

Hệ thống đặt vé xe khách / tàu điện, dự án học tập để lên senior Go, PostgreSQL, Node.js, React.

| Thư mục | Nội dung | Build / test |
|---|---|---|
| `com/tm/server` | Go: core, stats-worker, `pkg/`, migration | `go test -race ./...`, `bazel run //:gazelle`, `bazel test //...` |
| `com/tm/app` | Node (Fastify BFF) + React (Vite) | `pnpm --filter <app> lint test build` |
| `com/tm/docs` | Tài liệu, planning, handbook, ADR | — |

Tài liệu quan trọng:
- Planning & task: `com/tm/docs/technical/planning/` — mỗi phase có `README.md` (requirement, task, challenge, DoD), `test-cases.md`, `lessons-learned.md`
- Handbook kỹ thuật theo task: `com/tm/docs/technical/handbook/`
- Quy ước code: `com/tm/docs/technical/project-structure.md` (mục "Go cho người từ Java", quy tắc 3 tầng dùng chung)
- Quyết định kiến trúc: `com/tm/docs/technical/adr/`

---

## Quy trình làm một task — BẮT BUỘC, theo đúng thứ tự

Task được gọi bằng ID, ví dụ `P3-T05`. Không bỏ bước, không đảo thứ tự.

### Bước 0 — Validate trước khi làm

Kiểm tra tất cả, thiếu một điều kiện → **dừng**, báo điều kiện nào chưa đạt, không làm task:

1. **Task liền trước** (trong cùng phase, hoặc task cuối của phase trước) đã:
   - `[x]` ở dòng task và đủ `[x]` cả 6 bước trong checklist con;
   - có commit: bước 6 ghi commit hash, và hash đó có trong `git log`.
2. Phase trước đã đóng (✅ trong `planning/README.md`) nếu đây là task đầu của phase.
3. Working tree sạch (`git status` không có thay đổi chưa commit của task khác).
4. Task không mâu thuẫn với docs/ADR; thiếu thông tin → hỏi.

Chỉ bỏ qua điều kiện khi người dùng nói rõ cho phép, và ghi lại lý do vào checklist con.

> **Được enforce bằng hook**: `.claude/settings.json` chạy `python3 .claude/scripts/planning.py guard` trước mọi Edit/Write vào `com/tm/server/**` và `com/tm/app/**`. Hook từ chối nếu task trước chưa đủ checklist + commit, phase trước chưa đóng, task hiện tại chưa có checklist con, hoặc bước 1 (duyệt test case) chưa `[x]`. Không lách hook bằng Bash (`sed`, heredoc...) — hook bị chặn nghĩa là quy trình chưa đúng. Kiểm tra nhanh: `python3 .claude/scripts/planning.py validate`.

Đạt → đọc `README.md` của phase (mục tiêu, requirement, challenge trong `[...]`, DoD), rồi **thêm checklist con** ngay dưới dòng task:

```markdown
- [ ] **P3-T05** <mô tả task>
  - [ ] 1. Test case: <ID các test case> — đã được duyệt
  - [ ] 2. Code
  - [ ] 3. Unit test
  - [ ] 4. Build + unit test pass
  - [ ] 5. Test case pass + handbook
  - [ ] 6. Commit: `<hash>` · Push: có/không
```

Đánh `[x]` từng bước ngay khi bước đó xong. Phase đang ⬜ → 🟨 khi bắt đầu task đầu tiên.

### Bước 1 — Gen test case → chờ duyệt

1. Viết test case cho task vào `test-cases.md` của phase: ID tiếp theo, loại, kịch bản, kết quả mong đợi, map requirement/challenge, trạng thái ⬜. Tận dụng test case đã có nếu phù hợp.
2. Trình bày danh sách test case cho người dùng và **DỪNG**.
3. Chỉ sang bước 2 khi người dùng **duyệt rõ ràng**. Người dùng yêu cầu sửa → sửa rồi xin duyệt lại.
4. Duyệt xong → đánh `[x]` bước 1, ghi ID test case.

Test case đã duyệt là tiêu chí nghiệm thu: muốn thêm/sửa/xoá sau đó phải xin duyệt lại.

### Bước 2 — Code

- Theo `project-structure.md`: package theo nghiệp vụ, interface phía dùng, wiring tay, không DI framework; Fastify plugin, không NestJS.
- Chỉ làm trong phạm vi task. Việc ngoài phạm vi → ghi lại, báo người dùng, không tự làm.
- Thêm/xoá file Go hoặc đổi import → `bazel run //:gazelle`.

### Bước 3 — Agent tự viết unit test

- Unit test cho code vừa viết: logic thuần, nhánh lỗi, biên. Go: table-driven. TS: Vitest / Testing Library.
- Test cần DB/Docker thật → integration test với testcontainers, gắn tag theo `project-structure.md`.

### Bước 4 — Build và chạy lại unit test

Chạy đủ, không chỉ test mới:
- Server: `bazel run //:gazelle`, `go vet ./...`, `go test -race ./...`, `bazel test` cho target bị ảnh hưởng.
- App: `pnpm --filter "...[origin/main]" lint test build`.

Fail → sửa code (không sửa/skip test cho pass) → chạy lại đến khi pass. Báo kết quả thật kèm output tóm tắt.

### Bước 5 — Test theo test case + ghi handbook

1. Thực hiện từng test case đã duyệt ở bước 1 (tự động hoặc theo hướng dẫn manual trong test case). Pass → ⬜ → ✅ trong `test-cases.md`. Có test case fail → quay lại bước 2.
2. Ghi **handbook** cho task theo mục [Handbook](#handbook) bên dưới.
3. Cập nhật challenge (⬜ → 🟨 / ✅ khi đạt tiêu chí "Hoàn thành khi").
4. Code khác thiết kế trong docs (schema, API, luồng) → cập nhật docs.
5. Đánh `[x]` bước 5 **và** `[x]` dòng task.
6. Task cuối của phase và đủ DoD → nhắc người dùng chạy checklist đóng phase và viết `lessons-learned.md`.

### Bước 6 — Hỏi commit và push

1. Tóm tắt thay đổi, đề xuất commit message, hỏi người dùng: **có commit không? có push không?**
2. Chỉ commit/push khi người dùng đồng ý. Không tự push.
3. Commit xong → ghi hash vào bước 6, đánh `[x]`. (Thay đổi ghi hash này gộp vào commit tiếp theo hoặc amend nếu người dùng muốn.)
4. Người dùng chưa muốn commit → để bước 6 `[ ]`. **Task kế tiếp sẽ bị chặn ở bước 0** cho đến khi commit.

---

## Handbook

`com/tm/docs/technical/handbook/` — ghi chú kỹ thuật **theo task**, do agent viết ở bước 5, để người dùng ôn lại kiến thức đã áp dụng.

- Mỗi phase một file: `handbook/phase-<N>.md`. Mỗi task một mục `## <ID> — <tên task>`.
- Mỗi bài học trong task gồm:
  - **Chủ đề** (tag): ví dụ `go/channel`, `go/errgroup`, `go/context`, `pg/lock`, `pg/isolation`, `pg/index`, `node/event-loop`, `node/stream`, `react/useState`, `react/memo`...
  - **Bối cảnh**: vấn đề gặp trong task.
  - **Cách làm & lý do**: đã dùng kỹ thuật gì, vì sao chọn nó, phương án khác bị loại vì sao.
  - **Bẫy / lưu ý**: lỗi dễ mắc, điều bất ngờ gặp phải.
  - **Code**: link tương đối tới file + dải dòng + tên hàm/type, ví dụ `[booking/service.go#L40-L62 — Service.Confirm](../../../server/services/core/internal/booking/service.go#L40-L62)`. Luôn ghi tên symbol để tìm lại được khi dòng thay đổi.
  - **Tham khảo** (nếu có): tài liệu chính thức, bài viết.
- Chỉ ghi điều thực sự đã làm và học trong task, cụ thể, ngắn gọn. Không chép lại lý thuyết chung chung.
- Thêm dòng vào bảng chỉ mục trong `handbook/README.md` (task, chủ đề, link).

`lessons-learned.md` của phase là phần **người dùng tự viết** — agent không viết thay, chỉ gợi ý khi được yêu cầu.

---

## Không được

- Bỏ qua bước 0 hoặc làm task khi task trước chưa đủ checklist + commit mà người dùng chưa cho phép.
- Code trước khi test case được duyệt.
- Đánh `[x]` / ✅ khi chưa chạy hoặc còn test fail.
- Sửa, xoá, `skip` test để cho pass; sửa test case đã duyệt mà không xin duyệt lại.
- Commit hoặc push khi người dùng chưa đồng ý.
- Viết Go theo kiểu Java (`service/`, `repository/`, `IFoo`/`FooImpl`, DI framework, `utils/`).
- Dùng số thực cho tiền ở bất kỳ tầng nào.

## Commit

Conventional Commits, ghi ID task và challenge: `feat(core): hold seats with conditional update [P4-T03][P1]`.
