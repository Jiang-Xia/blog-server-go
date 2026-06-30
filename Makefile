.PHONY: dev ent-gen wire migrate-up migrate-down tidy ent-schema bootstrap-db

GO ?= go
CONFIG_PATH ?= configs/monolith.yaml
ENT_DIR := services/monolith/ent
APP_DIR := services/monolith/internal/app

dev:
	set CONFIG_PATH=$(CONFIG_PATH)&& $(GO) run ./services/monolith/cmd/main.go

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

migrate-up:
	$(GO) run scripts/bootstrap_x_my_blog.go

migrate-down:
	@echo "local x_my_blog uses bootstrap clone; no migrate-down"

tidy:
	$(GO) mod tidy
