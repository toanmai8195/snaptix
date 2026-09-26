#!/usr/bin/env bash
# Xác định vùng nào của monorepo bị ảnh hưởng bởi thay đổi.
#
#   scripts/ci/affected.sh <BASE_SHA>        # so sánh BASE...HEAD bằng git
#   git diff --name-only ... | scripts/ci/affected.sh -   # đọc danh sách file từ stdin
#
# In ra (định dạng $GITHUB_OUTPUT):
#   server=true|false   com/tm/server/** thay đổi
#   app=true|false      com/tm/app/** thay đổi
# Thay đổi cấu hình CI (.github/**, scripts/ci/**) → cả hai true.
# Không có BASE hợp lệ (vd push đầu tiên) → cả hai true.
set -euo pipefail

src="${1:--}"
if [[ "$src" == "-" ]]; then
  files="$(cat)"
elif [[ -z "$src" || "$src" =~ ^0+$ ]] || ! git cat-file -e "$src^{commit}" 2>/dev/null; then
  echo "server=true"; echo "app=true"; exit 0
else
  files="$(git diff --name-only "$src"...HEAD)"
fi

server=false app=false
while IFS= read -r f; do
  [[ -z "$f" ]] && continue
  case "$f" in
    com/tm/server/*) server=true ;;
    com/tm/app/*) app=true ;;
    .github/* | scripts/ci/*) server=true; app=true ;;
  esac
done <<<"$files"

echo "server=$server"
echo "app=$app"
