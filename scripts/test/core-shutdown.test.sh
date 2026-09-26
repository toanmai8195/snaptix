#!/usr/bin/env bash
# Kiểm thử graceful shutdown trên binary core (P0-T09, TC05).
#
#   scripts/test/core-shutdown.test.sh
#
# Không cần PG (pool mở kết nối lười). Cổng 8080 phải trống.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

pass=0 fail=0
check() {
  if [[ "$2" == 0 ]]; then pass=$((pass + 1)); echo "  ✅ $1"
  else fail=$((fail + 1)); echo "  ❌ $1"; [[ -n "${3:-}" ]] && echo "$3" | tail -15 | sed 's/^/     /'; fi
}
retry() { local n="$1"; shift; for _ in $(seq "$n"); do "$@" >/dev/null 2>&1 && return 0; sleep 0.2; done; return 1; }

if lsof -nP -iTCP:8080 -sTCP:LISTEN >/dev/null 2>&1; then echo "Cổng 8080 đang bận" >&2; exit 2; fi
(cd "$REPO/com/tm/server" && GOTOOLCHAIN=local go build -o "$TMP/core" ./services/core/cmd/server) || exit 2

for sig in TERM INT; do
  echo "TC05 — SIG$sig"
  "$TMP/core" >"$TMP/log-$sig" 2>&1 &
  pid=$!
  retry 50 curl -fs localhost:8080/healthz || { echo "core không lên"; cat "$TMP/log-$sig"; exit 2; }

  start=$(python3 -c 'import time; print(time.time())')
  kill -"$sig" "$pid"
  wait "$pid"; code=$?
  secs=$(python3 -c "import time; print(round(time.time()-$start, 2))")

  log="$(cat "$TMP/log-$sig")"
  check "thoát mã 0 (thực tế: $code)" "$([[ $code -eq 0 ]]; echo $?)" "$log"
  check "dừng trong < 2s (thực tế: ${secs}s)" "$(python3 -c "import sys; sys.exit(0 if $secs < 2 else 1)"; echo $?)"
  order="$(grep -o '"msg":"\(shutting down\|http server stopped\|database pool closed\)"' <<<"$log" | tr '\n' ' ')"
  check "thứ tự log: shutting down → http server stopped → database pool closed" \
    "$([[ "$order" == '"msg":"shutting down" "msg":"http server stopped" "msg":"database pool closed" ' ]]; echo $?)" "$log"
  check "cổng 8080 đã giải phóng" "$(! lsof -nP -iTCP:8080 -sTCP:LISTEN >/dev/null 2>&1; echo $?)"
done

echo
echo "Kết quả: $pass pass, $fail fail"
((fail == 0))
