# com/tm/server

Toàn bộ code **Go** của snaptix: core service, stats-worker, thư viện dùng chung, migration PostgreSQL. Build bằng Bazel, dev hằng ngày bằng `go`.

Thư mục con được tạo **ở task dùng tới chúng**, không tạo sẵn:

| Thư mục | Chứa gì | Xuất hiện ở |
|---|---|---|
| `db/core/migrations/` | Migration goose cho PG core | P0-T03 |
| `go.mod`, `MODULE.bazel`, `BUILD.bazel` | Go module + cấu hình Bazel | P0-T04 |
| `services/core/` | Core service (HTTP API, nguồn sự thật cho ghế, vé, ví) | P0-T05 |
| `tools/rules/` | Macro Bazel build image OCI | P0-T07 |
| `pkg/` | Code dùng chung giữa các service (vd kết nối PG) | khi service thứ hai cần |
| `services/stats-worker/`, `db/analytics/` | Thống kê | Phase 6 |

Quy ước tổ chức code: xem [project-structure](../docs/technical/project-structure.md#comtmserver--go--bazel).
