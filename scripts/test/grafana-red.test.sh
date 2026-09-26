#!/usr/bin/env bash
# Kiểm thử dashboard Grafana RED (P0-T06).
#
#   scripts/test/grafana-red.test.sh
#
# Bật observability bằng compose project "snaptix-test", gửi metric OTLP giả lập rồi truy vấn.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO"
C=(docker compose -p snaptix-test -f deploy/docker-compose.yml)
DASH=deploy/observability/grafana/provisioning/dashboards/json/red.json
M=http_server_request_duration_seconds

pass=0 fail=0
check() {
  if [[ "$2" == 0 ]]; then pass=$((pass + 1)); echo "  ✅ $1"
  else fail=$((fail + 1)); echo "  ❌ $1"; [[ -n "${3:-}" ]] && echo "$3" | tail -15 | sed 's/^/     /'; fi
}
retry() { local n="$1"; shift; for _ in $(seq "$n"); do "$@" >/dev/null 2>&1 && return 0; sleep 1; done; return 1; }

for p in 4318 9090 3100 3200; do
  if lsof -nP -iTCP:"$p" -sTCP:LISTEN >/dev/null 2>&1; then echo "Cổng $p đang bận — make down trước" >&2; exit 2; fi
done
trap '"${C[@]}" down -v --remove-orphans >/dev/null 2>&1' EXIT
"${C[@]}" up -d --wait otel-collector prometheus grafana tempo >/dev/null 2>&1 || { echo "Không bật được stack" >&2; exit 2; }
retry 60 curl -fsS localhost:3100/api/health
retry 60 curl -fsS localhost:9090/-/ready

promq() { # promq <expr> → in JSON kết quả
  curl -fsS --get localhost:9090/api/v1/query --data-urlencode "query=$1"
}

# ---------------------------------------------------------------------------
echo "TC01 — dashboard được provision"
found() { curl -fsS 'localhost:3100/api/search?query=RED' | grep -q '"uid":"snaptix-red"'; }
check "Grafana tìm thấy dashboard uid snaptix-red" "$(retry 30 found; echo $?)"
d="$(curl -fsS localhost:3100/api/dashboards/uid/snaptix-red 2>&1)"
check "tiêu đề 'snaptix — RED'" "$(grep -q '"title":"snaptix — RED"' <<<"$d"; echo $?)" "$d"
check "được provision (không tạo tay)" "$(grep -q '"provisioned":true' <<<"$d"; echo $?)"
check "có biến service" "$(grep -q '"name":"service"' <<<"$d"; echo $?)"
for p in "Rate" "Errors" "Duration"; do
  check "có panel $p" "$(grep -q "\"title\":\"$p" <<<"$d"; echo $?)"
done

# ---------------------------------------------------------------------------
echo "TC02 — mọi PromQL hợp lệ"
exprs="$(python3 -c '
import json,sys
d=json.load(open(sys.argv[1]))
for p in d["panels"]:
    for t in p.get("targets",[]): print(t["expr"])
' "$DASH")"
n=0; bad=0
while IFS= read -r e; do
  q="${e//\$service/.+}"; q="${q//\$__rate_interval/1m}"
  n=$((n + 1))
  promq "$q" | grep -q '"status":"success"' || { bad=$((bad + 1)); echo "     lỗi: $e"; }
done <<<"$exprs"
check "$n biểu thức, $bad lỗi" "$([[ $n -ge 7 && $bad -eq 0 ]]; echo $?)"

# ---------------------------------------------------------------------------
echo "TC03 — metric OTLP giả lập (red-probe: 90% 200, 10% 500) → panel có dữ liệu"
python3 - <<'PY' &
import json, time, urllib.request
bounds = [0.005, 0.01, 0.05, 0.1, 0.5, 1]
start = time.time_ns()
def point(code, count, bucket, sum_):
    counts = [0] * (len(bounds) + 1); counts[bucket] = count
    return {"attributes": [{"key": "http.response.status_code", "value": {"intValue": str(code)}},
                           {"key": "http.route", "value": {"stringValue": "/probe"}}],
            "startTimeUnixNano": str(start), "timeUnixNano": str(time.time_ns()),
            "count": str(count), "sum": sum_, "bucketCounts": [str(c) for c in counts], "explicitBounds": bounds}
for i in range(1, 31):
    body = {"resourceMetrics": [{"resource": {"attributes": [{"key": "service.name", "value": {"stringValue": "red-probe"}}]},
        "scopeMetrics": [{"metrics": [{"name": "http.server.request.duration", "unit": "s",
            "histogram": {"aggregationTemporality": 2, "dataPoints": [
                point(200, 9 * i, 2, 0.02 * 9 * i), point(500, i, 4, 0.2 * i)]}}]}]}]}
    req = urllib.request.Request("http://localhost:4318/v1/metrics", json.dumps(body).encode(),
                                 {"Content-Type": "application/json"})
    urllib.request.urlopen(req, timeout=5).read()
    time.sleep(1)
PY
sender=$!
series() { promq "${M}_count{service_name=\"red-probe\"}" | grep -q '"result":\[{'; }
check "Prometheus có ${M}_count{service_name=red-probe}" "$(retry 60 series; echo $?)"
wait "$sender"
sleep 6 # thêm vài lượt scrape sau lần gửi cuối
val() { promq "$1" | python3 -c 'import json,sys; r=json.load(sys.stdin)["data"]["result"]; print(r[0]["value"][1] if r else "")'; }
rate="$(val "sum(rate(${M}_count{service_name=\"red-probe\"}[1m]))")"
check "Rate > 0 (thực tế: $rate req/s)" "$(python3 -c "import sys; sys.exit(0 if '$rate' and float('$rate')>0 else 1)"; echo $?)"
err="$(val "sum(rate(${M}_count{service_name=\"red-probe\",http_response_status_code=~\"5..\"}[1m])) / clamp_min(sum(rate(${M}_count{service_name=\"red-probe\"}[1m])), 1e-9)")"
check "Errors ≈ 10% (thực tế: $err)" "$(python3 -c "import sys; sys.exit(0 if '$err' and 0.05<float('$err')<0.15 else 1)"; echo $?)"
p95="$(val "histogram_quantile(0.95, sum by (le) (rate(${M}_bucket{service_name=\"red-probe\"}[1m])))")"
check "Duration p95 có giá trị (thực tế: ${p95}s)" "$(python3 -c "import sys; sys.exit(0 if '$p95' and float('$p95')>0 else 1)"; echo $?)"

# ---------------------------------------------------------------------------
echo "TC04 — biến service"
vals="$(curl -fsS --get localhost:9090/api/v1/label/service_name/values --data-urlencode "match[]=${M}_count" 2>&1)"
check "label_values(${M}_count, service_name) chứa red-probe" "$(grep -q '"red-probe"' <<<"$vals"; echo $?)" "$vals"

echo
echo "Kết quả: $pass pass, $fail fail"
((fail == 0))
