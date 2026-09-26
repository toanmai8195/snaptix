#!/usr/bin/env bash
# Kiểm thử pnpm workspace của com/tm/app (P0-T01b).
#
#   scripts/test/pnpm-workspace.test.sh
#
# Tạo package tạm apps/probe, packages/probe-lib; luôn dọn dẹp khi kết thúc.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
APP="$REPO/com/tm/app"
cd "$APP"

cleanup() { rm -rf apps/probe packages/probe-lib; pnpm install --silent >/dev/null 2>&1 || true; }
trap cleanup EXIT

pass=0 fail=0
check() {
  if [[ "$2" == 0 ]]; then pass=$((pass + 1)); echo "  ✅ $1"
  else fail=$((fail + 1)); echo "  ❌ $1"; [[ -n "${3:-}" ]] && echo "$3" | tail -15 | sed 's/^/     /'; fi
}

echo "TC01 — pnpm install, lockfile, packageManager, engines"
out="$(pnpm install 2>&1)"; check "pnpm install" "$?" "$out"
check "có pnpm-lock.yaml" "$([[ -f pnpm-lock.yaml ]]; echo $?)"
check "node_modules bị git ignore" "$(git -C "$REPO" check-ignore -q com/tm/app/node_modules; echo $?)"
check "packageManager pin pnpm@x.y.z" "$(node -e 'process.exit(/^pnpm@\d+\.\d+\.\d+$/.test(require("./package.json").packageManager)?0:1)'; echo $?)"
check "engines.node >= 22" "$(node -e 'process.exit((require("./package.json").engines||{}).node===">=22"?0:1)'; echo $?)"

echo "TC02 — workspace chưa có package: build/test/lint exit 0"
for s in build test lint; do
  out="$(pnpm "$s" 2>&1)"; check "pnpm $s" "$?" "$out"
done

echo "TC03 — nhận apps/* và packages/*, filter theo tên"
mk() { # mk <dir> <name>
  mkdir -p "$1"
  cat > "$1/package.json" <<EOF
{ "name": "$2", "private": true, "version": "0.0.0",
  "scripts": { "build": "echo BUILD-$2", "test": "echo TEST-$2", "lint": "echo LINT-$2" } }
EOF
}
mk apps/probe probe
mk packages/probe-lib probe-lib
pnpm install --silent >/dev/null 2>&1
out="$(pnpm build 2>&1)"
check "pnpm build chạy apps/probe" "$(grep -q 'BUILD-probe$' <<<"$out"; echo $?)" "$out"
check "pnpm build chạy packages/probe-lib" "$(grep -q 'BUILD-probe-lib' <<<"$out"; echo $?)" "$out"
out="$(pnpm test 2>&1)"
check "pnpm test chạy cả hai package" "$(grep -q 'TEST-probe$' <<<"$out" && grep -q 'TEST-probe-lib' <<<"$out"; echo $?)" "$out"
out="$(pnpm --filter probe build 2>&1)"
check "--filter probe chỉ chạy apps/probe" "$(grep -q 'BUILD-probe$' <<<"$out" && ! grep -q 'BUILD-probe-lib' <<<"$out"; echo $?)" "$out"

echo "TC04 — filter theo thay đổi so với HEAD (repo git tạm)"
TMPREPO="$(mktemp -d)"
cp package.json pnpm-workspace.yaml "$TMPREPO/"
(
  cd "$TMPREPO" && git init -q && git config user.email t@t && git config user.name t
  mk apps/probe probe && mk packages/probe-lib probe-lib
  echo "export {}" > packages/probe-lib/index.js
  pnpm install --silent >/dev/null 2>&1 && git add -A && git commit -qm base
  # sửa file ĐÃ track: pnpm so sánh bằng git diff, file untracked không được tính
  echo "// changed" >> packages/probe-lib/index.js
)
out="$(cd "$TMPREPO" && pnpm --filter "...[HEAD]" build 2>&1)"
check "chạy package có thay đổi (probe-lib)" "$(grep -q 'BUILD-probe-lib' <<<"$out"; echo $?)" "$out"
check "không chạy package không đổi (probe)" "$(! grep -q 'BUILD-probe$' <<<"$out"; echo $?)" "$out"
rm -rf "$TMPREPO"

echo
echo "Kết quả: $pass pass, $fail fail"
((fail == 0))
