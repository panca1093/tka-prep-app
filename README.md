# TKAPrep

Web-based platform for Indonesian students preparing for university entrance examinations (TKA, SMBT).

## Monorepo Layout

```
tkaprep/
├── apps/
│   ├── backend/                # Go API (chi + pgx + PostgreSQL)
│   └── frontend/               # Vue 3 + TypeScript + Vite
├── packages/
│   └── shared-types/           # OpenAPI spec → generated Go + TS clients
├── docker-compose.yml          # Local dev: postgres + backend + frontend
├── Makefile                    # One-stop orchestration
└── spec.md                     # Source-of-truth product spec
```

## Getting Started

### Prerequisites
- Docker + Docker Compose
- Go 1.22+
- Node.js 20+
- Make

### First-time setup
```bash
make install        # install backend + frontend deps
make generate       # generate types from openapi.yaml
make up             # start postgres + backend + frontend in docker
```

Backend: http://localhost:8080
Frontend: http://localhost:5173
Postgres: localhost:5432

### Health check
```bash
curl http://localhost:8080/health
# → {"status":"ok"}
```

## Daily Workflow

```bash
make up                 # bring everything up
make logs               # tail container logs
make down               # stop everything

make backend-dev        # run backend locally (hot reload via air)
make frontend-dev       # run frontend locally (hot reload via vite)

make generate           # regenerate types after openapi.yaml changes
make migrate-up         # apply DB migrations
make migrate-down       # rollback last migration

make test               # run all tests
make lint               # run all linters
```

## API Workflow (OpenAPI-First)

The single source of truth for the API is `packages/shared-types/openapi.yaml`.

1. Edit `openapi.yaml`
2. Run `make generate` — produces:
   - `packages/shared-types/generated/go/` — Go server interfaces & types
   - `packages/shared-types/generated/ts/` — TypeScript client types
3. Backend handlers implement the generated interfaces
4. Frontend uses the generated TS client

**Never write handler logic without first updating the spec.** Drift = bugs.

## Project Conventions

- All paths in `Makefile` and `docker-compose.yml` assume monorepo root as working directory
- Go module path: `github.com/yourorg/tkaprep/apps/backend` (find/replace `yourorg` with your real org)
- Migrations use `golang-migrate` numbering: `000001_description.up.sql` / `.down.sql`
- Commit messages: conventional commits (`feat:`, `fix:`, `chore:` …)

## Documentation

- `spec.md` — full product specification, source of truth
- `apps/backend/README.md` — backend architecture notes
- `apps/frontend/README.md` — frontend architecture notes
- `packages/shared-types/openapi.yaml` — API contract
