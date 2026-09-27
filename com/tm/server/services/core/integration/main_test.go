// Package integration chứa integration test của core với hạ tầng thật (PG trong Docker).
// Target Bazel riêng gắn tag requires-docker — xem BUILD.bazel.
package integration

import (
	"os"
	"testing"

	"github.com/toanmai8195/snaptix/com/tm/server/services/core/internal/pgtest"
)

// Một container PG cho cả package; -short hoặc không có Docker → mọi test skip.
func TestMain(m *testing.M) {
	os.Exit(pgtest.Run(m))
}
