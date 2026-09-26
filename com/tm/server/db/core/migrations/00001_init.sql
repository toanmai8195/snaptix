-- Migration khởi tạo PG core: đánh dấu mốc version 1, schema nghiệp vụ bắt đầu từ Phase 1.

-- +goose Up
SELECT 'snaptix core: init';

-- +goose Down
SELECT 'snaptix core: init rollback';
