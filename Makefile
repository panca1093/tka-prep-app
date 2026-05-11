.PHONY: help install up down logs ps clean \
	backend-dev frontend-dev \
	generate generate-go generate-ts \
	migrate-up migrate-down migrate-create \
	test test-backend test-frontend \
	lint lint-backend lint-frontend

# Default
help:
	@echo "TKAPrep — Common commands"
	@echo ""
	@echo "  make install         Install backend + frontend dependencies"
	@echo "  make up              Start everything in docker (pg + backend + frontend)"
	@echo "  make down            Stop everything"
	@echo "  make logs            Tail container logs"
	@echo "  make ps              List running containers"
	@echo ""
	@echo "  make backend-dev     Run backend locally with hot reload (air)"
	@echo "  make frontend-dev    Run frontend locally with hot reload (vite)"
	@echo ""
	@echo "  make generate        Regenerate Go + TS types from openapi.yaml"
	@echo "  make migrate-up      Apply pending DB migrations"
	@echo "  make migrate-down    Rollback last migration"
	@echo "  make migrate-create name=add_users   Create new migration files"
	@echo ""
	@echo "  make test            Run all tests"
	@echo "  make lint            Run all linters"
	@echo "  make clean           Remove generated files and node_modules"

# ─── Setup ─────────────────────────────────────────────────────────────────

install:
	cd apps/backend && go mod download
	cd apps/frontend && npm install
	cd packages/shared-types && npm install

# ─── Docker orchestration ─────────────────────────────────────────────────

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

# ─── Dev (run locally, not in docker) ─────────────────────────────────────

backend-dev:
	cd apps/backend && air

frontend-dev:
	cd apps/frontend && npm run dev

# ─── OpenAPI codegen ───────────────────────────────────────────────────────

generate: generate-go generate-ts

generate-go:
	@echo "→ Generating Go server types from openapi.yaml"
	cd apps/backend && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		-config ../../packages/shared-types/tools/oapi-codegen.yaml \
		../../packages/shared-types/openapi.yaml

generate-ts:
	@echo "→ Generating TypeScript types from openapi.yaml"
	cd packages/shared-types && npm run generate:ts

# ─── Database migrations ───────────────────────────────────────────────────

DB_URL ?= postgres://tkaprep:tkaprep@localhost:5432/tkaprep?sslmode=disable
MIGRATIONS_DIR := apps/backend/migrations

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=description"; exit 1; fi
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# ─── Tests ─────────────────────────────────────────────────────────────────

test: test-backend test-frontend

test-backend:
	cd apps/backend && go test ./...

test-frontend:
	cd apps/frontend && npm run test

# ─── Lint ──────────────────────────────────────────────────────────────────

lint: lint-backend lint-frontend

lint-backend:
	cd apps/backend && go vet ./...
	cd apps/backend && gofmt -l .

lint-frontend:
	cd apps/frontend && npm run lint

# ─── Cleanup ───────────────────────────────────────────────────────────────

clean:
	rm -rf apps/frontend/node_modules
	rm -rf packages/shared-types/node_modules
	rm -rf packages/shared-types/generated/go/*.go
	rm -rf packages/shared-types/generated/ts/*.ts
