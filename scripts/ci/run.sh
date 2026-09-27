#!/usr/bin/env bash
# Các bước CI — workflow GitHub chỉ gọi script này, nên chạy local cho kết quả y hệt CI.
#
#   scripts/ci/run.sh server [BASE]   # Go: gazelle diff → lint → build → test bị ảnh hưởng
#
# BASE: commit so sánh để chọn test bị ảnh hưởng (CI: base của PR / commit trước khi push).
# Không có BASE → chạy mọi test.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SERVER="$ROOT/com/tm/server"
GOLANGCI_LINT_VERSION="v2.6.2"

# Trên GitHub Actions: gom log từng bước thành nhóm thu gọn được, lỗi hiện thành annotation.
step() {
  if [[ -n "${GITHUB_ACTIONS:-}" ]]; then echo "::endgroup::"; echo "::group::$1"; else echo; echo "==> $1"; fi
}
fail() {
  if [[ -n "${GITHUB_ACTIONS:-}" ]]; then echo "::error::$1"; else echo "LỖI: $1" >&2; fi
  exit 1
}

server() {
  local base="${1:-}"
  cd "$SERVER"
  [[ -n "${GITHUB_ACTIONS:-}" ]] && echo "::group::start"

  step "gazelle: BUILD.bazel phải khớp code Go"
  local diff
  diff="$(bazel run --noshow_progress //:gazelle -- -mode=diff 2>/dev/null || true)"
  if [[ -n "$diff" ]]; then
    echo "$diff"
    fail "BUILD.bazel chưa cập nhật — chạy: make gazelle (hoặc bazel run //:gazelle trong com/tm/server)"
  fi

  step "golangci-lint $GOLANGCI_LINT_VERSION"
  go run "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$GOLANGCI_LINT_VERSION" run ./... ||
    fail "golangci-lint báo lỗi"

  step "bazel build //..."
  bazel build //... || fail "bazel build lỗi"

  step "bazel test: target bị ảnh hưởng (BASE=${base:-<tất cả>})"
  local targets=()
  while IFS= read -r t; do
    [[ -n "$t" ]] && targets+=("$t")
  done < <("$ROOT/scripts/ci/bazel-affected-tests.sh" "$base")
  if [[ ${#targets[@]} -eq 0 ]]; then
    echo "Không có test nào bị ảnh hưởng."
  else
    printf '  %s\n' "${targets[@]}"
    local test_flags=()
    # Trên CI: integration test không được skip vì thiếu Docker (xem pgtest.Run).
    [[ -n "${GITHUB_ACTIONS:-}" ]] && test_flags+=(--test_env=PGTEST_REQUIRE_DOCKER=1)
    bazel test "${test_flags[@]+"${test_flags[@]}"}" "${targets[@]}" || fail "bazel test lỗi"
  fi

  [[ -n "${GITHUB_ACTIONS:-}" ]] && echo "::endgroup::"
  echo "server: OK"
}

case "${1:-}" in
  server) server "${2:-}" ;;
  *)
    echo "Dùng: $0 server [BASE]" >&2
    exit 2
    ;;
esac
