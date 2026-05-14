# Quickstart: Rich Text Question Description

## Overview

Upgrade question text fields from plain textarea to a TipTap rich text editor with formula rendering (KaTeX) and image upload support. No database migrations needed. An upload handler already exists.

## Files Changed

### Backend
| File | Change |
|------|--------|
| `apps/backend/internal/service/question.go` | Add HTML sanitization (bluemonday) on create/update; add orphan image cleanup on save |
| `apps/backend/internal/handler/http/upload.go` | Lower max size to 2 MB; add contributor role check |
| `apps/backend/go.mod` | Add `github.com/microcosm-cc/bluemonday` dependency |

### Frontend
| File | Change |
|------|--------|
| `apps/frontend/src/components/editor/RichTextEditor.vue` | New — TipTap editor wrapper with formula, image, formatting extensions |
| `apps/frontend/src/components/editor/RichTextViewer.vue` | New — read-only renderer with KaTeX + DOMPurify |
| `apps/frontend/src/views/contributor/QuestionForm.vue` | Replace `<textarea>` with `<RichTextEditor>` for question text, options, statements, explanation |
| `apps/frontend/src/views/student/TestSession.vue` | Replace plain text rendering with `<RichTextViewer>` |
| `apps/frontend/src/views/student/ResultReview.vue` | Replace plain text rendering with `<RichTextViewer>` |
| `apps/frontend/package.json` | Add `@tiptap/vue-3`, `@tiptap/starter-kit`, `@tiptap/extension-image`, `katex`, `dompurify` |

## Key Dependencies

```bash
# Backend
cd apps/backend
go get github.com/microcosm-cc/bluemonday

# Frontend
cd apps/frontend
npm install @tiptap/vue-3 @tiptap/starter-kit @tiptap/extension-image katex dompurify
npm install -D @types/dompurify
```

## Sanitization Flow

```
[RichTextEditor] → HTML string → API request → bluemonday policy → DB
                                                    ↓
[RichTextViewer] ← HTML string ← API response ← DB
        ↓
   DOMPurify (defense-in-depth) → KaTeX render formulas → Display
```

## Testing

- **Backend unit**: Test bluemonday policy strips `<script>`, allows `<b>/<i>/<img src="/uploads/...">`, blocks `<img src="https://evil.com/x.jpg">`
- **Backend unit**: Test orphan cleanup deletes unreferenced images, keeps referenced ones
- **Frontend component**: Test TipTap editor emits HTML with format/formula/image content
- **Frontend component**: Test RichTextViewer renders KaTeX and sanitizes with DOMPurify
- **E2E**: Create question with formula + image + bold text → take test as student → verify all render correctly
