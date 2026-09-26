-- Migration khởi tạo PG analytics: đánh dấu mốc version 1, schema phân tích bắt đầu từ Phase 6.

-- +goose Up
SELECT 'snaptix analytics: init';

-- +goose Down
SELECT 'snaptix analytics: init rollback';
