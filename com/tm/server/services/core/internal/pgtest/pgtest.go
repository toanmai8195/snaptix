// Package pgtest dựng PostgreSQL thật (testcontainers) cho integration test của core.
//
// Cách dùng trong một package test:
//
//	func TestMain(m *testing.M) { os.Exit(pgtest.Run(m)) }
//
//	func TestX(t *testing.T) {
//		pool := pgtest.NewDB(t) // database riêng, đã chạy migration, tự xoá khi test xong
//		...
//	}
//
// Cả package dùng MỘT container. Migration chạy một lần vào database template;
// mỗi test nhận bản sao bằng CREATE DATABASE ... TEMPLATE — nhanh và cô lập.
//
// Mới chỉ core dùng nên nằm ở internal/; service thứ hai cần thì chuyển lên pkg/.
package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // driver "pgx" cho database/sql — goose cần *sql.DB
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/toanmai8195/snaptix/com/tm/server/db/core/migrations"
)

// Image cùng deploy/docker-compose.yml: test chạy đúng phiên bản PG của môi trường thật.
const Image = "postgres:17.6-alpine"

// templateDB chứa schema đã migrate; không test nào kết nối thẳng vào nó.
const templateDB = "core_template"

// Container là một PG trong Docker, đã có database template chạy xong migration.
type Container struct {
	ctr     *postgres.PostgresContainer
	baseCfg *pgxpool.Config // trỏ tới templateDB; đổi Database khi mở DB của từng test
	admin   *pgxpool.Pool   // nối vào DB "postgres" để CREATE / DROP DATABASE
	seq     atomic.Int64    // đánh số DB test; atomic vì test chạy song song
}

// Start khởi động container, chạy migration vào database template.
// opts thêm tuỳ chọn cho container (vd FixedHostPort).
func Start(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	opts = append([]testcontainers.ContainerCustomizer{
		postgres.WithDatabase(templateDB),
		postgres.WithUsername("snaptix"),
		postgres.WithPassword("snaptix"),
		// Chờ PG sẵn sàng thật: log "ready" 2 lần (PG khởi động tạm rồi restart) + cổng đã map ra host.
		postgres.BasicWaitStrategies(),
	}, opts...)
	ctr, err := postgres.Run(ctx, Image, opts...)
	c := &Container{ctr: ctr}
	if err != nil {
		_ = c.Terminate(ctx) // Run lỗi giữa chừng vẫn có thể đã tạo container
		return nil, fmt.Errorf("start postgres: %w", err)
	}
	if err := c.init(ctx); err != nil {
		_ = c.Terminate(ctx)
		return nil, err
	}
	return c, nil
}

func (c *Container) init(ctx context.Context) error {
	// Cổng 5432 trong container được map ra cổng ngẫu nhiên trên host → hỏi lại connection string.
	dsn, err := c.ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return err
	}
	if err := migrate(ctx, dsn); err != nil {
		return err
	}

	if c.baseCfg, err = pgxpool.ParseConfig(dsn); err != nil {
		return err
	}
	adminCfg := c.baseCfg.Copy()
	adminCfg.ConnConfig.Database = "postgres"
	c.admin, err = pgxpool.NewWithConfig(ctx, adminCfg)
	return err
}

// migrate chạy goose (dùng như thư viện) với migration nhúng trong binary.
func migrate(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	// Đóng ngay khi xong: PG không cho dùng một database làm template khi còn kết nối vào nó.
	defer func() { _ = db.Close() }()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		return fmt.Errorf("goose provider: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

// NewDB tạo database mới từ template cho riêng test t; tự đóng pool và xoá DB khi test xong.
func (c *Container) NewDB(t testing.TB) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	name := fmt.Sprintf("test_%d", c.seq.Add(1))
	// Tên DB là identifier, không truyền bằng tham số $1 được → Sanitize để trích dẫn an toàn.
	ident := pgx.Identifier{name}.Sanitize()
	if _, err := c.admin.Exec(ctx, "CREATE DATABASE "+ident+" TEMPLATE "+templateDB); err != nil {
		t.Fatalf("create database %s: %v", name, err)
	}

	cfg := c.baseCfg.Copy()
	cfg.ConnConfig.Database = name
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("connect %s: %v", name, err)
	}

	t.Cleanup(func() {
		pool.Close()
		// WITH (FORCE): cắt kết nối còn sót (vd test quên đóng) thay vì lỗi "being accessed by other users".
		// Thử lại vài lần: test vừa restart PG thì kết nối cũ trong pool admin đã chết,
		// lần Exec đầu dùng đúng kết nối đó sẽ lỗi; pool bỏ nó và mở kết nối mới ở lần sau.
		var err error
		for range 3 {
			if _, err = c.admin.Exec(context.Background(), "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err == nil {
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
		t.Errorf("drop database %s: %v", name, err)
	})
	return pool
}

// FixedHostPort map 5432 của container ra một cổng host cố định (chọn cổng trống lúc gọi).
// Mặc định Docker cấp cổng host ngẫu nhiên MỖI lần container start → Stop/Start làm
// DSN cũ trỏ sai cổng. Test cần dừng rồi bật lại PG mà giữ nguyên kết nối thì dùng option này.
func FixedHostPort() (testcontainers.ContainerCustomizer, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0") // hỏi OS một cổng trống rồi trả lại ngay
	if err != nil {
		return nil, err
	}
	port := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	if err := ln.Close(); err != nil {
		return nil, err
	}
	return testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
		hc.PortBindings = nat.PortMap{
			"5432/tcp": []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: port}},
		}
	}), nil
}

