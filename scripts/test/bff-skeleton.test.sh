#!/usr/bin/env bash
# Kiểm thử tích hợp skeleton BFF (P0-T11).
#
#   scripts/test/bff-skeleton.test.sh
#
# MongoDB bằng compose project "snaptix-test"; BFF chạy ở cổng 3000.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
APP="$REPO/com/tm/app"
C=(docker compose -p snaptix-test -f "$REPO/deploy/docker-compose.yml")
TMP="$(mktemp -d)"
PID=""

pass=0 fail=0
check() {
  if [[ "$2" == 0 ]]; then pass=$((pass + 1)); echo "  ✅ $1"
  else fail=$((fail + 1)); echo "  ❌ $1"; [[ -n "${3:-}" ]] && echo "$3" | tail -15 | sed 's/^/     /'; fi
}
retry() { local n="$1"; shift; for _ in $(seq "$n"); do "$@" >/dev/null 2>&1 && return 0; sleep 0.5; done; return 1; }
stop_bff() { [[ -n "$PID" ]] && { pkill -TERM -P "$PID" 2>/dev/null; kill -TERM "$PID" 2>/dev/null; wait "$PID" 2>/dev/null; }; PID=""; }

for p in 27017 3000; do
  if lsof -nP -iTCP:"$p" -sTCP:LISTEN >/dev/null 2>&1; then echo "Cổng $p đang bận" >&2; exit 2; fi
done
trap 'stop_bff; "${C[@]}" down -v --remove-orphans >/dev/null 2>&1; rm -rf "$TMP"' EXIT
"${C[@]}" up -d --wait mongodb >/dev/null 2>&1 || { echo "Không bật được MongoDB" >&2; exit 2; }
code_of() { curl -s -o "$TMP/body" -w '%{http_code}' --max-time 5 "localhost:3000$1"; }

# ---------------------------------------------------------------------------
echo "TC01 — typecheck, build ESM, chạy dist và dev"
cd "$APP"
out="$(pnpm --filter bff typecheck 2>&1)"; check "typecheck strict" "$?" "$out"
out="$(pnpm --filter bff build 2>&1)"; check "tsup build" "$?" "$out"
check "dist/server.js là ESM" "$(grep -qE '^import ' apps/bff/dist/server.js; echo $?)"
(cd apps/bff && exec pnpm dev) >"$TMP/dev.log" 2>&1 &
PID=$!
check "pnpm dev (tsx watch) lên được" "$(retry 40 curl -fs localhost:3000/healthz; echo $?)" "$(cat "$TMP/dev.log")"
stop_bff; retry 20 bash -c '! lsof -nP -iTCP:3000 -sTCP:LISTEN'

# ---------------------------------------------------------------------------
echo "TC02 — cấu hình mặc định, log JSON"
env -u BFF_PORT -u LOG_LEVEL -u MONGODB_URI node apps/bff/dist/server.js >"$TMP/log" 2>&1 &
PID=$!
check "node dist/server.js lắng nghe :3000" "$(retry 40 curl -fs localhost:3000/healthz; echo $?)" "$(cat "$TMP/log")"
started="$(grep '"msg":"bff started"' "$TMP/log" | head -1)"
check "log khởi động JSON: level chữ, time, msg, service=bff" "$(python3 -c '
import json,sys
r=json.loads(sys.argv[1])
sys.exit(0 if r["level"]=="info" and r["service"]=="bff" and isinstance(r["time"],str) else 1)' "$started" 2>/dev/null; echo $?)" "$started"

# ---------------------------------------------------------------------------
echo "TC03 — Mongo chạy: healthz, readyz"
check "/healthz 200" "$([[ "$(code_of /healthz)" == 200 ]] && grep -q '"status":"ok"' "$TMP/body"; echo $?)"
check "/readyz 200" "$([[ "$(code_of /readyz)" == 200 ]] && grep -q '"status":"ok"' "$TMP/body"; echo $?)" "$(cat "$TMP/body")"

# ---------------------------------------------------------------------------
echo "TC04 — Mongo dừng / bật lại"
"${C[@]}" stop mongodb >/dev/null 2>&1
start=$(date +%s); c="$(code_of /readyz)"; secs=$(( $(date +%s) - start ))
check "/readyz 503 khi Mongo dừng (thực tế: $c)" "$([[ "$c" == 503 ]] && grep -q '"error":"mongodb"' "$TMP/body"; echo $?)" "$(cat "$TMP/body")"
check "/readyz trả trong ≤ 3s (thực tế: ${secs}s)" "$([[ $secs -le 3 ]]; echo $?)"
check "/healthz vẫn 200" "$([[ "$(code_of /healthz)" == 200 ]]; echo $?)"
"${C[@]}" start mongodb >/dev/null 2>&1
ready() { [[ "$(code_of /readyz)" == 200 ]]; }
check "Mongo bật lại → /readyz 200, không restart" "$(retry 60 ready; echo $?)"
stop_bff

# ---------------------------------------------------------------------------
echo "TC05 — cấu hình sai → thoát mã khác 0"
out="$(LOG_LEVEL=verbose node apps/bff/dist/server.js 2>&1)"; code=$?
check "LOG_LEVEL=verbose → exit $code, nêu LOG_LEVEL" "$([[ $code -ne 0 ]] && grep -q LOG_LEVEL <<<"$out"; echo $?)" "$out"
out="$(BFF_PORT=abc node apps/bff/dist/server.js 2>&1)"; code=$?
check "BFF_PORT=abc → exit $code, nêu BFF_PORT" "$([[ $code -ne 0 ]] && grep -q BFF_PORT <<<"$out"; echo $?)" "$out"

echo
echo "Kết quả: $pass pass, $fail fail"
((fail == 0))
