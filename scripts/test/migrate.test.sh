#!/usr/bin/env bash
# Kiểm thử scripts/migrate.sh và scripts/check-migrations.sh (P0-T03).
#
#   scripts/test/migrate.test.sh
#
# Bật PG core + analytics bằng compose project tạm "snaptix-test" (cổng 5432/5433 phải trống).
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO"
C=(docker compose -p snaptix-test -f deploy/docker-compose.yml)
TMP="$(mktemp -d)"

pass=0 fail=0
check() {
  if [[ "$2" == 0 ]]; then pass=$((pass + 1)); echo "  ✅ $1"
  else fail=$((fail + 1)); echo "  ❌ $1"; [[ -n "${3:-}" ]] && echo "$3" | tail -15 | sed 's/^/     /'; fi
}

for p in 5432 5433; do
  if lsof -nP -iTCP:"$p" -sTCP:LISTEN >/dev/null 2>&1; then
    echo "Cổng $p đang bận — dừng stack dev trước: docker compose -f deploy/docker-compose.yml down" >&2
    exit 2
  fi
done
trap '"${C[@]}" down -v --remove-orphans >/dev/null 2>&1; rm -rf "$TMP"' EXIT
"${C[@]}" up -d --wait postgres-core postgres-analytics >/dev/null 2>&1 || { echo "Không bật được PG" >&2; exit 2; }

sql() { # sql <core|analytics> <query>
  "${C[@]}" exec -T "postgres-$1" psql -U snaptix -d "$1" -tAc "$2"
}
db_version() { sql "$1" "SELECT COALESCE(MAX(version_id), -1) FROM goose_db_version WHERE is_applied" 2>/dev/null; }
has_version_table() { [[ "$(sql "$1" "SELECT to_regclass('public.goose_db_version') IS NOT NULL")" == t ]]; }

# ---------------------------------------------------------------------------
echo "TC01 — phiên bản goose đã pin"
pinned="$(sed -n 's/^GOOSE_VERSION="\(.*\)"/\1/p' scripts/migrate.sh)"
out="$(scripts/migrate.sh --tool-version 2>&1)"
check "in 'goose version: $pinned'" "$(grep -q "goose version: $pinned" <<<"$out"; echo $?)" "$out"
check "không tự tải toolchain Go khác" "$(! grep -q 'switching to go' <<<"$out"; echo $?)" "$out"

# ---------------------------------------------------------------------------
echo "TC02 — core up + status"
out="$(scripts/migrate.sh core up 2>&1)"; check "core up" "$?" "$out"
check "core version = 1" "$([[ "$(db_version core)" == 1 ]]; echo $?)"
out="$(scripts/migrate.sh core status 2>&1)"
check "status: 00001_init.sql Applied" "$(grep -E 'Applied At|[0-9]' <<<"$out" | grep -q '00001_init.sql' && ! grep -q 'Pending.*00001_init' <<<"$out"; echo $?)" "$out"

# ---------------------------------------------------------------------------
echo "TC03 — analytics độc lập"
check "analytics chưa có bảng version trước khi migrate" "$(! has_version_table analytics; echo $?)"
out="$(scripts/migrate.sh analytics up 2>&1)"; check "analytics up" "$?" "$out"
check "analytics version = 1" "$([[ "$(db_version analytics)" == 1 ]]; echo $?)"
check "bảng goose_db_version nằm ở cả 2 DB riêng" "$(has_version_table core && has_version_table analytics; echo $?)"

# ---------------------------------------------------------------------------
echo "TC04 — up idempotent, down, up lại"
out="$(scripts/migrate.sh core up 2>&1)"; check "up lần 2 exit 0" "$?" "$out"
check "vẫn version 1" "$([[ "$(db_version core)" == 1 ]]; echo $?)"
out="$(scripts/migrate.sh core down 2>&1)"; check "down" "$?" "$out"
check "version về 0" "$([[ "$(db_version core)" == 0 ]]; echo $?)"
check "analytics không bị ảnh hưởng (vẫn 1)" "$([[ "$(db_version analytics)" == 1 ]]; echo $?)"
out="$(scripts/migrate.sh core up 2>&1)"; check "up lại" "$?" "$out"
check "version lại = 1" "$([[ "$(db_version core)" == 1 ]]; echo $?)"

# ---------------------------------------------------------------------------
echo "TC05 — quy ước migration"
out="$(scripts/check-migrations.sh 2>&1)"; check "migration hiện có đúng quy ước" "$?" "$out"
mkcase() { mkdir -p "$TMP/$1"; printf -- '-- +goose Up\nSELECT 1;\n-- +goose Down\nSELECT 1;\n' > "$TMP/$1/00001_init.sql"; }
mkcase badname && printf -- '-- +goose Up\n-- +goose Down\n' > "$TMP/badname/2_Add-Trips.sql"
out="$(scripts/check-migrations.sh "$TMP/badname" 2>&1)"
check "fail khi tên sai quy ước" "$([[ $? -ne 0 ]] && grep -q '2_Add-Trips.sql' <<<"$out"; echo $?)" "$out"
mkcase dup && cp "$TMP/dup/00001_init.sql" "$TMP/dup/00001_again.sql"
out="$(scripts/check-migrations.sh "$TMP/dup" 2>&1)"
check "fail khi trùng số thứ tự" "$([[ $? -ne 0 ]] && grep -q 'trùng số thứ tự 00001' <<<"$out"; echo $?)" "$out"
mkcase nodown && printf -- '-- +goose Up\nSELECT 1;\n' > "$TMP/nodown/00002_no_down.sql"
out="$(scripts/check-migrations.sh "$TMP/nodown" 2>&1)"
check "fail khi thiếu -- +goose Down" "$([[ $? -ne 0 ]] && grep -q "thiếu '-- +goose Down': 00002_no_down.sql" <<<"$out"; echo $?)" "$out"
out="$(cd "$TMP" && CORE_DATABASE_URL=x "$REPO/scripts/migrate.sh" core create add_trips sql 2>&1; ls "$REPO/com/tm/server/db/core/migrations")"
created="$(ls com/tm/server/db/core/migrations | grep add_trips || true)"
check "create đánh số tuần tự (00002_add_trips.sql)" "$([[ "$created" == 00002_add_trips.sql ]]; echo $?)" "$out"
rm -f com/tm/server/db/core/migrations/*add_trips*

echo
echo "Kết quả: $pass pass, $fail fail"
((fail == 0))
