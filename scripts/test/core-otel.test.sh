#!/usr/bin/env bash
# Kiểm thử tích hợp OpenTelemetry của core (P0-T10, TC06): trace → Tempo, metric → Prometheus.
#
#   scripts/test/core-otel.test.sh
#
# Stack observability + PG bằng compose project "snaptix-test"; core chạy ở cổng 8080.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
C=(docker compose -p snaptix-test -f "$REPO/deploy/docker-compose.yml")
TMP="$(mktemp -d)"
PID=""

pass=0 fail=0
check() {
  if [[ "$2" == 0 ]]; then pass=$((pass + 1)); echo "  ✅ $1"
  else fail=$((fail + 1)); echo "  ❌ $1"; [[ -n "${3:-}" ]] && echo "$3" | tail -15 | sed 's/^/     /'; fi
}
retry() { local n="$1"; shift; for _ in $(seq "$n"); do "$@" >/dev/null 2>&1 && return 0; sleep 1; done; return 1; }

for p in 5432 4318 9090 3200 8080; do
  if lsof -nP -iTCP:"$p" -sTCP:LISTEN >/dev/null 2>&1; then echo "Cổng $p đang bận — make down trước" >&2; exit 2; fi
done
trap '[[ -n "$PID" ]] && kill "$PID" 2>/dev/null; "${C[@]}" down -v --remove-orphans >/dev/null 2>&1; rm -rf "$TMP"' EXIT
"${C[@]}" up -d --wait postgres-core otel-collector tempo prometheus >/dev/null 2>&1 || { echo "Không bật được stack" >&2; exit 2; }
retry 60 curl -fsS localhost:3200/ready

(cd "$REPO/com/tm/server" && GOTOOLCHAIN=local go build -o "$TMP/core" ./services/core/cmd/server) || exit 2
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 OTEL_METRIC_EXPORT_INTERVAL=1000 CORE_VERSION=test \
  "$TMP/core" >"$TMP/log" 2>&1 &
PID=$!
retry 40 curl -fs localhost:8080/healthz || { echo "core không lên"; cat "$TMP/log"; exit 2; }

echo "TC06 — request có traceparent → trace trong Tempo, metric trong Prometheus"
TRACE_ID="$(openssl rand -hex 16)"
for _ in 1 2 3; do
  curl -s -o /dev/null -H "traceparent: 00-$TRACE_ID-00f067aa0ba902b7-01" localhost:8080/khong-co-route
done

in_tempo() { curl -fsS "localhost:3200/api/traces/$TRACE_ID" | grep -q '"stringValue":"core"'; }
check "Tempo có trace $TRACE_ID với service core" "$(retry 60 in_tempo; echo $?)" "$(curl -s "localhost:3200/api/traces/$TRACE_ID" | head -c 400)"
check "span có service.version=test" "$(curl -fsS "localhost:3200/api/traces/$TRACE_ID" | grep -q '"stringValue":"test"'; echo $?)"

in_prom() {
  curl -fsS --get localhost:9090/api/v1/query \
    --data-urlencode 'query=http_server_request_duration_seconds_count{service_name="core"}' | grep -q '"result":\[{'
}
check "Prometheus có http_server_request_duration_seconds_count{service_name=\"core\"}" "$(retry 60 in_prom; echo $?)"
check "log core không có lỗi export OpenTelemetry" "$(! grep -q '"msg":"opentelemetry"' "$TMP/log"; echo $?)" "$(grep opentelemetry "$TMP/log")"

kill -TERM "$PID"; wait "$PID"; code=$?; PID=""
check "dừng êm (flush OTel) thoát mã 0" "$([[ $code -eq 0 ]]; echo $?)" "$(tail -5 "$TMP/log")"

echo
echo "Kết quả: $pass pass, $fail fail"
((fail == 0))
