package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/toanmai8195/snaptix/com/tm/server/services/core/internal/pgtest"
)

// TC01: PG đúng phiên bản của compose.
func TestPostgresVersion(t *testing.T) {
	pool := pgtest.NewDB(t)

	var version string
	if err := pool.QueryRow(context.Background(), "SELECT version()").Scan(&version); err != nil {
		t.Fatalf("SELECT version(): %v", err)
	}
	if !strings.HasPrefix(version, "PostgreSQL 17.6") {
		t.Errorf("version = %q, muốn PostgreSQL 17.6 (image %s)", version, pgtest.Image)
	}
}

// TC02: DB của test đã có migration mới nhất (copy từ template).
func TestMigrationsApplied(t *testing.T) {
	pool := pgtest.NewDB(t)

	var version int64
	err := pool.QueryRow(context.Background(),
		"SELECT max(version_id) FROM goose_db_version WHERE is_applied").Scan(&version)
	if err != nil {
		t.Fatalf("đọc goose_db_version: %v", err)
	}
	if version != 1 {
		t.Errorf("migration version = %d, muốn 1 (00001_init.sql)", version)
	}
}
