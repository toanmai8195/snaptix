-- Migration đầu tiên của PG core: chỉ là mốc version 1 để kiểm chứng cơ chế migration.
-- Bảng nghiệp vụ (stations, routes, trips...) bắt đầu từ Phase 1.

-- +goose Up
SELECT 'snaptix core: init';

-- +goose Down
SELECT 'snaptix core: init rollback';
