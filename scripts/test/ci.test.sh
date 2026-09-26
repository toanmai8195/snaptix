#!/usr/bin/env bash
# Kiểm thử workflow CI và scripts/ci/* (P0-T04).
#
#   scripts/test/ci.test.sh
#
# Tạo package Go tạm trong com/tm/server/pkg (ci_a, ci_b, ci_c, ...); luôn dọn khi kết thúc.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SERVER="$REPO/com/tm/server"
cd "$REPO"
ACTIONLINT_VERSION="v1.7.8"
WF=.github/workflows/ci.yml

cleanup() { rm -rf "$SERVER"/pkg/ci_*; }
trap cleanup EXIT
cleanup

pass=0 fail=0
check() {
  if [[ "$2" == 0 ]]; then pass=$((pass + 1)); echo "  ✅ $1"
  else fail=$((fail + 1)); echo "  ❌ $1"; [[ -n "${3:-}" ]] && echo "$3" | tail -15 | sed 's/^/     /'; fi
}
has() { grep -qE "$1" "$WF"; }

# ---------------------------------------------------------------------------
echo "TC01 — actionlint + cấu trúc workflow"
out="$(GOTOOLCHAIN=local go run "github.com/rhysd/actionlint/cmd/actionlint@$ACTIONLINT_VERSION" 2>&1)"
check "actionlint $ACTIONLINT_VERSION không lỗi" "$?" "$out"
check "trigger pull_request" "$(has '^  pull_request:'; echo $?)"
check "trigger push vào main" "$(has 'branches: \[main\]'; echo $?)"
check "checkout fetch-depth: 0" "$(has 'fetch-depth: 0'; echo $?)"
check "job server có điều kiện outputs.server" "$(has "if: needs.changes.outputs.server == 'true'"; echo $?)"
check "job app có điều kiện outputs.app" "$(has "if: needs.changes.outputs.app == 'true'"; echo $?)"
check "job repo chạy run.sh repo, không điều kiện" "$(awk '/^  repo:/{f=1} f&&/^  [a-z]+:/&&!/^  repo:/{exit} f' "$WF" | grep -q 'run.sh repo' && ! awk '/^  repo:/{f=1} f&&/^  [a-z]+:/&&!/^  repo:/{exit} f' "$WF" | grep -q '^    if:'; echo $?)"

# ---------------------------------------------------------------------------
echo "TC02 — affected.sh"
aff() { printf '%s\n' "$@" | scripts/ci/affected.sh - | tr '\n' ' '; }
check "chỉ server → server=true app=false" "$([[ "$(aff com/tm/server/pkg/x.go)" == "server=true app=false " ]]; echo $?)"
check "chỉ app → server=false app=true" "$([[ "$(aff com/tm/app/apps/bff/a.ts)" == "server=false app=true " ]]; echo $?)"
check "docs + deploy → cả hai false" "$([[ "$(aff com/tm/docs/README.md deploy/docker-compose.yml README.md)" == "server=false app=false " ]]; echo $?)"
check ".github → cả hai true" "$([[ "$(aff .github/workflows/ci.yml)" == "server=true app=true " ]]; echo $?)"
check "scripts/ci → cả hai true" "$([[ "$(aff scripts/ci/run.sh)" == "server=true app=true " ]]; echo $?)"
check "BASE không hợp lệ (0000…) → cả hai true" "$([[ "$(scripts/ci/affected.sh 0000000000000000000000000000000000000000 | tr '\n' ' ')" == "server=true app=true " ]]; echo $?)"
check "BASE = HEAD (không thay đổi) → cả hai false" "$([[ "$(scripts/ci/affected.sh HEAD | tr '\n' ' ')" == "server=false app=false " ]]; echo $?)"

