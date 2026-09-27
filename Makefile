# Lệnh thường dùng của snaptix. Chỉ là "mục lục" gọi công cụ có sẵn — logic nằm ở script/compose.
#
#   make            # = make help: liệt kê target
#   make up         # PG core (chờ healthy)
#   make migrate    # goose up cho PG core
#
# Recipe phải thụt bằng TAB. GNU Make trên macOS là 3.81 → chỉ dùng cú pháp có từ bản đó.

# Thư mục chứa Makefile (kết thúc bằng /): `make -C <repo>` hay gọi từ thư mục khác vẫn đúng đường dẫn.
ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
COMPOSE := docker compose -f $(ROOT)deploy/docker-compose.yml
SERVER := $(ROOT)com/tm/server

# Target không phải tên file: có file tên "test" trong thư mục thì `make test` vẫn chạy.
.PHONY: help up down migrate test gazelle
.DEFAULT_GOAL := help

help: ## Liệt kê target
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  %-10s %s\n", $$1, $$2}'

up: ## Khởi động hạ tầng local (PG core), chờ tới khi healthy
	$(COMPOSE) up -d --wait

down: ## Dừng hạ tầng local, GIỮ dữ liệu (xoá dữ liệu: docker compose ... down -v)
	$(COMPOSE) down

migrate: ## Áp dụng migration PG core (goose up)
	$(ROOT)scripts/migrate.sh up

# Mỗi dòng recipe chạy trong một shell riêng → `cd` phải cùng dòng với lệnh.
test: ## go vet + go test -race cho server (integration test cần Docker, tự skip nếu không có)
	cd $(SERVER) && go vet ./...
	cd $(SERVER) && go test -race ./...

gazelle: ## Sinh / cập nhật BUILD.bazel sau khi thêm, xoá file Go hoặc đổi import
	cd $(SERVER) && bazel run //:gazelle
