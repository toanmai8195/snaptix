package integration

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/toanmai8195/snaptix/com/tm/server/services/core/internal/pgtest"
)

// TC03: test chạy song song, mỗi test một database riêng trên CÙNG một container.
func TestEachTestGetsOwnDatabase(t *testing.T) {
	pgtest.RequireDocker(t) // -short / không Docker → skip
	shared, err := pgtest.Shared()
	if err != nil {
		t.Fatal(err)
	}
	sharedDSN, err := shared.DSN(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"a", "b"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := pgtest.NewDB(t)

			// Cùng host:port với container dùng chung → không bật container mới cho mỗi test.
			if cfg := pool.Config().ConnConfig; !sameServer(cfg.Host, cfg.Port, sharedDSN) {
				t.Errorf("DB test ở %s:%d, không phải container dùng chung %s", cfg.Host, cfg.Port, sharedDSN)
			}

			if _, err := pool.Exec(ctx, "CREATE TABLE note (owner text)"); err != nil {
				t.Fatalf("create table: %v", err) // trùng tên bảng sẽ lỗi nếu hai test chung DB
			}
			if _, err := pool.Exec(ctx, "INSERT INTO note VALUES ($1)", name); err != nil {
				t.Fatal(err)
			}
			var owners []string
			rows, err := pool.Query(ctx, "SELECT owner FROM note")
			if err != nil {
				t.Fatal(err)
			}
			for rows.Next() {
				var o string
				if err := rows.Scan(&o); err != nil {
					t.Fatal(err)
				}
				owners = append(owners, o)
			}
			if rows.Err() != nil {
				t.Fatal(rows.Err())
			}
			if len(owners) != 1 || owners[0] != name {
				t.Errorf("thấy %v, muốn chỉ [%s] — dữ liệu của test khác lọt sang", owners, name)
			}
		})
	}
}

func sameServer(host string, port uint16, dsn string) bool {
	cfg, err := pgconn.ParseConfig(dsn)
	return err == nil && cfg.Host == host && cfg.Port == port
}
