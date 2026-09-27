# Phase 0 — Nền móng (tối giản cho chặng A)

> Chỉ dựng những gì chặng A (Go + PostgreSQL) cần. MongoDB, Redis, BFF, observability đầy đủ được thêm ở phase dùng tới chúng.

## Mục tiêu

Có repo, Bazel + Go, PostgreSQL local, migration, một core service "rỗng" chạy được (health, log, dừng êm, image), CI cho Go.

**Mốc demo**: `make up && make migrate` → `go run` core → `curl /readyz` trả 200; CI xanh.

## Kiến thức trọng tâm

Go: module, package `main`, `net/http`, chi, `context`, `slog`, `signal` · PG: chạy bằng Docker, migration với goose · Công cụ: Bazel + Gazelle, Docker Compose, Makefile, GitHub Actions

## Phạm vi

- **Trong**: cấu trúc repo, Docker Compose (chỉ PG core), goose, Bazel + Go, skeleton core, graceful shutdown, image OCI, testcontainers-go, Makefile, CI Go.
- **Ngoài**: logic nghiệp vụ; MongoDB, Redis, BFF, web (chặng B, C); OpenTelemetry/Grafana (Phase 6); PG analytics (Phase 6).

## Workstream

`infra` · `db` · `core` · `qa`

## Requirement

| ID | Loại | Mô tả |
|---|---|---|
| P0-FR1 | FR | Một lệnh khởi động hạ tầng local (PG core) |
| P0-FR2 | FR | Migration chạy được cho PG core |
| P0-FR3 | FR | `core` có `/healthz` (sống) và `/readyz` (kết nối được DB) |
| P0-FR4 | FR | CI chạy lint, build, test Go trên mỗi PR |
| P0-NFR1 | NFR | Log dạng JSON một dòng |
| P0-NFR2 | NFR | Service dừng an toàn khi nhận SIGTERM, không cắt ngang request |

## Task

### infra
- [x] **P0-T01** Cấu trúc thư mục `com/tm/server`, `.gitignore` cho Go/Bazel
  - [x] 1. Test case: P0-T01-TC01..TC04 + kế hoạch subtask — đã được duyệt
  - [x] 2. Code
    - [x] 2.1 Tạo `com/tm/server/README.md`
    - [x] 2.2 Thêm pattern Go/Bazel/editor vào `.gitignore`
  - [x] 3. Unit test
  - [x] 4. Build + unit test pass
  - [x] 5. Test case pass + handbook
  - [x] 6. Commit: `chore(server): add com/tm/server and Go/Bazel gitignore [P0-T01]` · Push: có
- [x] **P0-T02** `deploy/docker-compose.yml` chỉ có PostgreSQL core (5432), healthcheck, volume
  - [x] 1. Test case: P0-T02-TC01..TC04 + kế hoạch subtask — đã được duyệt
  - [x] 2. Code
    - [x] 2.1 Compose tối thiểu: service `postgres-core`, env, cổng 5432
    - [x] 2.2 Named volume `pg-core`
    - [x] 2.3 Healthcheck `pg_isready`, tên project, `up --wait`
  - [x] 3. Unit test
  - [x] 4. Build + unit test pass
  - [x] 5. Test case pass + handbook
  - [x] 6. Commit: `build(deploy): docker-compose with PostgreSQL core [P0-T02]` · Push: không

### db
- [x] **P0-T03** goose migration cho PG core: `scripts/migrate.sh`, migration rỗng đầu tiên
  - [x] 1. Test case: P0-T03-TC01..TC05 + kế hoạch subtask — đã được duyệt
  - [x] 2. Code
    - [x] 2.1 Chạy goose bằng `go run ...@v3.26.0`
    - [x] 2.2 Migration `00001_init.sql`, chạy up/status/down thủ công
    - [x] 2.3 Gói vào `scripts/migrate.sh`
  - [x] 3. Unit test
  - [x] 4. Build + unit test pass
  - [x] 5. Test case pass + handbook
  - [x] 6. Commit: `build(db): goose migrations for PG core [P0-T03]` · Push: không

