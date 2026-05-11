# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

It captures the project's conventions, architectural decisions, and the reasoning behind them so any future work stays consistent.

**Always read `spec.md` first.** It is the source of truth for what we're building.

---

## Product One-Liner

Web-based platform for Indonesian students preparing for university entrance examinations (TKA Saintek, TKA Soshum, SMBT). Three roles: Student, Contributor, Super Admin.

## Tech Stack (Locked)

| Layer | Choice |
|---|---|
| Frontend | Vue 3 + TypeScript + Vite + Pinia + vue-router |
| Backend | Go 1.22 + chi + pgx + zerolog |
| Database | PostgreSQL 16 |
| Auth | JWT (HS256), access (15 min) + refresh (7 d) with rotation |
| API workflow | OpenAPI-first, code generated from `packages/shared-types/openapi.yaml` |
| Local dev | docker-compose + Makefile |

## Monorepo Layout

```
tkaprep/
├── apps/
│   ├── backend/        Go API
│   └── frontend/       Vue SPA
├── packages/
│   └── shared-types/   OpenAPI spec + generated Go + TS clients
├── spec.md             Product spec (source of truth)
├── CLAUDE.md           This file
└── docker-compose.yml  Local dev orchestration
```

---

## Architectural Decisions

### Auth strategy: Email-first, OAuth-ready
- Ship email + password in v1
- `users` schema is forward-compatible: `password_hash` is nullable, `auth_provider` enum exists (`email` only in v1, `google` later)
- No social login in v1 — adding Google OAuth is a v2 task
- Access token short (15 min) so a stolen token has limited blast radius
- Refresh token in DB so we can revoke; rotation on every refresh detects theft

### OpenAPI-first
- `packages/shared-types/openapi.yaml` is the **only** place API contracts live
- Run `make generate` after every spec change → produces:
  - `packages/shared-types/generated/go/api.gen.go` (oapi-codegen, chi server interfaces)
  - `packages/shared-types/generated/ts/index.ts` (openapi-typescript)
- **PRs that change handlers without updating the spec must be rejected**
- Frontend `src/api/client.ts` is hand-written for now; replace with generated client once auth endpoints land

### Backend architecture: clean layers
```
domain ← repository ← service ← handler
```
- `internal/domain/` — pure entities, no external deps, no imports from other internal packages
- `internal/repository/` — `interfaces.go` defines contracts, `postgres/` implements them
- `internal/service/` — all business logic; orchestrates domain + repos
- `internal/handler/http/` — HTTP-only; no business logic, no direct DB calls

### Leaderboard
- **No `leaderboard_snapshots` table.** Rank is computed live via SQL `RANK() OVER (ORDER BY total_score DESC)`.
- Top-100 leaderboard uses `LIMIT 100`
- Personal rank for users outside top 100 uses a separate query
- Required index: `test_results(total_score DESC)` — exists from migration #1 onwards (will add when we create the table)

### Question ownership
- All questions are shared across all contributors in v1
- `contributor_id` is tracked for accountability but does not restrict use
- Visibility flag (`private` / `public`) is **deferred** to v2 (see spec.md §12)
- Test visibility (per-class, per-group) also deferred to v2

### Scoring formula
- Stored per test in `scoring_configs` (1:1 with `tests`)
- Three fields: `correct_points`, `wrong_points`, `blank_points` (decimal, can be negative)
- Formula: `score = (correct_count × correct_points) + (wrong_count × wrong_points) + (blank_count × blank_points)`
- Applied **server-side only**, on submit, inside a DB transaction
- Result is immutable once computed

### Session integrity
- Timer is **server-authoritative**. Client time is never trusted.
- Every API call to a session checks `now() - started_at <= duration_minutes`
- Submit is wrapped in a DB transaction: update session status, compute result, insert into `test_results`
- Sessions can be resumed if `status = in_progress` and within duration

---

## Coding Conventions

