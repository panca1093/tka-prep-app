# TKAPrep Backend

Go API server. Chi router, pgx for Postgres, JWT for auth.

## Architecture

```
cmd/server/                 # entrypoint
internal/
├── config/                 # env loader
├── handler/http/           # HTTP handlers (translate HTTP ↔ services)
├── handler/middleware/     # auth, RBAC, logging middleware
├── pkg/server/             # router setup, middleware wiring
├── repository/             # data access (interfaces + postgres impl)
│   └── postgres/
├── service/                # business logic (orchestrates domain + repos)
└── domain/                 # pure entities, no external deps
    ├── user/  question/  test/  session/  result/
migrations/                 # golang-migrate SQL files
```

### Layer rules
- **domain** imports nothing from other internal packages
- **repository** depends on domain only
- **service** depends on domain + repository interfaces
- **handler** depends on service only — no direct DB calls

## Running locally

```bash
cp .env.example .env       # adjust as needed
go run ./cmd/server
```

With hot reload:
```bash
go install github.com/air-verse/air@latest
air
```

## OpenAPI generation

Server stubs are generated from `../../packages/shared-types/openapi.yaml` via `oapi-codegen`.

```bash
make generate-go    # from repo root
```

Output goes to `packages/shared-types/generated/go/api.gen.go`.

## Migrations

Uses `golang-migrate`:
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
make migrate-up
make migrate-create name=add_users_table
```