# ---------------------------------------------------------------------------
echo "TC03 — bazel-affected-tests.sh (a ← b, c độc lập)"
mkpkg() { # mkpkg <tên> [import]
  local d="$SERVER/pkg/$1"; mkdir -p "$d"
  if [[ -n "${2:-}" ]]; then
    printf 'package %s\n\nimport "github.com/toanmai8195/snaptix/com/tm/server/pkg/%s"\n\n// V trả giá trị kiểm thử.\nfunc V() int { return %s.V() + 1 }\n' "$1" "$2" "$2" > "$d/$1.go"
  else
    printf 'package %s\n\n// V trả giá trị kiểm thử.\nfunc V() int { return 1 }\n' "$1" > "$d/$1.go"
  fi
  printf 'package %s\n\nimport "testing"\n\nfunc TestV(t *testing.T) {\n\tif V() < 1 {\n\t\tt.Fatal("V")\n\t}\n}\n' "$1" > "$d/$1_test.go"
}
mkpkg ci_a; mkpkg ci_b ci_a; mkpkg ci_c
(cd "$SERVER" && bazel run //:gazelle >/dev/null 2>&1)
t() { printf '%s\n' "$@" | scripts/ci/bazel-affected-tests.sh | tr '\n' ' '; }
out="$(t com/tm/server/pkg/ci_a/ci_a.go)"
check "đổi a → test a và b" "$(grep -q '//pkg/ci_a:ci_a_test' <<<"$out" && grep -q '//pkg/ci_b:ci_b_test' <<<"$out"; echo $?)" "$out"
check "đổi a → không có test c" "$(! grep -q ci_c <<<"$out"; echo $?)" "$out"
out="$(t com/tm/server/pkg/ci_c/ci_c.go)"
check "đổi c → chỉ test c" "$([[ "$out" == "//pkg/ci_c:ci_c_test " ]]; echo $?)" "$out"
out="$(t com/tm/server/pkg/ci_b/BUILD.bazel)"
check "đổi BUILD.bazel của b → test b" "$(grep -q '//pkg/ci_b:ci_b_test' <<<"$out"; echo $?)" "$out"
for g in MODULE.bazel go.mod .bazelrc; do
  check "đổi $g → //..." "$([[ "$(t "com/tm/server/$g")" == "//... " ]]; echo $?)"
done
check "đổi file ngoài com/tm/server → rỗng" "$([[ -z "$(t com/tm/app/package.json README.md)" ]]; echo $?)"

# ---------------------------------------------------------------------------
echo "TC04 — run.sh repo / server"
out="$(scripts/ci/run.sh repo 2>&1)"; check "run.sh repo pass" "$?" "$out"
out="$(scripts/ci/run.sh server 2>&1)"; code=$?
check "run.sh server pass (có package: lint + build + test)" "$code" "$out"
check "server đã chạy golangci-lint" "$(grep -q 'golangci-lint' <<<"$out" && ! grep -q 'Chưa có package Go' <<<"$out"; echo $?)" "$out"
check "server đã chạy test của 3 package" "$(grep -c 'PASSED' <<<"$out" | grep -qx 3; echo $?)" "$out"

# Lint phải fail khi code sai (tương đương P0-AT04 ở local)
printf 'package ci_c\n\nimport "os"\n\n// F bỏ qua lỗi.\nfunc F() { os.Remove("x") }\n' > "$SERVER/pkg/ci_c/lintbad.go"
(cd "$SERVER" && bazel run //:gazelle >/dev/null 2>&1)
out="$(scripts/ci/run.sh server 2>&1)"; code=$?
check "code vi phạm lint (errcheck) → server fail" "$([[ $code -ne 0 ]] && grep -q 'errcheck' <<<"$out"; echo $?)" "$out"
rm -f "$SERVER/pkg/ci_c/lintbad.go"; (cd "$SERVER" && bazel run //:gazelle >/dev/null 2>&1)

# BUILD.bazel lỗi thời → fail ở bước gazelle
mkdir -p "$SERVER/pkg/ci_stale" && printf 'package ci_stale\n\n// X kiểm thử.\nconst X = 1\n' > "$SERVER/pkg/ci_stale/x.go"
out="$(scripts/ci/run.sh server 2>&1)"; code=$?
check "BUILD.bazel lỗi thời → server fail ở bước gazelle" "$([[ $code -ne 0 ]] && grep -q 'BUILD.bazel lỗi thời' <<<"$out"; echo $?)" "$out"
cleanup

# ---------------------------------------------------------------------------
echo "TC05 — run.sh app"
out="$(scripts/ci/run.sh app 2>&1)"; check "run.sh app pass (không BASE)" "$?" "$out"
check "dùng --frozen-lockfile" "$(grep -q 'frozen-lockfile' <<<"$out"; echo $?)" "$out"
out="$(scripts/ci/run.sh app HEAD 2>&1)"; code=$?
check "BASE=HEAD, không package khớp → vẫn pass" "$code" "$out"
check "BASE → dùng --filter ...[BASE]" "$(grep -q -- '--filter \.\.\.\[HEAD\]' <<<"$out"; echo $?)" "$out"

echo
echo "Kết quả: $pass pass, $fail fail"
((fail == 0))