### Go
- Use `gofmt` + `go vet` before committing
- Errors: wrap with `fmt.Errorf("context: %w", err)`, never `errors.New` for known error types
- Logging: structured via `zerolog`, no `fmt.Println` outside of debugging
- Handlers must NOT contain business logic — delegate to services
- Repos return `domain.X` types, never raw `pgx.Rows`
- Context: every function that does I/O takes `context.Context` as first arg

### TypeScript / Vue
- Use Composition API + `<script setup lang="ts">`
- Pinia stores per domain: `useAuthStore`, `useSessionStore`, etc.
- All API calls go through `src/api/` — never `fetch()` inline in components
- Strict TypeScript — no `any`, use generated types from `@tkaprep/shared-types`

### Database
- All entities use `uuid` PKs (generated via `gen_random_uuid()`)
- All tables have `created_at` and (where mutable) `updated_at`
- Foreign keys are enforced — no orphan rows
- Use `golang-migrate` numbering: `000001_description.up.sql` + `.down.sql`
- Indexes documented in spec.md §6 must exist before going live

### API
- Base path: `/api/v1/...`
- Auth: `Authorization: Bearer <access_token>`
- Errors follow the format in `spec.md` §10.2
- All list endpoints support `?page=&limit=`
- Successful mutations return 200/201 with the resource; never bare 204

---

## Local Development Workflow

```bash
# Setup (first time)
make install
make up
make migrate-up

# Daily
make backend-dev    # hot reload via air
make frontend-dev   # hot reload via vite

# Run a single backend test
cd apps/backend && go test ./internal/service/... -run TestFoo

# After editing openapi.yaml
make generate

# After editing schema
make migrate-create name=add_users_table
# edit the generated .up.sql and .down.sql
make migrate-up
```

---

## Implementation Priority (v1)

In rough order. Each item should be a single PR ideally.

1. ✅ Working skeleton (done — backend boots, frontend boots, postgres connects)
2. Wire OpenAPI pipeline end-to-end (replace hand-written health handler with generated interface)
3. Auth slice — migrations for `users` + `refresh_tokens`, then `/auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout`, `/auth/me`
4. Topics + Questions modules — migrations, repos, services, handlers
5. Tests + Scoring config — migrations, builder API
6. Test sessions — start, save answer, flag, resume, submit (transactional)
7. Results & Review endpoints
8. Leaderboard endpoints
9. Admin endpoints
10. Frontend: replace mock screens (`tka-platform-v2.jsx`) with real Vue views wired to the API

---

## Constraints & Reminders

### Never
- Trust the client clock for session timing
- Store passwords without bcrypt
- Put business logic in HTTP handlers
- Update the API without updating `openapi.yaml` first
- Delete a question that's used in a published test
- Compute `total_score` on the client
- Cache leaderboard data without an invalidation plan

### Always
- Wrap submit logic in a DB transaction
- Validate enum values server-side (don't trust the client)
- Use prepared statements (pgx does this by default with `db.Query`)
- Return structured error responses, not plain strings
- Add database indexes when introducing new query patterns
- Run `make generate` after editing `openapi.yaml`

---

## Open Questions (for the team)

These need decisions but are not blocking v1:

- **Question quality control**: should Super Admin review questions before contributors can use them in tests, or post-hoc only?
- **Leaderboard ethics**: should low-rank entries be anonymized to reduce social pressure?
- **Pending contributors**: should they be able to draft questions before approval, or wait?
- **Email verification**: currently students self-register and become active immediately. Add email verification?

When tackling any of these, update `spec.md` and this file accordingly.

---

## How to Work With Claude Code on This Project

When starting a session, Claude Code should:
1. Read `spec.md` for product requirements
2. Read this `CLAUDE.md` for conventions
3. Read the relevant `apps/{backend,frontend}/README.md` for the area being touched

When making changes:
- For API changes: edit `openapi.yaml` → run `make generate` → implement
- For DB changes: create a migration → run it → write the repo code
- For UI changes: check `tka-platform-v2.jsx` (the mockup) for the intended design

When in doubt, the order of precedence is: `spec.md` > `CLAUDE.md` > existing code.

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
<!-- SPECKIT END -->
