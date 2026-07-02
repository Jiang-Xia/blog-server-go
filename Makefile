.PHONY: dev dev-all dev-all-stop dev-gateway dev-user dev-blog dev-rpg proto ent-gen ent-gen-user ent-gen-blog ent-gen-rpg wire wire-user wire-blog wire-rpg build up down logs tidy \
	test-unit test-smoke test-integration test-e2e test-all test-coverage test-ci test-infra-up test-infra-down test-run

GO ?= go
CONFIG_PATH ?= configs/monolith.yaml
ENT_DIR := services/monolith/ent
USER_ENT_DIR := services/user/ent
BLOG_ENT_DIR := services/blog/ent
RPG_ENT_DIR := services/rpg/ent
APP_DIR := services/monolith/internal/app
USER_APP_DIR := services/user/internal/app
BLOG_APP_DIR := services/blog/internal/app
RPG_APP_DIR := services/rpg/internal/app
COMPOSE_FILE := deploy/docker/docker-compose.yml

dev:
	set CONFIG_PATH=$(CONFIG_PATH)&& $(GO) run ./services/monolith/cmd/main.go

dev-gateway:
	set CONFIG_PATH=configs/gateway.yaml&& $(GO) run ./services/gateway/cmd/main.go

dev-user:
	set CONFIG_PATH=configs/user.yaml&& $(GO) run ./services/user/cmd/main.go

dev-blog:
	set CONFIG_PATH=configs/blog.yaml&& $(GO) run ./services/blog/cmd/main.go

dev-rpg:
	set CONFIG_PATH=configs/rpg.yaml&& $(GO) run ./services/rpg/cmd/main.go

dev-all:
	powershell -ExecutionPolicy Bypass -File scripts/dev-all.ps1

dev-all-windows:
	powershell -ExecutionPolicy Bypass -File scripts/dev-all.ps1 -Windows

dev-all-stop:
	powershell -ExecutionPolicy Bypass -File scripts/dev-all-stop.ps1

dev-login:
	$(GO) run scripts/dev_login.go

dev-login-token:
	$(GO) run scripts/dev_login.go --token-only

proto:
	buf generate

ent-gen:
	cd $(ENT_DIR) && $(GO) generate ./...

ent-gen-user:
	cd $(USER_ENT_DIR) && $(GO) generate ./...

ent-gen-blog:
	cd $(BLOG_ENT_DIR) && $(GO) generate ./...

ent-gen-rpg:
	cd $(RPG_ENT_DIR) && $(GO) generate ./...

ent-schema:
	$(GO) run scripts/gen_ent_schema.go

ent-schema-gen: ent-schema ent-gen

bootstrap-db:
	$(GO) run scripts/bootstrap_x_my_blog.go
	$(GO) run scripts/gen_ent_schema.go
	cd $(ENT_DIR) && $(GO) generate ./...

wire:
	$(GO) run github.com/google/wire/cmd/wire@latest ./$(APP_DIR)

wire-user:
	$(GO) run github.com/google/wire/cmd/wire@latest ./$(USER_APP_DIR)

wire-blog:
	$(GO) run github.com/google/wire/cmd/wire@latest ./$(BLOG_APP_DIR)

wire-rpg:
	$(GO) run github.com/google/wire/cmd/wire@latest ./$(RPG_APP_DIR)

build:
	@CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o bin/gateway ./services/gateway/cmd
	@CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o bin/monolith ./services/monolith/cmd
	@CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o bin/user ./services/user/cmd
	@CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o bin/blog ./services/blog/cmd
	@CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o bin/rpg ./services/rpg/cmd

up:
	docker compose -f $(COMPOSE_FILE) up -d --build

down:
	docker compose -f $(COMPOSE_FILE) down

logs:
	docker compose -f $(COMPOSE_FILE) logs -f

migrate-up:
	$(GO) run scripts/bootstrap_x_my_blog.go

migrate-down:
	@echo "local x_my_blog uses bootstrap clone; no migrate-down"

sync-data:
	$(GO) run scripts/sync_data_x_my_blog.go

tidy:
	$(GO) mod tidy

test-unit:
	$(GO) test ./pkg/... -count=1

test-coverage:
	bash scripts/ci/check-coverage.sh

test-smoke:
	$(GO) test -tags=smoke ./test/smoke/... -count=1 -v

test-integration:
	$(GO) test -tags=integration ./test/integration/... -count=1 -v

test-e2e:
	$(GO) test -tags=e2e ./test/e2e/... -count=1 -v

test-all: test-coverage test-smoke test-integration test-e2e

test-infra-up:
	docker compose -f deploy/docker/docker-compose.test.yml up -d --wait

test-infra-down:
	docker compose -f deploy/docker/docker-compose.test.yml down -v

test-run:
	bash scripts/test-run.sh

test-ci: test-coverage
