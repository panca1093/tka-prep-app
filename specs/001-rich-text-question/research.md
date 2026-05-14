# Research: Rich Text Question Description

## Decision 1: Frontend Rich Text Editor

**Decision**: TipTap (ProseMirror-based) with custom formula and image extensions.

**Rationale**:
- TipTap has first-class Vue 3 support with `<script setup>` compatibility via `@tiptap/vue-3`.
- `@tiptap/extension-image` provides inline image with custom upload handler — full control over the upload flow.
- Formula support via custom node extension that renders LaTeX using KaTeX. No prepackaged math extension covers our exact use case (LaTeX input → rendered output), so we build a thin `Formula` node that stores the LaTeX source as `data-formula` attribute and renders via KaTeX.
- Quill was considered but has weaker Vue 3 Composition API integration and its formula module is less flexible.
- TipTap's output is clean HTML — exactly what FR-003 requires.

**Alternatives considered**:
- Quill: older ecosystem, Vue 3 wrapper was unofficial/volatile, less control over paste handling.
- TinyMCE: commercial license, heavy bundle size, overkill for our needs.
- Building from scratch with contenteditable: enormous effort; rejected immediately.

## Decision 2: Formula Rendering

**Decision**: KaTeX (bundled, no CDN).

**Rationale**:
- KaTeX is ~10x faster than MathJax for initial render — critical when 50+ questions load in a test.
- Bundle size: KaTeX ~250KB (min+gzip ~60KB) vs MathJax ~2MB+. Indonesian students may be on slow connections; smaller bundle matters.
- KaTeX covers the LaTeX subset used in TKA/SMBT exams: equations, fractions, square roots, sums, limits, integrals, matrices. MathJax's extra features (AMSmath, AsciiMath) are unlikely to be needed.
- Bundled as a project dependency (`katex` npm package), not CDN — no external network dependency during test-taking.
- KaTeX CSS is imported once in the Vue app; the `katex.renderToString()` API renders inline.

**Alternatives considered**:
- MathJax: more complete LaTeX but heavier, slower, and requires async loading which causes formula flicker (edge case already noted in spec).

## Decision 3: Server-Side HTML Sanitization

**Decision**: bluemonday (`github.com/microcosm-cc/bluemonday`) with a custom policy.

**Rationale**:
- bluemonday is the standard Go HTML sanitizer (used by Hugo, Caddy, etc.).
- Custom policy allows: `<b>`, `<i>`, `<u>`, `<ul>`, `<ol>`, `<li>`, `<p>`, `<br>`, `<img src="...">`, `<span data-formula="...">`.
- `<img>` src is restricted to our uploads URL prefix (`/uploads/`) — no external images.
- `<span data-formula="...">` preserves the LaTeX source while stripping all other attributes and event handlers.
- All script tags, event handlers (`onclick`, `onerror`), and unknown elements are stripped.
- Applied in the service layer on create/update, NOT in the handler.

**Policy sketch**:
```go
policy := bluemonday.NewPolicy()
policy.AllowStandardAttributes()
policy.AllowElements("b", "i", "u", "p", "br", "ul", "ol", "li")
policy.AllowAttrs("src").Matching(regexp.MustCompile(`^/uploads/[a-f0-9-]+\.(png|jpe?g|gif|webp)$`)).OnElements("img")
policy.AllowAttrs("data-formula").OnElements("span")
```

**Alternatives considered**:
- None. bluemonday is the only well-maintained Go HTML sanitizer.

## Decision 4: Image Upload — Reuse Existing Handler

**Decision**: Extend the existing `POST /api/v1/upload` handler for question images.

**Rationale**:
- An upload handler already exists at `apps/backend/internal/handler/http/upload.go` with MIME validation, size limits, UUID filenames, and `/uploads/` serving.
- The current max size is 5 MB; we lower it to 2 MB per the clarified spec.
- The handler already checks authentication via JWT — we add a role check for contributor role.
- The editor's image extension calls this endpoint with a multipart form `file` field, receives `{"url": "/uploads/<uuid>.<ext>"}`, and inserts an `<img src="...">`.
- No new API endpoint needed — a significant simplification.

## Decision 5: Orphan Image Cleanup

**Decision**: On question save, diff the set of image URLs in the saved content against all images stored for that question. Delete unreferenced images from disk.

**Rationale**:
- Aligns with the clarified spec (garbage-collect on save).
- Implementation: extract all `/uploads/<uuid>.<ext>` URLs from the new HTML content using a regex. Query a tracking table (or the question's previous content) for the old set. Delete old files not in the new set.
- No background job, no cron — simple, synchronous, easy to test.

## Decision 6: No Schema Migration Needed

**Decision**: Zero database migrations for this feature.

**Rationale**:
- The `text` columns (`questions.text`, `question_options.text`, `question_statements.text`, `questions.explanation`) are already `text` type and will store HTML strings.
- The `ImageURL` fields already exist on domain entities (e.g., `QuestionOption.ImageURL`). These are separate from the rich text content — they hold a single image per option. We don't touch these.
- Existing plain-text questions remain as-is in the database — they're just text without HTML tags, which renders identically.
- No new tables needed. The upload handler already stores images to disk and returns URLs.
