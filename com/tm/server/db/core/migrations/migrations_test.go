package migrations

import (
	"fmt"
	"io/fs"
	"regexp"
	"testing"
)

// Tên file goose đánh số tuần tự: NNNNN_ten_viet_thuong.sql (scripts/migrate.sh create ... dùng -s).
var fileName = regexp.MustCompile(`^(\d{5})_[a-z0-9_]+\.sql$`)

func TestFSContainsSequentialMigrations(t *testing.T) {
	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("FS rỗng: go:embed không nhúng được file .sql nào")
	}
	for i, e := range entries { // ReadDir trả theo tên đã sắp xếp
		m := fileName.FindStringSubmatch(e.Name())
		if m == nil {
			t.Errorf("%s: sai quy ước tên NNNNN_ten.sql", e.Name())
			continue
		}
		if want := fmt.Sprintf("%05d", i+1); m[1] != want {
			t.Errorf("%s: số thứ tự %s, muốn %s (không nhảy cóc, không trùng)", e.Name(), m[1], want)
		}
	}
}

func TestMigrationHasGooseAnnotations(t *testing.T) {
	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	up := regexp.MustCompile(`(?m)^-- \+goose Up$`)
	down := regexp.MustCompile(`(?m)^-- \+goose Down$`)
	for _, e := range entries {
		b, err := fs.ReadFile(FS, e.Name())
		if err != nil {
			t.Fatal(err)
		}
		if !up.Match(b) || !down.Match(b) {
			t.Errorf("%s: thiếu '-- +goose Up' hoặc '-- +goose Down'", e.Name())
		}
	}
}
