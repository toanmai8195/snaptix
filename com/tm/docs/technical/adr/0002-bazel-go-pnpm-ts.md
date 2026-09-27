# ADR-0002: Bazel cho Go, pnpm cho TypeScript — hai hệ build tách biệt

- **Trạng thái**: Chấp nhận
- **Ngày**: 2026-09-27

## Bối cảnh

- Monorepo có hai thế giới: `com/tm/server` (Go: core, stats-worker, thư viện) và `com/tm/app` (Node BFF + React). Hai bên nối với nhau qua hợp đồng OpenAPI.
- Cần: build tái lập giữa máy dev và CI; CI chỉ test phần bị ảnh hưởng khi repo lớn dần (challenge G14); build image OCI cho Go không cần Docker daemon.
- Người phát triển học Go/Node lần đầu — công cụ không được che mất công cụ chuẩn của từng ngôn ngữ (`go test`, `pnpm test` vẫn phải chạy được).

## Các phương án

1. **Chỉ công cụ chuẩn** (`go` + Makefile, pnpm) — Ưu: đơn giản, ít thứ phải học. Nhược: không có đồ thị phụ thuộc để chọn test bị ảnh hưởng; build image cần Dockerfile + daemon; cache CI tự làm.
2. **Bazel cho cả Go lẫn TS** (rules_go + rules_js/ts) — Ưu: một đồ thị, một cách build. Nhược: rules_js/ts phức tạp, lệch khỏi luồng Vite/tsx quen thuộc của hệ sinh thái Node; chi phí học rất lớn cho phần không phải trọng tâm.
3. **Bazel cho Go, pnpm workspace cho TS** — Ưu: Go có đồ thị phụ thuộc (`rdeps` → test bị ảnh hưởng), build hermetic + image bằng rules_oci; TS giữ công cụ chuẩn (pnpm, Vite, Vitest, `pnpm --filter "...[BASE]"` đã có chọn gói bị ảnh hưởng). Nhược: hai hệ build, hai bộ cache CI.

## Quyết định

Chọn **phương án 3**.

Go (đã dựng ở Phase 0):
- Một `go.mod` cho toàn bộ `com/tm/server`; `go.mod` là nguồn sự thật cho dependency, `MODULE.bazel` chỉ `use_repo` (bzlmod + `go_deps.from_file`).
- Gazelle sinh `BUILD.bazel` (`go_naming_convention import`, `map_kind go_binary com_tm_go_image`); code phải build được bằng cả `go` lẫn Bazel — gopls/IDE dùng `go.mod` trực tiếp.
- Bazel 8.7.0 (rules_oci chưa hỗ trợ Bazel 9); Go SDK do Bazel tự tải (`go_sdk.download`).
- Image: macro `com_tm_go_image` (rules_oci, distroless static nonroot, pin digest) — theo cấu trúc project thor.
- CI: `scripts/ci/run.sh server` — gazelle diff, golangci-lint, `bazel build //...`, `bazel test` các target từ `rdeps` của file thay đổi.

TS (từ chặng C): pnpm workspace, CI `pnpm --filter "...[BASE]" lint test build`; image BFF bằng Dockerfile multi-stage.

## Hệ quả

- Dễ hơn: CI Go chỉ test target bị ảnh hưởng (đo được: sửa `httpx/health.go` → 3/4 test target, sửa docs → 0); image Go tái lập (digest giống hệt sau `bazel clean`); build chéo linux/arm64 và amd64 từ macOS.
- Khó hơn: phải nhớ `bazel run //:gazelle` sau khi thêm file/import (CI chặn nếu quên); `bazel mod tidy` sau khi thêm dependency; hai hệ build, hai cách cache.
- Theo dõi: thời gian job `server` trên CI (lần đầu ~3 phút với cache trống); khi rules_oci hỗ trợ Bazel 9 thì nâng cấp.
