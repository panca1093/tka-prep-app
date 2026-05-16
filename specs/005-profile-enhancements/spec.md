# Feature Specification: Enhanced Student Profile

**Feature Branch**: `scholar-redesign`
**Created**: 2026-05-17
**Status**: Draft
**Input**: User description: "working on profil page, can we add more information, like gender, phone number and profile picture"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Student Views and Edits Their Enhanced Profile (Priority: P1)

A student opens their profile page and sees their name, email, education level, gender, phone number, and profile picture. They can edit their gender, phone number, and profile picture directly from the profile page.

**Why this priority**: Core profile completeness. Students expect a profile page with basic identity information.

**Independent Test**: Student navigates to /profile → sees all fields populated → changes phone number → saves → refresh → phone number persists.

**Acceptance Scenarios**:

1. **Given** a student is logged in, **When** they open the profile page, **Then** they see their name (read-only), email (read-only), education level (read-only), gender (editable), phone number (editable), and profile picture (uploadable)
2. **Given** a student edits their gender and saves, **When** the save completes, **Then** the new gender is displayed and persisted
3. **Given** a student uploads a profile picture, **When** the upload completes, **Then** the picture is displayed as their avatar across the app (sidebar, leaderboard, etc.)

---

### User Story 2 — Profile Picture Display Across the App (Priority: P2)

The student's profile picture appears in the sidebar, on leaderboard entries, and on any page that currently shows their name or initials.

**Why this priority**: Visual identity across the platform. Makes the app feel personal.

**Independent Test**: Upload profile picture → check sidebar, leaderboard, and result review → picture appears in all locations.

**Acceptance Scenarios**:

1. **Given** a student has uploaded a profile picture, **When** they view the sidebar, **Then** their picture is shown instead of text initials
2. **Given** a student has uploaded a profile picture, **When** they appear on the leaderboard, **Then** their picture is shown alongside their name

---

### Edge Cases

- What happens when a student uploads a very large image? → Reject images over 2 MB; resize/crop client-side before upload.
- What happens when no profile picture is set? → Show initials avatar (existing behavior).
- What happens when phone number contains non-digit characters? → Strip non-digits on save; validate minimum 10 digits.
- What if the student is not registered as "student" role? → Gender/phone fields are only meaningful for students; hidden or read-only for contributors/admins.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The student profile page MUST display: name (read-only), email (read-only), education level (read-only), gender (selectable: Male/Female/Prefer not to say), phone number (text input), and profile picture (image upload).
- **FR-002**: Gender MUST be persisted to the `users` table via `PATCH /auth/me`.
- **FR-003**: Phone number MUST be persisted to the `users` table via `PATCH /auth/me`.
- **FR-004**: Profile picture MUST be uploaded via the existing `POST /api/v1/upload` endpoint; the resulting URL stored as `avatar_url` on the user record.
- **FR-005**: Profile picture upload MUST enforce a 2 MB size limit and image-only file types (JPEG, PNG, WebP).
- **FR-006**: The sidebar MUST display the profile picture when available, falling back to initials when not set.
- **FR-007**: Leaderboard entries (podium + rows) MUST display the profile picture or initials avatar.
- **FR-008**: The `PATCH /auth/me` endpoint MUST accept `gender`, `phone`, and `avatar_url` fields.

### Key Entities

- **User** (modified): Adds `gender` (nullable enum: male/female/other), `phone` (nullable varchar), `avatar_url` (nullable text)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A student can update their gender, phone number, and profile picture from the profile page in under 30 seconds.
- **SC-002**: Profile picture appears in sidebar and leaderboard within 2 seconds of upload completion.
- **SC-003**: Images over 2 MB are rejected with a clear error message before upload begins.
- **SC-004**: Existing students with no gender/phone/avatar set see empty/default states without errors.

## Assumptions

- Gender is a fixed enum: male, female, other (nullable — existing users default to null).
- Phone number is free-text with basic validation (10+ digits after stripping non-digits).
- Profile picture uses the existing `/api/v1/upload` endpoint (already supports image upload with 2 MB limit).
- Avatar display across the app reuses the initials-avatar component pattern already in LeaderboardView.
- This feature does not change how education level works (still read-only after registration).
