#!/usr/bin/env bash
# Unit test cho scripts/check-structure.sh và scripts/check-gitignore.sh (P0-T01).
#
#   scripts/test/check-scripts.test.sh
set -uo pipefail

SCRIPTS="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO="$(cd "$SCRIPTS/.." && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

pass=0 fail=0
check() { # check <tên> <điều kiện đúng: 0/1>
  if [[ "$2" == 0 ]]; then pass=$((pass + 1)); echo "  ✅ $1"; else fail=$((fail + 1)); echo "  ❌ $1"; fi
}

# Danh sách thư mục lấy từ chính script để fixture luôn khớp
DIRS=()
while IFS= read -r d; do DIRS+=("$d"); done < <(sed -n '/^REQUIRED_DIRS=(/,/^)/p' "$SCRIPTS/check-structure.sh" | sed '1d;$d' | tr -d ' ')

make_fixture() { # make_fixture <dir> [bỏ-qua...]
  local root="$1"; shift
  mkdir -p "$root"
  for d in "${DIRS[@]}"; do
    local skip=0
    for s in "$@"; do [[ "$d" == "$s" ]] && skip=1; done
    ((skip)) || mkdir -p "$root/$d"
  done
}

echo "check-structure.sh"

check "danh sách thư mục đọc được (17 mục)" "$([[ ${#DIRS[@]} -eq 17 ]]; echo $?)"

make_fixture "$TMP/full"
"$SCRIPTS/check-structure.sh" "$TMP/full" >/dev/null 2>&1
check "fixture đủ thư mục → exit 0" "$?"

make_fixture "$TMP/partial" com/tm/server/pkg loadtest
out="$("$SCRIPTS/check-structure.sh" "$TMP/partial" 2>&1)"; code=$?
check "thiếu thư mục → exit 1" "$([[ $code -eq 1 ]]; echo $?)"
check "in tên thư mục thiếu com/tm/server/pkg" "$(grep -q 'com/tm/server/pkg' <<<"$out"; echo $?)"
check "in tên thư mục thiếu loadtest" "$(grep -q 'loadtest' <<<"$out"; echo $?)"
check "không in thư mục đang có (deploy)" "$(! grep -q -- '- deploy' <<<"$out"; echo $?)"

mkdir -p "$TMP/empty"
out="$("$SCRIPTS/check-structure.sh" "$TMP/empty" 2>&1)"; code=$?
check "thư mục rỗng → exit 1, báo thiếu 17" "$([[ $code -eq 1 ]] && grep -q 'Thiếu 17' <<<"$out"; echo $?)"

touch "$TMP/full/deploy-file" && rm -rf "$TMP/full/deploy" && touch "$TMP/full/deploy"
"$SCRIPTS/check-structure.sh" "$TMP/full" >/dev/null 2>&1; code=$?
check "file trùng tên thư mục không được tính là thư mục" "$([[ $code -eq 1 ]]; echo $?)"

"$SCRIPTS/check-structure.sh" >/dev/null 2>&1
check "repo hiện tại (không truyền ROOT) → exit 0" "$?"

echo "check-gitignore.sh"

(cd "$REPO" && "$SCRIPTS/check-gitignore.sh" >/dev/null 2>&1)
check "repo hiện tại → exit 0" "$?"

# Repo tạm với .gitignore rỗng: phải fail và báo đường dẫn cần ignore
mkdir -p "$TMP/gi/scripts" && cp "$SCRIPTS/check-gitignore.sh" "$TMP/gi/scripts/"
(cd "$TMP/gi" && git init -q && : > .gitignore)
out="$("$TMP/gi/scripts/check-gitignore.sh" 2>&1)"; code=$?
check ".gitignore rỗng → exit 1" "$([[ $code -eq 1 ]]; echo $?)"
check ".gitignore rỗng → báo '.env' phải bị ignore" "$(grep -q 'phải bị ignore: .env' <<<"$out"; echo $?)"

# .gitignore ignore quá tay: go.mod bị ignore → phải báo
cp "$REPO/.gitignore" "$TMP/gi/.gitignore" && echo "go.mod" >> "$TMP/gi/.gitignore"
out="$("$TMP/gi/scripts/check-gitignore.sh" 2>&1)"; code=$?
check "ignore nhầm go.mod → exit 1 và báo" "$([[ $code -eq 1 ]] && grep -q 'không được ignore: com/tm/server/go.mod' <<<"$out"; echo $?)"

echo
echo "Kết quả: $pass pass, $fail fail"
((fail == 0))
