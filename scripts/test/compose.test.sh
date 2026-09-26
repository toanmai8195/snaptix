#!/usr/bin/env bash
# Kiểm thử deploy/docker-compose.yml (P0-T02).
#
#   scripts/test/compose.test.sh
#
# Chạy với project riêng "snaptix-test" (volume riêng) để không đụng dữ liệu dev.
# Cần các cổng host trống: dừng stack dev trước (docker compose -f deploy/docker-compose.yml down).
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO"
C=(docker compose -p snaptix-test -f deploy/docker-compose.yml)
SERVICES=(postgres-core postgres-analytics mongodb redis otel-collector prometheus grafana tempo)
PORTS=(5432 5433 27017 6379 4317 4318 9090 3100 3200)

pass=0 fail=0
check() {
  if [[ "$2" == 0 ]]; then pass=$((pass + 1)); echo "  ✅ $1"
  else fail=$((fail + 1)); echo "  ❌ $1"; [[ -n "${3:-}" ]] && echo "$3" | tail -15 | sed 's/^/     /'; fi
}
retry() { # retry <số lần> <lệnh...>: thử lại mỗi giây
  local n="$1"; shift
  for _ in $(seq "$n"); do "$@" >/dev/null 2>&1 && return 0; sleep 1; done
  return 1
}

for p in "${PORTS[@]}"; do
  if lsof -nP -iTCP:"$p" -sTCP:LISTEN >/dev/null 2>&1; then
    echo "Cổng $p đang bận — dừng stack dev trước: docker compose -f deploy/docker-compose.yml down" >&2
    exit 2
  fi
done
trap '"${C[@]}" down -v --remove-orphans >/dev/null 2>&1' EXIT

psql_core()      { "${C[@]}" exec -T postgres-core psql -U snaptix -d core -tAc "$1"; }
psql_analytics() { "${C[@]}" exec -T postgres-analytics psql -U snaptix -d analytics -tAc "$1"; }

# ---------------------------------------------------------------------------
echo "TC01 — compose config hợp lệ, đủ service, image pin phiên bản"
cfg="$("${C[@]}" config 2>&1)"; check "docker compose config" "$?" "$cfg"
listed="$("${C[@]}" config --services 2>/dev/null)"
for s in "${SERVICES[@]}"; do check "có service $s" "$(grep -qx "$s" <<<"$listed"; echo $?)"; done
images="$("${C[@]}" config --images 2>/dev/null)"
check "không image nào dùng latest / thiếu tag" "$(awk -F: 'NF<2 || $NF=="latest"{bad=1} END{exit bad}' <<<"$images"; echo $?)" "$images"

# ---------------------------------------------------------------------------
echo "TC02 — up --wait, healthy, cổng host"
start=$(date +%s)
out="$("${C[@]}" up -d --wait --wait-timeout 120 2>&1)"; check "up -d --wait (≤120s)" "$?" "$out"
echo "     (mất $(( $(date +%s) - start ))s)"
ps="$("${C[@]}" ps --format '{{.Service}} {{.State}} {{.Health}}' 2>/dev/null)"
for s in "${SERVICES[@]}"; do check "$s đang chạy" "$(grep -q "^$s running" <<<"$ps"; echo $?)" "$ps"; done
check "không service nào unhealthy" "$(! grep -q unhealthy <<<"$ps"; echo $?)" "$ps"
for p in "${PORTS[@]}"; do check "cổng host $p đang lắng nghe" "$(retry 10 nc -z localhost "$p"; echo $?)"; done

# ---------------------------------------------------------------------------
echo "TC03 — kết nối kho dữ liệu"
check "PG core: SELECT 1" "$([[ "$(psql_core 'SELECT 1')" == 1 ]]; echo $?)"
check "PG analytics: SELECT 1" "$([[ "$(psql_analytics 'SELECT 1')" == 1 ]]; echo $?)"
check "PG core và analytics là 2 instance khác nhau" \
  "$([[ "$(psql_core 'SELECT current_database()')" == core && "$(psql_analytics 'SELECT current_database()')" == analytics ]]; echo $?)"
out="$("${C[@]}" exec -T mongodb mongosh --quiet --eval 'db.adminCommand("ping").ok' 2>&1)"
check "Mongo ping ok: 1" "$([[ "$out" == 1 ]]; echo $?)" "$out"
out="$("${C[@]}" exec -T redis redis-cli ping 2>&1)"
check "Redis PING → PONG" "$([[ "$out" == PONG ]]; echo $?)" "$out"

# ---------------------------------------------------------------------------
echo "TC04 — observability"
check "Prometheus /-/ready" "$(retry 30 curl -fsS localhost:9090/-/ready; echo $?)"
target_up() { curl -fsS 'localhost:9090/api/v1/targets?state=active' | grep -q '"job":"otel-collector".*"health":"up"'; }
check "target otel-collector: up" "$(retry 40 target_up; echo $?)"
check "Grafana /api/health" "$(retry 30 curl -fsS localhost:3100/api/health; echo $?)"
ds="$(curl -fsS localhost:3100/api/datasources 2>&1)"
check "Grafana có datasource Prometheus" "$(grep -q '"type":"prometheus"' <<<"$ds"; echo $?)" "$ds"
check "Grafana có datasource Tempo" "$(grep -q '"type":"tempo"' <<<"$ds"; echo $?)" "$ds"
check "Tempo /ready" "$(retry 60 curl -fsS localhost:3200/ready; echo $?)"
check "otel-collector health (13133)" "$(retry 30 curl -fsS localhost:13133/; echo $?)"

TRACE_ID="$(openssl rand -hex 16)"; SPAN_ID="$(openssl rand -hex 8)"; NOW="$(date +%s)000000000"
span='{"resourceSpans":[{"resource":{"attributes":[{"key":"service.name","value":{"stringValue":"compose-test"}}]},
"scopeSpans":[{"spans":[{"traceId":"'"$TRACE_ID"'","spanId":"'"$SPAN_ID"'","name":"probe","kind":1,
"startTimeUnixNano":"'"$NOW"'","endTimeUnixNano":"'"$NOW"'"}]}]}]}'
send_span() { curl -fsS -X POST localhost:4318/v1/traces -H 'Content-Type: application/json' -d "$span"; }
check "gửi span qua OTLP HTTP :4318" "$(retry 20 send_span; echo $?)"
trace_found() { curl -fsS "localhost:3200/api/traces/$TRACE_ID" | grep -q compose-test; }
check "truy vấn được trace $TRACE_ID trong Tempo" "$(retry 60 trace_found; echo $?)"

# ---------------------------------------------------------------------------
echo "TC05 — dữ liệu bền qua restart, down -v xoá sạch"
psql_core "CREATE TABLE persist_probe (id int); INSERT INTO persist_probe VALUES (42);" >/dev/null
"${C[@]}" down >/dev/null 2>&1
out="$("${C[@]}" up -d --wait --wait-timeout 120 2>&1)"; check "up lại sau down" "$?" "$out"
check "bảng persist_probe còn sau restart" "$([[ "$(psql_core 'SELECT id FROM persist_probe')" == 42 ]]; echo $?)"
"${C[@]}" down -v >/dev/null 2>&1
vols="$(docker volume ls -q --filter label=com.docker.compose.project=snaptix-test)"
check "down -v xoá hết volume" "$([[ -z "$vols" ]]; echo $?)" "$vols"

echo
echo "Kết quả: $pass pass, $fail fail"
((fail == 0))
