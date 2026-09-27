// Package migrations nhúng các file migration goose của PG core vào code.
//
// scripts/migrate.sh đọc thẳng thư mục này; test (pgtest) dùng FS nhúng
// nên chạy giống nhau bằng `go test` lẫn `bazel test`, không phụ thuộc thư mục làm việc.
package migrations

import "embed"

// FS chứa mọi file NNNNN_ten.sql trong thư mục này.
//
//go:embed *.sql
var FS embed.FS