// DSN của database template — cho test cần tự mở kết nối (không ghi vào template).
func (c *Container) DSN(ctx context.Context) (string, error) {
	return c.ctr.ConnectionString(ctx, "sslmode=disable")
}

// Testcontainer trả container gốc cho test cần điều khiển hạ tầng (vd Stop/Start PG).
func (c *Container) Testcontainer() *postgres.PostgresContainer {
	return c.ctr
}

// Terminate đóng kết nối admin và xoá container.
func (c *Container) Terminate(ctx context.Context) error {
	if c.admin != nil {
		c.admin.Close()
	}
	if c.ctr == nil {
		return nil
	}
	return c.ctr.Terminate(ctx)
}

// ---- Container dùng chung cho cả package test ----

var (
	shared     *Container
	skipReason string // khác rỗng → NewDB skip test với lý do này
)

// Run dùng trong TestMain: khởi động container dùng chung (trừ khi -short hoặc không có Docker),
// chạy test, rồi xoá container. Trả exit code cho os.Exit.
func Run(m *testing.M) int {
	flag.Parse() // testing.Short() cần flag đã parse; m.Run() không parse lại
	switch {
	case testing.Short():
		skipReason = "integration test: bỏ qua khi chạy -short"
	case !dockerAvailable():
		// CI đặt PGTEST_REQUIRE_DOCKER=1: không có Docker là lỗi môi trường, không được lặng lẽ skip
		// (CI xanh phải nghĩa là integration test đã chạy thật).
		if os.Getenv("PGTEST_REQUIRE_DOCKER") != "" {
			log.Print("pgtest: PGTEST_REQUIRE_DOCKER đặt nhưng không kết nối được Docker")
			return 1
		}
		skipReason = "integration test: không có Docker"
	default:
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		var err error
		shared, err = Start(ctx)
		cancel()
		if err != nil {
			log.Printf("pgtest: %v", err)
			return 1
		}
		defer func() { _ = shared.Terminate(context.Background()) }()
	}
	return m.Run()
}

// NewDB tạo database riêng trên container dùng chung; skip test nếu container không có.
func NewDB(t testing.TB) *pgxpool.Pool {
	t.Helper()
	if shared == nil {
		if skipReason == "" {
			t.Fatal("pgtest: chưa gọi pgtest.Run trong TestMain")
		}
		t.Skip(skipReason)
	}
	return shared.NewDB(t)
}

// RequireDocker skip test khi -short hoặc không có Docker — cho test tự dựng container riêng.
func RequireDocker(t testing.TB) {
	t.Helper()
	if skipReason != "" {
		t.Skip(skipReason)
	}
}

// dockerAvailable: testcontainers tìm được Docker và Docker trả lời health check.
func dockerAvailable() (ok bool) {
	defer func() {
		if recover() != nil { // testcontainers có thể panic khi không tìm thấy Docker
			ok = false
		}
	}()
	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		return false
	}
	defer func() { _ = provider.Close() }()
	return provider.Health(context.Background()) == nil
}

// ErrNoContainer: dùng Shared khi container dùng chung chưa được khởi động.
var ErrNoContainer = errors.New("pgtest: container dùng chung chưa khởi động")

// Shared trả container dùng chung (sau pgtest.Run), cho test cần DSN hoặc điều khiển container.
func Shared() (*Container, error) {
	if shared == nil {
		return nil, ErrNoContainer
	}
	return shared, nil
}
