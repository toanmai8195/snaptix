#!/usr/bin/env bash
# Kiểm thử tích hợp middleware của core trên binary thật (P0-T08, TC06).
#
#   scripts/test/core-middleware.test.sh
#
# Không cần PG: chỉ kiểm tra request ID + trace_id trong access log. Cổng 8080 phải trống.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TMP="$(mktemp -d)"
PID=""

pass=0 fail=0
check() {
  if [[ "$2" == 0 ]]; then pass=$((pass + 1)); echo "  ✅ $1"
  else fail=$((fail + 1)); echo "  ❌ $1"; [[ -n "${3:-}" ]] && echo "$3" | tail -15 | sed 's/^/     /'; fi
}
retry() { local n="$1"; shift; for _ in $(seq "$n"); do "$@" >/dev/null 2>&1 && return 0; sleep 0.5; done; return 1; }

if lsof -nP -iTCP:8080 -sTCP:LISTEN >/dev/null 2>&1; then echo "Cổng 8080 đang bận" >&2; exit 2; fi
trap '[[ -n "$PID" ]] && kill "$PID" 2>/dev/null; rm -rf "$TMP"' EXIT

(cd "$REPO/com/tm/server" && GOTOOLCHAIN=local go build -o "$TMP/core" ./services/core/cmd/server) || exit 2
"$TMP/core" >"$TMP/log" 2>&1 &
PID=$!
retry 40 curl -fs localhost:8080/healthz || { echo "core không lên" >&2; cat "$TMP/log"; exit 2; }

echo "TC06 — request có traceparent → X-Request-ID + trace_id trong access log"
TRACE_ID="$(openssl rand -hex 16)"
hdr="$(curl -s -D - -o /dev/null -H "traceparent: 00-$TRACE_ID-00f067aa0ba902b7-01" localhost:8080/khong-co-route)"
rid="$(sed -n 's/^[Xx]-[Rr]equest-[Ii][Dd]: *\([^[:space:]]*\).*/\1/p' <<<"$hdr")"
check "response có X-Request-ID (32 hex)" "$([[ "$rid" =~ ^[0-9a-f]{32}$ ]]; echo $?)" "$hdr"
sleep 0.3
line="$(grep '"msg":"http request"' "$TMP/log" | grep '"path":"/khong-co-route"' | tail -1)"
check "có dòng access log cho request" "$([[ -n "$line" ]]; echo $?)" "$(cat "$TMP/log")"
check "access log trace_id = trace ID của traceparent" "$(grep -q "\"trace_id\":\"$TRACE_ID\"" <<<"$line"; echo $?)" "$line"
check "access log request_id = X-Request-ID trả về" "$(grep -q "\"request_id\":\"$rid\"" <<<"$line"; echo $?)" "$line"
check "access log status 404" "$(grep -q '"status":404' <<<"$line"; echo $?)" "$line"

echo
echo "Kết quả: $pass pass, $fail fail"
((fail == 0))
