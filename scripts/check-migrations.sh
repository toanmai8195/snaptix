#!/usr/bin/env bash
# Kiểm tra quy ước file migration goose.
#
#   scripts/check-migrations.sh [DIR...]   # mặc định: com/tm/server/db/{core,analytics}/migrations
#
# Quy ước: tên NNNNN_ten_snake.sql, số thứ tự không trùng, mỗi file có -- +goose Up và -- +goose Down.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
if [[ $# -eq 0 ]]; then
  set -- "$ROOT/com/tm/server/db/core/migrations" "$ROOT/com/tm/server/db/analytics/migrations"
fi

errors=()
for dir in "$@"; do
  [[ -d "$dir" ]] || { errors+=("không có thư mục: $dir"); continue; }
  seen=""
  for f in "$dir"/*; do
    [[ -e "$f" ]] || continue
    name="$(basename "$f")"
    [[ "$name" == .gitkeep ]] && continue
    if [[ ! "$name" =~ ^([0-9]{5})_[a-z0-9_]+\.sql$ ]]; then
      errors+=("$dir: tên sai quy ước NNNNN_ten.sql: $name"); continue
    fi
    num="${BASH_REMATCH[1]}"
    if [[ " $seen " == *" $num "* ]]; then errors+=("$dir: trùng số thứ tự $num: $name"); fi
    seen="$seen $num"
    grep -q -- '^-- +goose Up' "$f"   || errors+=("$dir: thiếu '-- +goose Up': $name")
    grep -q -- '^-- +goose Down' "$f" || errors+=("$dir: thiếu '-- +goose Down': $name")
  done
done

if ((${#errors[@]})); then
  echo "Migration sai quy ước (${#errors[@]}):" >&2
  printf '  - %s\n' "${errors[@]}" >&2
  exit 1
fi
echo "OK: migration đúng quy ước."
