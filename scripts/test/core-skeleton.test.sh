#!/usr/bin/env bash
# Kiểm thử tích hợp skeleton core service (P0-T07).
#
#   scripts/test/core-skeleton.test.sh
#
# PG bằng compose project "snaptix-test"; core build bằng go, chạy ở cổng 8080.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SERVER="$REPO/com/tm/server"
C=(docker compose -p snaptix-test -f "$REPO/deploy/docker-compose.yml")
TMP="$(mktemp -d)"
BIN="$TMP/core"
PID=""

pass=0 fail=0
check() {
  if [[ "$2" == 0 ]]; then pass=$((pass + 1)); echo "  ✅ $1"
  else fail=$((fail + 1)); echo "  ❌ $1"; [[ -n "${3:-}" ]] && echo "$3" | tail -15 | sed 's/^/     /'; fi
}
retry() { local n="$1"; shift; for _ in $(seq "$n"); do "$@" >/dev/null 2>&1 && return 0; sleep 0.5; done; return 1; }

for p in 5432 8080; do
  if lsof -nP -iTCP:"$p" -sTCP:LISTEN >/dev/null 2>&1; then echo "Cổng $p đang bận" >&2; exit 2; fi
done
stop_core() { [[ -n "$PID" ]] && kill "$PID" 2>/dev/null && wait "$PID" 2>/dev/null; PID=""; }
trap 'stop_core; "${C[@]}" down -v --remove-orphans >/dev/null 2>&1; rm -rf "$TMP"' EXIT
"${C[@]}" up -d --wait postgres-core >/dev/null 2>&1 || { echo "Không bật được PG" >&2; exit 2; }

code_of() { curl -s -o "$TMP/body" -w '%{http_code}' --max-time 5 "localhost:8080$1"; }

# ---------------------------------------------------------------------------
echo "TC01 — build, vet, gazelle, bazel"
cd "$SERVER"
out="$(GOTOOLCHAIN=local go build -o "$BIN" ./services/core/cmd/server 2>&1 && GOTOOLCHAIN=local go vet ./... 2>&1)"
check "go build + go vet" "$?" "$out"
out="$(bazel run //:gazelle -- -mode=diff 2>&1)"; check "gazelle không tạo diff" "$?" "$out"
out="$(bazel build //services/core/... 2>&1)"; check "bazel build //services/core/..." "$?" "$out"
check "có target :server_image" "$(bazel query //services/core/cmd/server:server_image >/dev/null 2>&1; echo $?)"

# ---------------------------------------------------------------------------
echo "TC02 — cấu hình mặc định, log JSON"
env -u CORE_HTTP_ADDR -u CORE_DATABASE_URL -u LOG_LEVEL "$BIN" >"$TMP/log" 2>&1 &
PID=$!
check "core lắng nghe :8080" "$(retry 40 curl -fs localhost:8080/healthz; echo $?)" "$(cat "$TMP/log")"
first="$(head -1 "$TMP/log")"
check "dòng log đầu là JSON có time/level/msg/addr" "$(python3 -c '
import json,sys
r=json.loads(sys.argv[1])
sys.exit(0 if all(k in r for k in ("time","level","msg","addr")) and r["msg"]=="core started" and r["addr"]==":8080" else 1)' "$first" 2>/dev/null; echo $?)" "$first"

# ---------------------------------------------------------------------------
echo "TC03 — PG chạy: healthz, readyz"
check "/healthz 200 {\"status\":\"ok\"}" "$([[ "$(code_of /healthz)" == 200 ]] && grep -q '"status":"ok"' "$TMP/body"; echo $?)"
check "/readyz 200 {\"status\":\"ok\"}" "$([[ "$(code_of /readyz)" == 200 ]] && grep -q '"status":"ok"' "$TMP/body"; echo $?)" "$(cat "$TMP/body")"

# ---------------------------------------------------------------------------
echo "TC04 — PG dừng / bật lại"
"${C[@]}" stop postgres-core >/dev/null 2>&1
start=$(date +%s)
c="$(code_of /readyz)"; secs=$(( $(date +%s) - start ))
check "/readyz 503 khi PG dừng (thực tế: $c)" "$([[ "$c" == 503 ]] && grep -q '"status":"unavailable"' "$TMP/body"; echo $?)" "$(cat "$TMP/body")"
check "/readyz trả trong ≤ 3s (thực tế: ${secs}s)" "$([[ $secs -le 3 ]]; echo $?)"
check "/healthz vẫn 200" "$([[ "$(code_of /healthz)" == 200 ]]; echo $?)"
"${C[@]}" start postgres-core >/dev/null 2>&1
ready() { [[ "$(code_of /readyz)" == 200 ]]; }
check "PG bật lại → /readyz về 200, không restart core" "$(retry 60 ready; echo $?)"

# ---------------------------------------------------------------------------
echo "TC05 — /metrics"
check "/metrics 200 có go_goroutines" "$([[ "$(code_of /metrics)" == 200 ]] && grep -q '^go_goroutines' "$TMP/body"; echo $?)"

# ---------------------------------------------------------------------------
echo "TC06 — cấu hình sai, route lạ"
check "route lạ → 404" "$([[ "$(code_of /khong-ton-tai)" == 404 ]]; echo $?)"
stop_core
out="$(LOG_LEVEL=verbose "$BIN" 2>&1)"; code=$?
check "LOG_LEVEL=verbose → exit khác 0, lỗi nêu LOG_LEVEL" "$([[ $code -ne 0 ]] && grep -q 'LOG_LEVEL' <<<"$out"; echo $?)" "$out"

echo
echo "Kết quả: $pass pass, $fail fail"
((fail == 0))
