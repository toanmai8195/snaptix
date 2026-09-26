#!/usr/bin/env bash
# Kiểm thử image OCI của core server + worker (P0-T10a).
#
#   scripts/test/core-image.test.sh
#
# PG bằng compose project "snaptix-test"; container server map cổng 18080 → 8080.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SERVER="$REPO/com/tm/server"
C=(docker compose -p snaptix-test -f "$REPO/deploy/docker-compose.yml")
SRV_IMG="com.tm.go.core-server:v1.0.0"
WRK_IMG="com.tm.go.core-worker:v1.0.0"
SRV_C="snaptix-test-core-server"
WRK_C="snaptix-test-core-worker"

pass=0 fail=0
check() {
  if [[ "$2" == 0 ]]; then pass=$((pass + 1)); echo "  ✅ $1"
  else fail=$((fail + 1)); echo "  ❌ $1"; [[ -n "${3:-}" ]] && echo "$3" | tail -15 | sed 's/^/     /'; fi
}
retry() { local n="$1"; shift; for _ in $(seq "$n"); do "$@" >/dev/null 2>&1 && return 0; sleep 0.5; done; return 1; }

for p in 5432 18080; do
  if lsof -nP -iTCP:"$p" -sTCP:LISTEN >/dev/null 2>&1; then echo "Cổng $p đang bận" >&2; exit 2; fi
done
cleanup() {
  docker rm -f "$SRV_C" "$WRK_C" >/dev/null 2>&1
  docker rmi -f "$SRV_IMG" "$WRK_IMG" >/dev/null 2>&1
  "${C[@]}" down -v --remove-orphans >/dev/null 2>&1
}
trap cleanup EXIT
"${C[@]}" up -d --wait postgres-core >/dev/null 2>&1 || { echo "Không bật được PG" >&2; exit 2; }
cd "$SERVER"

# ---------------------------------------------------------------------------
echo "TC01 — target và tag image"
t="$(bazel query '//services/core/cmd/server:all + //services/core/cmd/worker:all' 2>/dev/null)"
for x in server/server server/server_image server/server_docker worker/worker worker/worker_image worker/worker_docker; do
  pkg="${x%%/*}"; name="${x##*/}"
  check "có //services/core/cmd/$pkg:$name" "$(grep -qx "//services/core/cmd/$pkg:$name" <<<"$t"; echo $?)"
done
out="$(bazel run //:gazelle -- -mode=diff 2>&1)"; check "gazelle giữ image_name (không diff)" "$?" "$out"

# ---------------------------------------------------------------------------
echo "TC02 — build hai kiến trúc"
for cfg in linux-arm64 linux-amd64; do
  out="$(bazel build --config=$cfg //services/core/cmd/server:server_image //services/core/cmd/worker:worker_image 2>&1)"
  check "build image --config=$cfg" "$?" "$out"
done

# ---------------------------------------------------------------------------
echo "TC03 — chạy container server"
out="$(bazel run --config=linux-arm64 //services/core/cmd/server:server_docker 2>&1)"; check "load image server vào Docker" "$?" "$out"
check "image có tag $SRV_IMG" "$(docker image inspect "$SRV_IMG" >/dev/null 2>&1; echo $?)" "$out"
docker run -d --name "$SRV_C" -p 18080:8080 \
  -e CORE_DATABASE_URL="postgres://snaptix:snaptix@host.docker.internal:5432/core?sslmode=disable" \
  -e OTEL_SDK_DISABLED=true "$SRV_IMG" >/dev/null
check "/healthz 200" "$(retry 40 curl -fs localhost:18080/healthz; echo $?)" "$(docker logs "$SRV_C" 2>&1)"
check "/readyz 200 (PG qua host.docker.internal)" "$(retry 20 curl -fs localhost:18080/readyz; echo $?)" "$(docker logs "$SRV_C" 2>&1)"
check "log JSON ra stdout" "$(docker logs "$SRV_C" 2>/dev/null | head -1 | python3 -c 'import json,sys; json.loads(sys.stdin.readline())' 2>/dev/null; echo $?)"

# ---------------------------------------------------------------------------
echo "TC04 — docker stop → dừng êm"
docker stop -t 10 "$SRV_C" >/dev/null
code="$(docker inspect -f '{{.State.ExitCode}}' "$SRV_C")"
logs="$(docker logs "$SRV_C" 2>/dev/null)"
check "thoát mã 0 (thực tế: $code)" "$([[ "$code" == 0 ]]; echo $?)" "$logs"
check "log shutting down → http server stopped" "$(grep -q '"msg":"shutting down"' <<<"$logs" && grep -q '"msg":"http server stopped"' <<<"$logs"; echo $?)" "$logs"

# ---------------------------------------------------------------------------
echo "TC05 — cấu hình image"
user="$(docker image inspect -f '{{.Config.User}}' "$SRV_IMG")"
check "user non-root 65532:65532 (thực tế: $user)" "$([[ "$user" == 65532:65532 ]]; echo $?)"
check "expose 8080/tcp" "$(docker image inspect -f '{{json .Config.ExposedPorts}}' "$SRV_IMG" | grep -q '8080/tcp'; echo $?)"
arch="$(docker image inspect -f '{{.Architecture}}' "$SRV_IMG")"
check "kiến trúc arm64 (thực tế: $arch)" "$([[ "$arch" == arm64 ]]; echo $?)"
size=$(( $(docker image inspect -f '{{.Size}}' "$SRV_IMG") / 1024 / 1024 ))
check "kích thước ${size}MB < 60MB" "$([[ $size -lt 60 ]]; echo $?)"

# ---------------------------------------------------------------------------
echo "TC06 — worker"
out="$(bazel run --config=linux-arm64 //services/core/cmd/worker:worker_docker 2>&1)"; check "load image worker" "$?" "$out"
check "image có tag $WRK_IMG" "$(docker image inspect "$WRK_IMG" >/dev/null 2>&1; echo $?)" "$out"
docker run -d --name "$WRK_C" -e OTEL_SDK_DISABLED=true "$WRK_IMG" >/dev/null
started() { docker logs "$WRK_C" 2>&1 | grep -q '"msg":"worker started"'; }
check "log worker started" "$(retry 20 started; echo $?)" "$(docker logs "$WRK_C" 2>&1)"
docker stop -t 10 "$WRK_C" >/dev/null
code="$(docker inspect -f '{{.State.ExitCode}}' "$WRK_C")"
check "stop → worker stopped, thoát mã 0 (thực tế: $code)" "$([[ "$code" == 0 ]] && docker logs "$WRK_C" 2>&1 | grep -q '"msg":"worker stopped"'; echo $?)" "$(docker logs "$WRK_C" 2>&1)"

echo
echo "Kết quả: $pass pass, $fail fail"
((fail == 0))
