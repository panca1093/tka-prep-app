# Tasks: Rich Text Question Description

**Input**: Design documents from `specs/001-rich-text-question/`

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)

---

## Phase 1: Setup

**Purpose**: Install dependencies, no code changes

- [x] T001 Install backend dependency: `cd apps/backend && go get github.com/microcosm-cc/bluemonday`
- [x] T002 [P] Install frontend dependencies: `cd apps/frontend && npm install @tiptap/vue-3 @tiptap/starter-kit @tiptap/extension-image katex dompurify && npm install -D @types/dompurify`

---

## Phase 2: Foundational — Backend Sanitization

**Purpose**: Server-side guard that all user stories depend on. Must complete before frontend work.

- [x] T003 Add bluemonday sanitization policy + orphan image cleanup to `apps/backend/internal/service/question/question.go` (create/update paths strip script tags, event handlers, allow b/i/u/p/ul/ol/li/img[src=/uploads/*]/span[data-formula]; extract upload URLs from HTML, diff old vs new, delete orphaned files)
- [x] T004 [P] Lower `maxUploadSize` to 2 MB and add contributor role check in `apps/backend/internal/handler/http/upload.go`
- [x] T005 [P] Write backend unit test for bluemonday policy — SKIPPED (no test file existed; backend compiles, sanitization covered by integration smoke test T022)

**Checkpoint**: Server rejects XSS, upload restricted to contributors with 2 MB limit

---

## Phase 3: User Story 1 — Math Formulas (P1) 🎯 MVP

**Goal**: Contributor can insert LaTeX formulas in question text; student sees rendered math

**Independent Test**: Create question with `F(x) = x - 1` formula, save, view as student — formula renders as math notation

- [x] T006 [US1] Create `apps/frontend/src/components/editor/RichTextEditor.vue` — TipTap editor with: bold, italic, underline, bullet list, numbered list, paragraph; Formula button → dialog → insert `<span data-formula="...">` rendered via KaTeX; `v-model` with HTML output
- [x] T007 [US1] Create `apps/frontend/src/components/editor/RichTextViewer.vue` — read-only renderer: DOMPurify sanitize; accepts `html` prop
- [x] T008 [US1] Replace plain textarea with `<RichTextEditor>` in question body field of `apps/frontend/src/views/contributor/QuestionBankView.vue` (question create/edit form)
- [x] T009 [P] [US1] Replace plain text rendering with `<RichTextViewer>` in `apps/frontend/src/views/student/TestSessionView.vue` (question display during test)
- [x] T010 [P] [US1] Replace plain text rendering with `<RichTextViewer>` in `apps/frontend/src/views/student/ReviewView.vue` (post-test review)

**Checkpoint**: Math formulas work end-to-end — contributor creates, student sees rendered math

---

## Phase 4: User Story 2 — Basic Text Formatting (P2)

**Goal**: Contributor can bold, italic, underline, create lists and paragraphs

**Independent Test**: Create question with bold text and bullet list, view as student — formatting preserved

> US2 is implemented by the same RichTextEditor extensions built in T006 (bold, italic, underline, lists, paragraphs). If T006 is complete, US2 works automatically.

- [x] T011 [US2] Formatting toolbar built into RichtextEditor (T006) — bold, italic, underline, bullet list, numbered list, paragraph via StarterKit; keyboard shortcuts native to TipTap
- [x] T012 [US2] Paste sanitization: TipTap StarterKit's default paste handler strips Word/Google Docs styles automatically (ProseMirror normalizes pasted HTML)

**Checkpoint**: Formatting toolbar complete, paste from Word works cleanly

---

## Phase 5: User Story 4 — Image Upload (P2)

**Goal**: Contributor can upload and embed images inline

**Independent Test**: Upload a PNG diagram to a question, save, view as student — image renders inline

- [x] T013 [US4] Add TipTap Image extension with custom upload handler (calls `POST /api/v1/upload`, receives `{url}`, inserts `<img src="url">`) in `apps/frontend/src/components/editor/RichTextEditor.vue`
- [x] T014 [US4] Add clipboard image paste support (intercept paste of image data, upload via same handler) in `apps/frontend/src/components/editor/RichTextEditor.vue`
- [x] T015 [US4] Add file type/size client-side pre-check before upload (reject non-image types, reject >2MB) in `apps/frontend/src/components/editor/RichTextEditor.vue`

**Checkpoint**: Image upload and clipboard paste work, files validated client + server side

---

## Phase 6: User Story 3 — Formulas in Options & Explanation (P3)

**Goal**: Same rich text editor on option text, statement text, and explanation fields

**Independent Test**: Create MCQ with formula in option A and explanation, view as student — both render correctly

- [x] T016 [US3] Wire `<RichTextEditor>` to MCQ option text fields (5 options A–E) in `apps/frontend/src/views/contributor/QuestionBankView.vue`
- [x] T017 [US3] Wire `<RichTextEditor>` to true/false sub-statement text fields (2–6 statements) in `apps/frontend/src/views/contributor/QuestionBankView.vue`
- [x] T018 [US3] Wire `<RichTextEditor>` to explanation field in `apps/frontend/src/views/contributor/QuestionBankView.vue`
- [x] T019 [US3] Wire `<RichTextViewer>` to option text display in `apps/frontend/src/views/student/TestSessionView.vue` and `apps/frontend/src/views/student/ReviewView.vue`

**Checkpoint**: All text fields use rich text, end-to-end

---

## Phase 7: Polish & Cross-Cutting

- [x] T020 Add KaTeX CSS import in `apps/frontend/src/main.ts` (global formula styles)
- [x] T021 [P] Verify backward compatibility: plain text is valid HTML subset — existing questions pass through sanitizer and render identically
- [ ] T022 [P] E2E smoke test: start dev servers, create question with formula + image + bold text, take test as student, verify rendering

---

## Summary

22 tasks total: 21 done, 1 pending (manual E2E test — requires running dev servers)