### core
- [x] **P0-T04** Bazel + Gazelle + một `go.mod` cho `com/tm/server` `[G14]`
  - [x] 1. Test case: P0-T04-TC01..TC05 + kế hoạch subtask — đã được duyệt
  - [x] 2. Code
    - [x] 2.1 `go mod init`, thử chương trình tạm
    - [x] 2.2 Bazel tối thiểu: `.bazelversion`, `MODULE.bazel`, `.bazelrc`
    - [x] 2.3 Target `//:gazelle` + directive; thử với `pkg/probe`
    - [x] 2.4 Luồng thêm thư viện ngoài (`uuid`)
  - [x] 3. Unit test
  - [x] 4. Build + unit test pass
  - [x] 5. Test case pass + handbook
  - [x] 6. Commit: `build(server): Go module, Bazel and Gazelle for com/tm/server [P0-T04][G14]` · Push: không
- [x] **P0-T05** Skeleton core: hello world → HTTP server → chi router + `/healthz` → config từ env → log `slog` JSON → pgxpool + `/readyz`
  - [x] 1. Test case: P0-T05-TC01..TC11 + kế hoạch subtask — đã được duyệt
  - [x] 2. Code
    - [x] 2.1 Hello world `services/core/cmd/server`
    - [x] 2.2 HTTP server bằng `net/http` + `/healthz`
    - [x] 2.3 chi router, package `internal/httpx`
    - [x] 2.4 Config từ env (`loadConfig`)
    - [x] 2.5 Log `slog` JSON + middleware log request
    - [x] 2.6 pgxpool + `/readyz`
  - [x] 3. Unit test
  - [x] 4. Build + unit test pass
  - [x] 5. Test case pass + handbook
  - [x] 6. Commit: `feat(core): skeleton core server with chi, slog, pgxpool and health endpoints [P0-T05]` · Push: có
- [ ] **P0-T06** Graceful shutdown: SIGTERM, `http.Server.Shutdown` có timeout, đóng pool sau cùng `[G3]`
- [ ] **P0-T07** Image OCI cho core bằng macro `com_tm_go_image` (rules_oci, distroless, non-root) `[G14]`

### qa
- [ ] **P0-T08** Khung testcontainers-go cho integration test với PG

### infra
- [ ] **P0-T09** Makefile: `up`, `down`, `migrate`, `test`, `gazelle`
- [ ] **P0-T10** GitHub Actions cho Go: gazelle diff, golangci-lint, `bazel test` target bị ảnh hưởng `[G14]`

## Challenge

| # | Công nghệ | Challenge | Bối cảnh | Hướng giải | Hoàn thành khi | Trạng thái |
|---|---|---|---|---|---|---|
| G3 | Golang | Graceful shutdown | Deploy khi đang có giao dịch | Bắt SIGTERM, ngừng nhận request, chờ in-flight, đóng pool theo thứ tự | Rolling deploy dưới tải không mất/không lỗi request | ⬜ |
| G14 | Golang | Monorepo Go với Bazel | `com/tm/server` nhiều service + thư viện | rules_go + Gazelle + bzlmod, một `go.mod`, visibility, test theo target bị ảnh hưởng | Code build được bằng cả `go` và Bazel; CI chỉ test target bị ảnh hưởng; core build được thành image | 🟨 |

## Definition of Done

- [ ] Clone repo mới → `make up && make migrate` chạy thành công trong < 5 phút
- [ ] Core chạy được bằng `go run` và bằng image Docker; `/readyz` phản ánh đúng trạng thái PG
- [ ] Gửi SIGTERM khi đang có request chậm → request hoàn thành, service thoát sạch
- [ ] CI xanh trên `main`
- [ ] Mọi test nghiệm thu trong [acceptance-tests.md](acceptance-tests.md) và test case của các task trong [tasks/](tasks/) pass

## Checklist đóng phase

- [ ] Cập nhật [local-setup](../../local-setup.md) theo thực tế
- [ ] ADR: Bazel cho Go + pnpm cho TS; công cụ migration
- [ ] Viết [lessons-learned.md](lessons-learned.md)
- [ ] Cập nhật trạng thái phase trong [planning](../README.md)
