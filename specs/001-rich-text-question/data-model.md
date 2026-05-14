# Data Model: Rich Text Question Description

## Summary

No schema migrations. Existing `text` columns store HTML instead of plain text. The upload handler already exists. This feature is purely a frontend + sanitization change.

## Entities Affected

### `questions`
| Field | Change |
|-------|--------|
| `text` | Was plain text → now stores HTML (bold, italic, lists, images, formulas) |
| `explanation` | Was plain text → now stores HTML |

No new columns, no type changes. Backward compatible — plain text is valid HTML (no tags means rendered as-is).

### `question_options`
| Field | Change |
|-------|--------|
| `text` | Was plain text → now stores HTML |

### `question_statements`
| Field | Change |
|-------|--------|
| `text` | Was plain text → now stores HTML |

## Entities NOT Changed

- `question_options.image_url` — existing field for a single image per option. Untouched. Inline images in rich text use `<img src="/uploads/...">` directly in the `text` field, not this column.
- `question_statements.image_url` — same.
- `questions.image_url` — same.
- `session_answers` — students still select options/statements by UUID. No change.
- `test_results` — no change.

## Upload Storage

Existing handler at `POST /api/v1/upload` stores files to `UPLOAD_DIR` on disk. Files are named `<uuid>.<ext>`. The handler already:
- Validates MIME type (image/jpeg, image/png, image/gif, image/webp)
- Enforces size limit (currently 5 MB — reduced to 2 MB per spec)
- Returns `{"url": "/uploads/<uuid>.<ext>"}`

Files are served via a static file handler at `/uploads/`.

## Orphan Cleanup

On question save, the service layer:
1. Extracts all `/uploads/<uuid>.<ext>` paths from the new HTML using regex
2. Compares to the set of paths in the previous content (if updating)
3. Deletes local files for paths that are in the old set but not the new set (orphaned by edit)

No tracking table needed — we diff the content directly.
