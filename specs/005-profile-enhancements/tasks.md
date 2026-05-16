# Tasks: Enhanced Student Profile

**Input**: Design documents from `specs/005-profile-enhancements/`
**Prerequisites**: plan.md ✅ | spec.md ✅ | research.md ✅ | data-model.md ✅ | contracts/api-contracts.md ✅

**Tests**: Not requested — no test tasks generated.

**Organization**: Tasks grouped by user story for independent implementation and delivery.

---

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1–US2)

---

## Phase 1: Setup (Database Migration)

**Purpose**: Add gender, phone, and avatar_url columns to the users table.

- [ ] T001 Create migration `000022_profile_fields.up.sql` in `apps/backend/migrations/` — `CREATE TYPE gender AS ENUM ('male', 'female', 'other')`; `ALTER TABLE users ADD COLUMN gender gender`; `ALTER TABLE users ADD COLUMN phone varchar(20)`; `ALTER TABLE users ADD COLUMN avatar_url text`
- [ ] T002 Create migration `000022_profile_fields.down.sql` in `apps/backend/migrations/` — drop columns and type
- [ ] T003 Apply migration — run SQL directly on the postgres container

**Checkpoint**: `\d users` shows gender, phone, avatar_url columns (all nullable).

---

## Phase 2: Domain & OpenAPI (Foundational — blocks all user stories)

**Purpose**: Go types, API contract updates, code regeneration. ALL user stories depend on this phase.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T004 Define `Gender` type and constants (`GenderMale`, `GenderFemale`, `GenderOther`) in `apps/backend/internal/domain/user.go`; add `Gender *Gender`, `Phone *string`, `AvatarURL *string` fields to `User` struct
- [ ] T005 Add `gender`, `phone`, `avatar_url` to `UpdateProfileRequest` schema (nullable) and `UserResponse` schema (nullable) in `packages/shared-types/openapi.yaml`
- [ ] T006 Run `make generate` to regenerate Go and TS client code

**Checkpoint**: Generated code compiles. User struct has new fields. TS types include gender/phone/avatar_url.

---

## Phase 3: User Story 1 — Student Views and Edits Their Enhanced Profile (Priority: P1) 🎯 MVP

**Goal**: Profile page shows gender, phone, avatar. Student can edit all three.

**Independent Test**: Student opens /profile → sees gender/phone/avatar fields → changes phone → saves → reload → phone persists.

### Implementation for User Story 1

- [ ] T007 [US1] Add `UpdateProfile(ctx, id uuid.UUID, gender *domain.Gender, phone *string, avatarURL *string)` to `UserRepository` interface in `apps/backend/internal/repository/interfaces.go`
- [ ] T008 [US1] Implement `UpdateProfile` in `apps/backend/internal/repository/postgres/user.go` — UPDATE users SET gender=$1, phone=$2, avatar_url=$3, updated_at=NOW() WHERE id=$4
- [ ] T009 [US1] Update `UpdateProfile` in `apps/backend/internal/service/auth/auth.go` — accept gender/phone/avatarURL params; validate gender enum, strip non-digits from phone (require ≥10 digits if provided); call repository
- [ ] T010 [US1] Update `PatchAuthMe` handler in `apps/backend/internal/handler/http/auth.go` — map `gender`, `phone`, `avatar_url` from request body to service call
- [ ] T011 [US1] Update `toUserResponse` in `apps/backend/internal/handler/http/auth.go` — map `Gender`, `Phone`, `AvatarURL` from domain User to API response
- [ ] T012 [US1] Redesign `apps/frontend/src/views/student/ProfileView.vue` — read-only fields: name, email, education_level; editable fields: gender (select: Male/Female/Prefer not to say), phone (text input), avatar (ImageUpload component); sign out button at bottom; save button triggers PATCH /auth/me
- [ ] T013 [US1] Update auth store in `apps/frontend/src/stores/auth.ts` — ensure `user` object includes `gender`, `phone`, `avatar_url` fields

**Checkpoint**: Student can view and edit gender, phone, avatar from profile page. Data persists across refresh.

---

## Phase 4: User Story 2 — Profile Picture Display Across the App (Priority: P2)

**Goal**: Avatar appears in sidebar and leaderboard when set.

**Independent Test**: Upload avatar → check sidebar → check leaderboard → avatar visible in both places.

### Implementation for User Story 2

- [ ] T014 [US2] Update `apps/frontend/src/layouts/AppLayout.vue` — replace initials text in sidebar user-info area with `<img>` when `auth.user.avatar_url` is set; fallback to initials when null
- [ ] T015 [US2] Verify `apps/frontend/src/views/student/LeaderboardView.vue` — podium avatars and row mini-avatars already use `initials()` with gradient background; update to use `<img>` when `avatar_url` is present on the entry

**Checkpoint**: Avatar image visible in sidebar and leaderboard for students who have uploaded one.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Validation, edge cases, and final verification.

- [ ] T016 Add server-side phone validation in `apps/backend/internal/service/auth/auth.go` — strip all non-digit characters; reject if length < 10 after stripping
- [ ] T017 Add client-side avatar upload validation in `apps/frontend/src/views/student/ProfileView.vue` — reject files > 2 MB, reject non-image types before upload
- [ ] T018 Verify ProfileView renders correctly at mobile (≤768px) and desktop breakpoints — responsive form, no overflow
- [ ] T019 Run `docker compose up -d --build backend frontend` and verify full flow: open profile page → update gender/phone → upload avatar → check sidebar + leaderboard → refresh → all data persists

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Migration)**: No dependencies — start immediately
- **Phase 2 (Domain & OpenAPI)**: Depends on Phase 1 — BLOCKS all user stories
- **Phase 3 (US1)**: Depends on Phase 2 — profile page with edit
- **Phase 4 (US2)**: Depends on Phase 3 (avatar must be uploadable before displaying) — can start after T012
- **Phase 5 (Polish)**: Depends on all previous phases

### User Story Dependencies

- **US1 (P1)**: After Phase 2 — no other story dependencies
- **US2 (P2)**: After US1 — needs avatar_url field populated before display code matters

### Within Each User Story

- Repo → Service → Handler → Frontend
- Backend complete before wiring the frontend

### Parallel Opportunities

- T001 and T002 can use the same migration file (paired up/down)
- T014 and T015 touch different files — can run in parallel after US1 completes
- T016 and T017 touch different files (backend validation vs frontend validation) — can run in parallel

---

## Parallel Example: Phase 4 (US2)

```
After Phase 3 completes:
  T014: Update AppLayout.vue sidebar avatar
  T015: Update LeaderboardView.vue avatar
Both touch different files — launch in parallel.
```

---

## Implementation Strategy

### MVP (US1 Only)

1. Complete Phase 1: Migration
2. Complete Phase 2: Domain & OpenAPI
3. Complete Phase 3: Profile page with gender, phone, avatar edit
4. **STOP and VALIDATE**: Open /profile → edit gender → save → refresh → persists

### Incremental Delivery

1. Migration + Domain + OpenAPI → Foundation ready
2. US1 → Students can edit enhanced profile (**MVP**) ✅
3. US2 → Avatar visible in sidebar + leaderboard ✅
4. Polish → Validation + responsive + regression ✅

---

## Notes

- All new fields are nullable — no data migration needed for existing users
- Reuses existing `/api/v1/upload` endpoint — no new upload handler needed
- Gender enum values stored lowercase (`male`, `female`, `other`), displayed as "Male"/"Female"/"Prefer not to say"
- Avatar URL comes from the upload endpoint response, stored as-is
- Phone validation strips non-digits server-side; client-side only does basic type checking
