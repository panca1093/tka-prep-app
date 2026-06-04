# Brief: TKAPrep — Image Storage GCS Migration + Per-Contributor Question Visibility

## Problem

Two improvements needed for the contributor experience and production readiness:

1. **Image storage is local disk only.** Uploads go to `./uploads/` on the backend filesystem, served via `http.FileServer`. This doesn't scale — images are lost on redeploy, can't be shared across instances, and local I/O is not production-grade.

2. **Contributors see all questions in the question bank.** The `GET /questions` endpoint returns every question in the system regardless of who created it. A contributor's question bank should only show their own questions. Admins retain full visibility.

## Idea

### Scope 1: GCS Image Storage (Hybrid)

Add a `STORAGE_BACKEND` env var (`local` | `gcs`). When `gcs`, the upload handler writes to GCS instead of local disk, and static file serving is replaced with a GCS proxy. When `local`, behavior is unchanged.

New env vars:
- `STORAGE_BACKEND` — `local` (default) or `gcs`
- `GCS_BUCKET` — bucket name
- `GCS_CREDENTIALS_PATH` — path to service account JSON (optional, fall back to ADC)

Upload flow with GCS:
1. Contributor POSTs multipart to `/api/v1/upload` (unchanged API)
2. Backend validates MIME + size (2MB cap) as before
3. Writes to GCS bucket under `questions/<uuid>.<ext>`
4. Returns the GCS object path (frontend-agnostic — it just renders the URL)

Serving flow with GCS:
- Replace `http.FileServer` with a handler that streams from GCS or generates signed URLs
- Frontend doesn't change — it still receives and renders URLs

Orphan image cleanup (`cleanupOrphanImages`) needs a GCS equivalent.

### Scope 2: Per-Contributor Question Visibility

Auto-filter `GET /questions` by `contributor_id = caller_id` when the caller is a contributor. Admins see all questions. Enforced at the SQL/repository layer.

Same filter on `GET /questions/{id}` — contributor fetching a question they don't own gets 404. Admin bypasses.

## Users

- **Contributors** — question bank only shows their own questions. They can still use any question in their tests (questions remain shared for test composition — no change to test builder behavior).
- **Admins** — see all questions, no change.

## Success Criteria

1. `STORAGE_BACKEND=gcs` — image uploads persist to GCS, served correctly, orphan cleanup works
2. `STORAGE_BACKEND=local` — existing behavior fully preserved (default)
3. Contributor A only sees own questions in `GET /questions`
4. Admin sees all questions in `GET /questions`
5. Contributor fetching another's question by ID → 404
6. Admin fetching any question by ID → works
7. Existing tests pass

## Constraints

- Go 1.22, chi, pgx, zerolog — clean layers
- No breaking API changes
- Frontend: no required changes (server-side filter)
- Docker Compose local dev: keep `STORAGE_BACKEND=local` default
- 2MB upload cap stays

## Edge Cases

- **GCS misconfigured**: fail fast on boot, clear error
- **Orphan cleanup on GCS**: list + delete objects by prefix instead of `os.Remove`
- **Contributor with 0 questions**: 200 OK, empty list
- **Student calling GET /questions**: still 403 (unchanged)
