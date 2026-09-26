#!/usr/bin/env bash
# Chạy goose migration cho PG core hoặc PG analytics.
#
#   scripts/migrate.sh core up
#   scripts/migrate.sh analytics status
#   scripts/migrate.sh core create add_trips sql     # tạo migration mới (đánh số tuần tự)
#   scripts/migrate.sh --tool-version                # phiên bản goose đang pin
#
# Kết nối lấy từ CORE_DATABASE_URL / ANALYTICS_DATABASE_URL, mặc định theo deploy/docker-compose.yml.
# goose được pin phiên bản và chạy bằng `go run` — không cần cài global.
set -euo pipefail

GOOSE_VERSION="v3.26.0"
# Chỉ build driver postgres để go run nhanh và nhẹ
GOOSE_TAGS="no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATIONS="$ROOT/com/tm/server/db"

usage() { echo "Dùng: $0 <core|analytics> <lệnh goose...> | $0 --tool-version" >&2; exit 2; }

goose() { GOTOOLCHAIN=local go run -tags "$GOOSE_TAGS" "github.com/pressly/goose/v3/cmd/goose@$GOOSE_VERSION" "$@"; }

if [[ "${1:-}" == "--tool-version" ]]; then goose -version; exit; fi
[[ $# -ge 2 ]] || usage

db="$1"; shift
case "$db" in
  core)      url="${CORE_DATABASE_URL:-postgres://snaptix:snaptix@localhost:5432/core?sslmode=disable}" ;;
  analytics) url="${ANALYTICS_DATABASE_URL:-postgres://snaptix:snaptix@localhost:5433/analytics?sslmode=disable}" ;;
  *) usage ;;
esac

# Migration mới luôn đánh số tuần tự (-s), không dùng timestamp
seq_flag=()
[[ "$1" == "create" ]] && seq_flag=(-s)

goose -dir "$MIGRATIONS/$db/migrations" "${seq_flag[@]}" postgres "$url" "$@"
