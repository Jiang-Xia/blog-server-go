.PHONY: dev dev-all dev-all-windows dev-all-stop dev-all-status dev-all-logs dev-gateway dev-user dev-blog dev-rpg proto kitex ent-gen ent-gen-user ent-gen-blog ent-gen-rpg wire wire-user wire-blog wire-rpg build up down logs up-monolith down-monolith logs-monolith deploy deploy-monolith rollback rollback-list sync-pm2-config tidy \
	test-unit test-smoke test-integration test-e2e test-all test-coverage test-ci test-infra-up test-infra-down test-run \
	swag-apidoc swag-user swag-blog swag-rpg swag-gateway swag-all

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
COMPOSE_MONOLITH_FILE := deploy/docker/docker-compose.monolith.yml

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

dev-all-status:
	powershell -ExecutionPolicy Bypass -File scripts/dev-all-status.ps1

dev-all-logs:
	powershell -ExecutionPolicy Bypass -File scripts/dev-all-logs.ps1

dev-login:
	$(GO) run scripts/dev_login.go

dev-login-token:
	$(GO) run scripts/dev_login.go --token-only

# Kitex + protobuf 生成（需安装：go install github.com/cloudwego/kitex/tool/cmd/kitex@v0.16.3；本机 protoc）
# 输出：proto/kitex/{user,blog,rpg}/v1/；会短暂生成根目录 main.go/handler.go/build.sh，生成后删除
kitex:
	powershell -ExecutionPolicy Bypass -File scripts/gen-kitex.ps1

proto: kitex

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

# 学习：user/blog/rpg/gateway 各三实例（edge nginx 对外 :8000）
up-scale:
	docker compose -f $(COMPOSE_FILE) -f deploy/docker/docker-compose.scale.yml up -d --build \
		--scale user=3 --scale blog=3 --scale rpg=3 --scale gateway=3

# 兼容旧目标名
up-scale-user: up-scale

down:
	docker compose -f $(COMPOSE_FILE) down

down-scale:
	docker compose -f $(COMPOSE_FILE) -f deploy/docker/docker-compose.scale.yml down

logs:
	docker compose -f $(COMPOSE_FILE) logs -f

# 本地试验：单体 + MySQL + Redis + uniapp H5（见 deploy/docker/README.md）
up-monolith:
	docker compose -f $(COMPOSE_MONOLITH_FILE) up -d --build

down-monolith:
	docker compose -f $(COMPOSE_MONOLITH_FILE) down

logs-monolith:
	docker compose -f $(COMPOSE_MONOLITH_FILE) logs -f

deploy:
	powershell -ExecutionPolicy Bypass -File deploy/pm2/deploy.ps1

deploy-monolith:
	powershell -ExecutionPolicy Bypass -File deploy/pm2/deploy.ps1 -EnvFileName deploy.monolith.local.env

rollback:
	powershell -ExecutionPolicy Bypass -File deploy/pm2/rollback.ps1

rollback-list:
	powershell -ExecutionPolicy Bypass -File deploy/pm2/rollback.ps1 -List

sync-pm2-config:
	go run ./scripts/sync_pm2_config_from_nest.go --env deploy/pm2/env.production

bootstrap-prod-db:
	go run ./scripts/bootstrap_prod_db.go --env deploy/pm2/env.production

migrate-up:
	$(GO) run scripts/bootstrap_x_my_blog.go

migrate-down:
	@echo "local x_my_blog uses bootstrap clone; no migrate-down"

sync-data:
	$(GO) run scripts/sync_data_x_my_blog.go

tidy:
	$(GO) mod tidy

# Swagger：从 api-routes.md 生成 apidoc 桩，再 swag init 各微服务 docs/
swag-apidoc:
	$(GO) run scripts/gen_swag_apidoc.go

SWAG_DIRS = .,../../pkg/apidoc
SWAG_FLAGS = --parseDependency --parseInternal -d $(SWAG_DIRS)

swag-user: swag-apidoc
	cd services/user && swag init -g internal/apidoc/doc.go -o docs $(SWAG_FLAGS)

swag-blog: swag-apidoc
	cd services/blog && swag init -g internal/apidoc/doc.go -o docs $(SWAG_FLAGS)

swag-rpg: swag-apidoc
	cd services/rpg && swag init -g internal/apidoc/doc.go -o docs $(SWAG_FLAGS)

swag-gateway: swag-apidoc
	cd services/gateway && swag init -g internal/apidoc/doc.go -o docs $(SWAG_FLAGS)

swag-all: swag-user swag-blog swag-rpg swag-gateway

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
	@echo "optional: docker isolated test db (3307/6380); default test-run uses local 3306/6379"
	docker compose -f deploy/docker/docker-compose.test.yml up -d --wait

test-infra-down:
	docker compose -f deploy/docker/docker-compose.test.yml down -v

test-run:
	bash scripts/test-run.sh

test-ci: test-coverage
