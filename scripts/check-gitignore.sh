#!/usr/bin/env bash
# Kiểm tra .gitignore bỏ qua đúng file build/bí mật và giữ đúng file cần commit (P0-T01).
#
#   scripts/check-gitignore.sh
#
# Exit 0 nếu mọi đường dẫn mẫu cho kết quả đúng; exit 1 và liệt kê đường dẫn sai.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

MUST_IGNORE=(
  com/tm/server/bazel-out/x
  com/tm/server/bazel-bin/x
  com/tm/app/node_modules/x
  com/tm/app/apps/bff/dist/x
  com/tm/app/apps/web-client/dist/x
  .env
  com/tm/server/coverage.out
)

MUST_KEEP=(
  com/tm/server/go.mod
  com/tm/server/go.sum
  .env.example
  com/tm/app/pnpm-lock.yaml
  .claude/settings.json
)

wrong=()
for p in "${MUST_IGNORE[@]}"; do
  git check-ignore -q --no-index "$p" || wrong+=("phải bị ignore: $p")
done
for p in "${MUST_KEEP[@]}"; do
  if git check-ignore -q --no-index "$p"; then wrong+=("không được ignore: $p"); fi
done

if ((${#wrong[@]})); then
  echo "Sai ${#wrong[@]} đường dẫn:" >&2
  printf '  - %s\n' "${wrong[@]}" >&2
  exit 1
fi

echo "OK: ${#MUST_IGNORE[@]} đường dẫn bị ignore, ${#MUST_KEEP[@]} đường dẫn được giữ."
