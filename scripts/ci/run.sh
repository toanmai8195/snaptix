#!/usr/bin/env bash
# Các bước CI, chạy được y hệt ở local.
#
#   scripts/ci/run.sh repo               # kiểm tra repo (luôn chạy)
#   scripts/ci/run.sh server [BASE_SHA]  # Go/Bazel; có BASE thì chỉ test target bị ảnh hưởng
#   scripts/ci/run.sh app [BASE_SHA]     # Node/React; có BASE thì chỉ package bị ảnh hưởng
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
GOLANGCI_LINT_VERSION="v2.6.2"

step() { echo; echo "::group::$*"; }
end() { echo "::endgroup::"; }

valid_base() { [[ -n "${1:-}" && ! "$1" =~ ^0+$ ]] && git -C "$REPO" cat-file -e "$1^{commit}" 2>/dev/null; }

repo() {
  cd "$REPO"
  step "Cấu trúc repo";      scripts/check-structure.sh; end
  step ".gitignore";         scripts/check-gitignore.sh; end
  step "Quy ước migration";  scripts/check-migrations.sh; end
  step "Test script repo";   scripts/test/check-scripts.test.sh; end
}

server() {
  local base="${1:-}"
  cd "$REPO/com/tm/server"

  step "Gazelle: BUILD.bazel phải cập nhật"
  if ! bazel run //:gazelle -- -mode=diff; then
    echo "BUILD.bazel lỗi thời — chạy: bazel run //:gazelle" >&2; exit 1
  fi
  end

  step "golangci-lint $GOLANGCI_LINT_VERSION"
  if [[ -n "$(GOTOOLCHAIN=local go list ./... 2>/dev/null)" ]]; then
    GOTOOLCHAIN=local go run "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$GOLANGCI_LINT_VERSION" run ./...
  else
    echo "Chưa có package Go — bỏ qua"
  fi
  end

  step "bazel build //..."; bazel build //...; end

  step "bazel test (target bị ảnh hưởng)"
  local targets
  if valid_base "$base"; then
    targets="$(git -C "$REPO" diff --name-only "$base"...HEAD | "$REPO/scripts/ci/bazel-affected-tests.sh")"
  else
    targets="//..."
  fi
  if [[ -z "$targets" ]]; then
    echo "Không có test bị ảnh hưởng"
  else
    echo "$targets"
    # exit 4 = không có test target nào để chạy → không phải lỗi
    set +e; bazel test $targets; code=$?; set -e
    if [[ $code -ne 0 && $code -ne 4 ]]; then exit "$code"; fi
  fi
  end
}

app() {
  local base="${1:-}"
  cd "$REPO/com/tm/app"
  step "pnpm install --frozen-lockfile"; pnpm install --frozen-lockfile; end
  local filter=()
  if valid_base "$base"; then filter=(--filter "...[$base]"); fi
  for s in lint test build; do
    step "pnpm $s ${filter[*]:-}"
    if ((${#filter[@]})); then
      pnpm --recursive --if-present "${filter[@]}" "$s"
    else
      pnpm "$s"
    fi
    end
  done
}

case "${1:-}" in
  repo) repo ;;
  server) server "${2:-}" ;;
  app) app "${2:-}" ;;
  *) echo "Dùng: $0 repo | server [BASE] | app [BASE]" >&2; exit 2 ;;
esac
