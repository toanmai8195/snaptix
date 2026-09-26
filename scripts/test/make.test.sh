#!/usr/bin/env bash
# Kiểm thử Makefile gốc (P0-T05).
#
#   scripts/test/make.test.sh
#
# Dùng compose project "snaptix-test" — cổng hạ tầng phải trống (dừng stack dev trước).
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO"
PROJ=snaptix-test
TMP="$(mktemp -d)"

pass=0 fail=0
check() {
  if [[ "$2" == 0 ]]; then pass=$((pass + 1)); echo "  ✅ $1"
  else fail=$((fail + 1)); echo "  ❌ $1"; [[ -n "${3:-}" ]] && echo "$3" | tail -15 | sed 's/^/     /'; fi
}

for p in 5432 5433 27017 6379 9090 3100; do
  if lsof -nP -iTCP:"$p" -sTCP:LISTEN >/dev/null 2>&1; then
    echo "Cổng $p đang bận — dừng stack dev trước: make down" >&2; exit 2
  fi
done
cleanup() {
  make -s nuke PROJECT=$PROJ >/dev/null 2>&1
  [[ -d "$TMP/clone" ]] && (cd "$TMP/clone" && make -s nuke PROJECT=$PROJ >/dev/null 2>&1)
  rm -rf "$REPO/com/tm/server/pkg/mk_fail" "$TMP"
}
trap cleanup EXIT

db_version() { docker compose -p $PROJ -f deploy/docker-compose.yml exec -T "postgres-$1" psql -U snaptix -d "$1" -tAc \
  "SELECT COALESCE(MAX(version_id),-1) FROM goose_db_version WHERE is_applied" 2>/dev/null; }

# ---------------------------------------------------------------------------
echo "TC01 — make / make help"
out="$(make 2>&1)"; check "make (mặc định) exit 0" "$?" "$out"
for t in up down nuke migrate test lint build gazelle; do
  check "help có target '$t' kèm mô tả" "$(grep -qE "^  .*$t +.*[[:alpha:]]" <<<"$(sed 's/\x1b\[[0-9;]*m//g' <<<"$out")"; echo $?)" "$out"
done
check "make mặc định không chạy docker" "$(! grep -q 'docker compose' <<<"$out"; echo $?)"
check "make = make help" "$([[ "$out" == "$(make help 2>&1)" ]]; echo $?)"

# ---------------------------------------------------------------------------
echo "TC02 — up / ps / down / nuke"
out="$(make up PROJECT=$PROJ 2>&1)"; check "make up (chờ healthy)" "$?" "$out"
out="$(make ps PROJECT=$PROJ 2>&1)"
check "make ps liệt kê postgres-core, grafana" "$(grep -q postgres-core <<<"$out" && grep -q grafana <<<"$out"; echo $?)" "$out"
docker compose -p $PROJ -f deploy/docker-compose.yml exec -T postgres-core psql -U snaptix -d core -c "CREATE TABLE mk_keep(i int)" >/dev/null 2>&1
out="$(make down PROJECT=$PROJ 2>&1)"; check "make down" "$?" "$out"
check "down giữ volume" "$([[ -n "$(docker volume ls -q --filter label=com.docker.compose.project=$PROJ)" ]]; echo $?)"
make up PROJECT=$PROJ >/dev/null 2>&1
check "dữ liệu còn sau down/up" "$(docker compose -p $PROJ -f deploy/docker-compose.yml exec -T postgres-core psql -U snaptix -d core -tAc "SELECT to_regclass('mk_keep') IS NOT NULL" | grep -qx t; echo $?)"

# ---------------------------------------------------------------------------
echo "TC03 — migrate / migrate-status"
docker compose -p $PROJ -f deploy/docker-compose.yml exec -T postgres-core psql -U snaptix -d core -c "DROP TABLE mk_keep" >/dev/null 2>&1
out="$(make migrate 2>&1)"; check "make migrate" "$?" "$out"
check "core version = 1" "$([[ "$(db_version core)" == 1 ]]; echo $?)"
check "analytics version = 1" "$([[ "$(db_version analytics)" == 1 ]]; echo $?)"
out="$(make migrate-status 2>&1)"; check "make migrate-status" "$?" "$out"
out="$(make nuke PROJECT=$PROJ 2>&1)"; check "make nuke" "$?" "$out"
check "nuke xoá volume" "$([[ -z "$(docker volume ls -q --filter label=com.docker.compose.project=$PROJ)" ]]; echo $?)"

# ---------------------------------------------------------------------------
echo "TC04 — make test"
out="$(make test 2>&1)"; check "make test exit 0" "$?" "$out"
check "đã chạy repo, server, app" "$(grep -q 'run.sh repo' <<<"$out" && grep -q 'run.sh server' <<<"$out" && grep -q 'run.sh app' <<<"$out"; echo $?)"
d="$REPO/com/tm/server/pkg/mk_fail"; mkdir -p "$d"
printf 'package mk_fail\n\n// V kiểm thử.\nfunc V() int { return 1 }\n' > "$d/v.go"
printf 'package mk_fail\n\nimport "testing"\n\nfunc TestFail(t *testing.T) {\n\tif V() != 2 {\n\t\tt.Fatal("cố ý fail")\n\t}\n}\n' > "$d/v_test.go"
(cd com/tm/server && bazel run //:gazelle >/dev/null 2>&1)
out="$(make test 2>&1)"; code=$?
check "test Go fail → make test exit khác 0" "$([[ $code -ne 0 ]] && grep -q 'mk_fail' <<<"$out"; echo $?)" "$out"
rm -rf "$d"

# ---------------------------------------------------------------------------
echo "TC05 — P0-AT01: clone mới → make up && make migrate < 5 phút"
if [[ -z "$(git status --porcelain)" ]]; then
  git clone -q "$REPO" "$TMP/clone"; how="git clone"
else
  # Tree chưa commit: dựng bản giống clone từ index tạm (tôn trọng .gitignore)
  GIT_INDEX_FILE="$TMP/index" git read-tree HEAD
  GIT_INDEX_FILE="$TMP/index" git add -A
  GIT_INDEX_FILE="$TMP/index" git checkout-index -a --prefix="$TMP/clone/"
  how="snapshot index (tree chưa commit)"
fi
echo "     ($how)"
start=$(date +%s)
out="$(cd "$TMP/clone" && chmod +x scripts/*.sh scripts/ci/*.sh 2>/dev/null; make up PROJECT=$PROJ 2>&1 && make migrate 2>&1)"; code=$?
secs=$(( $(date +%s) - start ))
check "make up && make migrate thành công" "$code" "$out"
check "mất ${secs}s < 300s" "$([[ $secs -lt 300 ]]; echo $?)"
check "core version = 1" "$([[ "$(db_version core)" == 1 ]]; echo $?)"
check "analytics version = 1" "$([[ "$(db_version analytics)" == 1 ]]; echo $?)"

echo
echo "Kết quả: $pass pass, $fail fail"
((fail == 0))
