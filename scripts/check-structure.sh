#!/usr/bin/env bash
# Kiểm tra cấu trúc thư mục khung của monorepo (P0-T01).
# Nguồn: com/tm/docs/technical/project-structure.md
#
#   scripts/check-structure.sh [ROOT]   # ROOT mặc định: thư mục gốc repo
#
# Exit 0 nếu đủ; exit 1 và in các thư mục thiếu nếu không đủ.
set -euo pipefail

ROOT="${1:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"

REQUIRED_DIRS=(
  com/tm/server/api
  com/tm/server/db/core/migrations
  com/tm/server/db/analytics/migrations
  com/tm/server/pkg
  com/tm/server/services/core
  com/tm/server/services/stats-worker
  com/tm/app/api
  com/tm/app/apps/bff
  com/tm/app/apps/web-client
  com/tm/app/apps/web-admin
  com/tm/app/packages/types
  com/tm/app/packages/api-client
  com/tm/app/packages/ui
  com/tm/app/packages/config
  com/tm/docs
  deploy
  loadtest
)

missing=()
for d in "${REQUIRED_DIRS[@]}"; do
  [[ -d "$ROOT/$d" ]] || missing+=("$d")
done

if ((${#missing[@]})); then
  echo "Thiếu ${#missing[@]} thư mục trong $ROOT:" >&2
  printf '  - %s\n' "${missing[@]}" >&2
  exit 1
fi

echo "OK: đủ ${#REQUIRED_DIRS[@]} thư mục khung."
