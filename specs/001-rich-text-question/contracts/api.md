# API Contracts: Rich Text Question Description

## Existing Endpoints — Behavioral Changes Only

The existing OpenAPI spec endpoints do NOT change their request/response shapes. The `text` fields in question/option/statement bodies were already typed as `string` and remain `string`. The only change is the *content* of those strings (now HTML).

### Sanitization Contract

All `text` and `explanation` fields are sanitized on create/update:

**Input**: Any HTML string from the rich text editor.
**Transformation**: bluemonday policy applied — strips all tags except: `b, i, u, p, br, ul, ol, li, img, span`.
- `<img src>` must match `/uploads/[uuid].[png|jpg|jpeg|gif|webp]`
- `<span>` may only have `data-formula` attribute
- All other attributes, event handlers, scripts stripped

**Output**: Sanitized HTML string stored in database.

### Upload Endpoint (modified)

**Endpoint**: `POST /api/v1/upload`  
**Auth**: Contributor role required (added constraint)  
**Max size**: 2 MB (changed from 5 MB)  
**Accepted MIME types**: `image/jpeg`, `image/png`, `image/gif`, `image/webp`

**Request**: `multipart/form-data` with field `file`  
**Response 200**:
```json
{"url": "/uploads/<uuid>.<ext>"}
```
**Response 400** (file too large):
```json
{"error": "file too large (max 2 MB)"}
```
**Response 400** (unsupported type):
```json
{"error": "unsupported file type"}
```
**Response 401**:
```json
{"error": "not authenticated"}
```
**Response 403**:
```json
{"error": "contributor role required"}
```

### Static File Serving

`GET /uploads/<filename>` serves uploaded images. Already exists.

## Frontend Contract

The rich text editor component MUST:
- Accept `modelValue: string` (HTML content)
- Emit `update:modelValue` with sanitized HTML on change
- Handle image upload by calling `POST /api/v1/upload` with multipart form
- Accept clipboard paste for images (convert to upload flow)

The render-only component (student view) MUST:
- Accept `html: string` prop
- Render HTML content with KaTeX for all `[data-formula]` spans
- Render `<img>` tags with the existing `/uploads/` URLs
- Use DOMPurify (client-side) as defense-in-depth before rendering
