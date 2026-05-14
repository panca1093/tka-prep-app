# Implementation Plan: Rich Text Question Description

**Branch**: `001-rich-text-question` | **Date**: 2026-05-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/001-rich-text-question/spec.md`

## Summary

Replace plain textarea inputs with a TipTap rich text editor across all question-related text fields (question body, options, true/false statements, explanation). Add KaTeX-based formula rendering (LaTeX input → rendered math notation) and inline image upload via the existing `POST /api/v1/upload` endpoint. No database migrations required — text columns already store `text` type and will now hold HTML. Server-side sanitization via bluemonday.

## Technical Context

**Language/Version**: Go 1.25 (backend) + TypeScript 5.6 / Vue 3.5 (frontend)
**Primary Dependencies**: TipTap (`@tiptap/vue-3`, `@tiptap/starter-kit`, `@tiptap/extension-image`), KaTeX, DOMPurify, bluemonday (`github.com/microcosm-cc/bluemonday`)
**Storage**: PostgreSQL 16 (no schema changes); local filesystem for image uploads (existing `UPLOAD_DIR`)
**Testing**: Go standard `testing` + `go vet`; Vitest for frontend components
**Target Platform**: Web (desktop + mobile browsers)
**Project Type**: Web application (Go backend + Vue 3 SPA frontend)
**Performance Goals**: Formula render < 1s per question; image upload < 3s for 200KB PNG; no regression on existing plain-text question load time
**Constraints**: No database migrations; backward compatible with existing plain-text questions; 2 MB max image upload; XSS-safe HTML output
**Scale/Scope**: 4 text fields across question forms (question body, 5 options, 2-6 statements, explanation); 2 new Vue components; 1 backend service change; 0 new API endpoints

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Clean Layer Architecture | ✅ PASS | Sanitization in service layer (`service/question.go`), not handler. Upload handler role check added; business logic stays in service. |
| II. OpenAPI-First API Design | ✅ PASS | No API schema changes — `text` fields remain `string` type. Upload endpoint already exists. No OpenAPI spec edits needed. |
| III. Server-Authoritative Session Integrity | ✅ PASS | Feature touches question CRUD only — session timing and scoring unchanged. |
| IV. Database Integrity by Design | ✅ PASS | No migrations. No new tables. No data integrity risk from content format change. |
| V. Security Non-Negotiables | ✅ PASS | HTML sanitized server-side (bluemonday). RBAC on upload (contributor only). Max upload size enforced. DOMPurify client-side as defense-in-depth. |

**Gate result**: PASS — no violations.

## Project Structure

### Documentation (this feature)

```text
specs/001-rich-text-question/
├── plan.md              # This file
├── research.md          # Phase 0: library choices & rationale
├── data-model.md        # Phase 1: entity changes (none needed)
├── quickstart.md        # Phase 1: file checklist & setup
├── contracts/           # Phase 1: API & frontend contracts
│   └── api.md
└── tasks.md             # Phase 2: /speckit-tasks output (NOT created here)
```

### Source Code (repository root)

```text
apps/backend/
├── internal/
│   ├── service/
│   │   └── question.go          # ADD: HTML sanitization + orphan image cleanup
│   └── handler/http/
│       └── upload.go            # MODIFY: 2MB limit, contributor role check

apps/frontend/
├── src/
│   ├── components/
│   │   └── editor/
│   │       ├── RichTextEditor.vue   # NEW: TipTap editor (formatting, formula, image)
│   │       └── RichTextViewer.vue   # NEW: Read-only renderer (KaTeX + DOMPurify)
│   └── views/
│       └── contributor/
│           └── QuestionForm.vue     # MODIFY: textarea → RichTextEditor
│       └── student/
│           ├── TestSession.vue      # MODIFY: plain text → RichTextViewer
│           └── ResultReview.vue     # MODIFY: plain text → RichTextViewer
```

**Structure Decision**: Web application (Option 2). Feature is entirely within existing `apps/backend` and `apps/frontend` directories.

## Complexity Tracking

No violations — table is empty.
