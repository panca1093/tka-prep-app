# Implementation Plan: Enhanced Student Profile

**Branch**: `scholar-redesign` | **Date**: 2026-05-17 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/005-profile-enhancements/spec.md`

## Summary

Add gender, phone number, and profile picture (avatar) fields to the user profile. Students can edit these fields from the profile page. The avatar displays across the app (sidebar, leaderboard). Backend: new columns on `users` table + `PATCH /auth/me` extended. Frontend: redesigned ProfileView with image upload.

## Technical Context

**Language/Version**: Go 1.22 (backend), TypeScript + Vue 3 (frontend)
**Primary Dependencies**: chi, pgx, zerolog (backend); Vite, Pinia, vue-router (frontend)
**Storage**: PostgreSQL 16 — new columns on `users`: `gender` (nullable enum), `phone` (nullable varchar), `avatar_url` (nullable text)
**Testing**: go test (backend), manual browser (frontend)
**Target Platform**: Web SPA
**Project Type**: Web application (monorepo: `apps/backend` + `apps/frontend`)
**Performance Goals**: Profile update < 500ms; avatar upload < 2s
**Constraints**: OpenAPI-first; clean layer architecture; reuse existing `/api/v1/upload` endpoint; 2 MB image limit

## Constitution Check

| Principle | Status | Notes |
|---|---|---|
| I. Clean Layer Architecture | ✅ PASS | domain → repository ← service → handler — all new fields follow existing patterns |
| II. OpenAPI-First API Design | ✅ PASS | openapi.yaml edited first for UpdateProfileRequest/UserResponse |
| III. Server-Authoritative Session | ✅ N/A | No session changes |
| IV. Database Integrity | ✅ PASS | Nullable columns, enum for gender |
| V. Security Non-Negotiables | ✅ PASS | Image upload reuses existing upload handler (2 MB limit, type check); phone number validated server-side |

**Gate result**: All gates pass. Proceed.

## Project Structure

### Source Code

```text
apps/backend/
├── migrations/
│   ├── 000022_profile_fields.up.sql   ← NEW
│   └── 000022_profile_fields.down.sql ← NEW
└── internal/
    ├── domain/
    │   └── user.go                    ← add Gender type, Gender/Phone/AvatarURL fields
    ├── repository/
    │   ├── interfaces.go              ← no changes (UpdateEducationLevel pattern)
    │   └── postgres/
    │       └── user.go                ← UpdateProfile method (gender, phone, avatar)
    ├── service/
    │   └── auth/
    │       └── auth.go                ← UpdateProfile: add gender, phone, avatar_url
    └── handler/http/
        └── auth.go                    ← PatchAuthMe: map new fields

packages/shared-types/
└── openapi.yaml                       ← add gender, phone, avatar_url to schemas

apps/frontend/src/
├── views/student/
│   └── ProfileView.vue                ← redesigned with gender, phone, avatar upload
└── layouts/
    └── AppLayout.vue                  ← avatar in sidebar
```

## Implementation Sequence

### Step 1 — Migration

```sql
CREATE TYPE gender AS ENUM ('male', 'female', 'other');
ALTER TABLE users ADD COLUMN gender gender;
ALTER TABLE users ADD COLUMN phone varchar(20);
ALTER TABLE users ADD COLUMN avatar_url text;
```

### Step 2 — Domain

Add to `user.go`:
- `type Gender string` with constants `GenderMale`, `GenderFemale`, `GenderOther`
- Fields `Gender *Gender`, `Phone *string`, `AvatarURL *string` on `User` struct

### Step 3 — OpenAPI + codegen

Add to `UpdateProfileRequest`: `gender` (enum male/female/other, nullable), `phone` (string, nullable), `avatar_url` (string, nullable)
Add to `UserResponse`: same fields
Run `make generate`

### Step 4 — Repository

Add `UpdateProfile(ctx, id, gender, phone, avatarURL)` to `UserRepository` interface
Implement in postgres: `UPDATE users SET gender=$1, phone=$2, avatar_url=$3, updated_at=NOW() WHERE id=$4`

### Step 5 — Service

Update `UpdateProfile` in auth service to accept and validate gender/phone/avatar_url
- Gender: must be one of male/female/other or nil
- Phone: strip non-digits, validate ≥10 digits if provided
- AvatarURL: no validation (comes from trusted upload endpoint)

### Step 6 — Handler

Map `gender`, `phone`, `avatar_url` from `UpdateProfileRequest` → service
Map fields in `toUserResponse`

### Step 7 — Frontend

**ProfileView.vue**: Redesign with:
- Read-only fields: name, email, education level
- Editable fields: gender (select), phone (input), avatar (ImageUpload component)
- Sign out button at bottom
**AppLayout.vue**: Replace text initials with `<img>` when `user.avatar_url` is set
**LeaderboardView.vue**: Already done (from linter changes) — verify avatar_url is used
