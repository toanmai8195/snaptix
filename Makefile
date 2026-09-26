# Thao tác thường dùng của snaptix. Chạy `make` để xem danh sách.
#
# PROJECT: compose project (mặc định snaptix). Dùng project khác để không đụng dữ liệu dev:
#   make up PROJECT=snaptix-test

PROJECT ?= snaptix
COMPOSE := docker compose -p $(PROJECT) -f deploy/docker-compose.yml
SERVER  := com/tm/server
APP     := com/tm/app

.DEFAULT_GOAL := help
.PHONY: help up down nuke ps logs migrate migrate-status test test-repo test-server test-app lint build gazelle

help: ## Danh sách target
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# --- Hạ tầng -----------------------------------------------------------------
up: ## Bật hạ tầng local (PG, Mongo, Redis, observability), chờ healthy
	$(COMPOSE) up -d --wait --wait-timeout 180

down: ## Dừng hạ tầng, giữ dữ liệu
	$(COMPOSE) down

nuke: ## Dừng hạ tầng và XOÁ toàn bộ dữ liệu (volume)
	$(COMPOSE) down -v --remove-orphans

ps: ## Trạng thái container
	$(COMPOSE) ps

logs: ## Theo dõi log (make logs s=postgres-core để lọc 1 service)
	$(COMPOSE) logs -f $(s)

# --- Database ------------------------------------------------------------------
migrate: ## Chạy migration cho PG core và PG analytics
	scripts/migrate.sh core up
	scripts/migrate.sh analytics up

migrate-status: ## Trạng thái migration
	scripts/migrate.sh core status
	scripts/migrate.sh analytics status

# --- Build / test (giống CI: scripts/ci/run.sh) --------------------------------
test: test-repo test-server test-app ## Toàn bộ kiểm tra như CI

test-repo: ## Kiểm tra repo (cấu trúc, .gitignore, migration)
	scripts/ci/run.sh repo

test-server: ## Go: gazelle diff, lint, bazel build, bazel test
	scripts/ci/run.sh server

test-app: ## Node/React: install, lint, test, build
	scripts/ci/run.sh app

lint: ## Lint Go và TS
	cd $(SERVER) && if [ -n "$$(GOTOOLCHAIN=local go list ./... 2>/dev/null)" ]; then \
		GOTOOLCHAIN=local go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2 run ./...; \
	else echo "Chưa có package Go"; fi
	cd $(APP) && pnpm lint

build: ## Build Go (Bazel) và TS
	cd $(SERVER) && bazel build //...
	cd $(APP) && pnpm build

gazelle: ## Cập nhật BUILD.bazel cho Go
	cd $(SERVER) && bazel run //:gazelle
