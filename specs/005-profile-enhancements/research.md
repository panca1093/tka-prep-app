# Research: Enhanced Student Profile

## Decision 1: Gender enum vs. varchar

**Decision**: PostgreSQL native ENUM `gender` with values `male`, `female`, `other`.

**Rationale**: Consistent with existing patterns (`role`, `status`, `education_level` all use native enums). Nullable — existing users default to null.

## Decision 2: Phone number storage

**Decision**: `varchar(20)` nullable. Server-side validation: strip non-digits, require ≥10 digits if provided.

**Rationale**: Simple text column avoids complexity of phone number libraries. Indonesian phone numbers are 10-13 digits. No SMS verification in v1.

## Decision 3: Avatar storage

**Decision**: Store as `avatar_url text` — a URL pointing to the uploaded file on the server's `/uploads/` directory. Reuse existing `/api/v1/upload` endpoint.

**Rationale**: Already have image upload infrastructure. No need for S3 or external storage. URL stored in DB, image served from local filesystem.

## Decision 4: Profile update endpoint

**Decision**: Extend existing `PATCH /auth/me` endpoint (already used for education_level). Add `gender`, `phone`, `avatar_url` to `UpdateProfileRequest`.

**Rationale**: RESTful — same endpoint for all profile field updates. No new routes needed.

## Decision 5: Avatar display pattern

**Decision**: Use `<img>` with fallback to initials. The LeaderboardView already has an `initials()` + gradient avatar pattern. When `avatar_url` is non-null, render `<img>` instead of the initials div.

**Rationale**: Reuses existing component patterns. No new component needed.
