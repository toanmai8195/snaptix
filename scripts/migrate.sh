#!/usr/bin/env bash
# Chạy goose migration cho PostgreSQL core.
#
#   scripts/migrate.sh up                       # áp dụng mọi migration chưa chạy
#   scripts/migrate.sh status                   # xem migration nào đã / chưa chạy
#   scripts/migrate.sh down                     # hoàn tác migration gần nhất
#   scripts/migrate.sh create add_trips sql     # tạo file migration mới (đánh số tuần tự)
#   scripts/migrate.sh --tool-version           # phiên bản goose đang pin
#
# Kết nối: CORE_DATABASE_URL, mặc định trỏ tới PG trong deploy/docker-compose.yml.
# goose được pin phiên bản và chạy bằng `go run` — không cần cài.

# -e: dừng khi một lệnh lỗi · -u: lỗi khi dùng biến chưa đặt · -o pipefail: pipe lỗi nếu bất kỳ lệnh nào lỗi
set -euo pipefail

GOOSE_VERSION="v3.26.0" # bản mới hơn đòi Go >= 1.26
# Chỉ build driver postgres (bỏ driver DB khác) để go run nhanh
GOOSE_TAGS="no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATIONS_DIR="$ROOT/com/tm/server/db/core/migrations"
DB_URL="${CORE_DATABASE_URL:-postgres://snaptix:snaptix@localhost:5432/core?sslmode=disable}"

goose() {
  # GOTOOLCHAIN=local: không để Go âm thầm tải toolchain khác
  GOTOOLCHAIN=local go run -tags "$GOOSE_TAGS" "github.com/pressly/goose/v3/cmd/goose@$GOOSE_VERSION" "$@"
}

usage() {
  echo "Dùng: $0 <up|down|status|create <ten> sql|...lệnh goose khác> | --tool-version" >&2
  exit 2
}

[[ $# -ge 1 ]] || usage

if [[ "$1" == "--tool-version" ]]; then
  goose -version
  exit 0
fi

# create: -s = đánh số tuần tự (00002_...) thay vì timestamp
extra=()
[[ "$1" == "create" ]] && extra=(-s)

goose -dir "$MIGRATIONS_DIR" "${extra[@]}" postgres "$DB_URL" "$@"
