.PHONY: dev dev-all dev-all-stop dev-gateway dev-user dev-blog dev-rpg proto ent-gen wire build up down logs tidy

GO ?= go
CONFIG_PATH ?= configs/monolith.yaml
ENT_DIR := services/monolith/ent
APP_DIR := services/monolith/internal/app
COMPOSE_FILE := deploy/docker/docker-compose.yml

dev:
	set CONFIG_PATH=$(CONFIG_PATH)&& $(GO) run ./services/monolith/cmd/main.go

dev-gateway:
	set CONFIG_PATH=configs/gateway.yaml&& $(GO) run ./services/gateway/cmd/main.go

dev-user:
	set CONFIG_PATH=configs/user.yaml&& $(GO) run ./services/monolith/cmd/user/main.go

dev-blog:
	set CONFIG_PATH=configs/blog.yaml&& $(GO) run ./services/monolith/cmd/blog/main.go

dev-rpg:
	set CONFIG_PATH=configs/rpg.yaml&& $(GO) run ./services/monolith/cmd/rpg/main.go

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

ent-schema:
	$(GO) run scripts/gen_ent_schema.go

ent-schema-gen: ent-schema ent-gen

bootstrap-db:
	$(GO) run scripts/bootstrap_x_my_blog.go
	$(GO) run scripts/gen_ent_schema.go
	cd $(ENT_DIR) && $(GO) generate ./...

wire:
	$(GO) run github.com/google/wire/cmd/wire@latest ./$(APP_DIR)

build:
	@CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o bin/gateway ./services/gateway/cmd
	@CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o bin/monolith ./services/monolith/cmd
	@CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o bin/user ./services/monolith/cmd/user
	@CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o bin/blog ./services/monolith/cmd/blog
	@CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o bin/rpg ./services/monolith/cmd/rpg

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
