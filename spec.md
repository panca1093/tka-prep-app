# TKAPrep — Online Test Preparation Platform

## Overview

A web-based platform for Indonesian students preparing for university entrance examinations (TKA Saintek, TKA Soshum, SMBT). Students take realistic timed test simulations, receive instant scoring with topic-level analytics, and track their progress on a global leaderboard. Contributors author questions and assemble tests with flexible scoring formulas. Super Admins manage platform users and content.

---

## 1. Goals & Non-Goals

### Goals
- Deliver realistic test simulations that mirror real TKA/SMBT exam mechanics
- Provide instant, granular scoring feedback with per-topic breakdown
- Enable contributors to author questions and tests independently
- Allow flexible scoring formulas (correct/wrong/blank points configurable per test)
- Foster engagement via leaderboards and streaks

### Non-Goals (Out of Scope for v1)
- Real-time proctoring or anti-cheat enforcement (focus mode is "Later" phase)
- Mobile native apps (web-first)
- Payment processing (freemium gating deferred to future iteration)
- Question ownership / visibility rules (all questions shared across contributors)
- Test visibility scoping (all published tests visible to all students)
- Social features (commenting, discussion forums)
- AI-generated questions

---

## 2. User Roles

| Role | Description |
|---|---|
| **Student** | Registers freely, takes tests, views own results and rank |
| **Contributor** | Authors questions, builds tests, sets scoring formulas, views attempt analytics on own tests |
| **Super Admin** | Manages all users and contributors, approves/rejects pending contributors, oversees all platform content |

---

## 3. User Scenarios

### 3.1 Student — Take a test
**Primary Flow:**
1. Student logs in and lands on dashboard showing stats (tests taken, best score, streak, rank)
2. Browses available tests filtered by category (TKA/SMBT) or subtype (Saintek/Soshum)
3. Starts a test → session begins with countdown timer
4. Navigates through questions, can flag questions for review, can answer in any order
5. Submits test (with confirmation modal showing answered/flagged/blank counts) or test auto-submits when timer expires
6. Sees instant result: total score, correct/wrong/blank counts, per-topic breakdown, rank change
7. Optionally reviews each question with correct answer and explanation

**Acceptance:**
- Timer is enforced server-side; auto-submit fires at 0:00
- Partial answers are autosaved as student progresses
- A student can resume an in-progress session if they close the browser
- Flagged questions are visually distinct in the question navigation grid

### 3.2 Student — Track progress
1. Views Score History showing all past attempts with score, date, rank
2. Clicks Review on any past attempt to see per-question breakdown
3. Views global leaderboard (top 100) and own personal rank if outside top 100

### 3.3 Contributor — Author a question
1. Navigates to Question Bank
2. Clicks "+ Add Question" to open inline form
3. Enters question text, 5 options (A–E), marks the correct option, selects topic and difficulty, optionally adds explanation
4. Saves → question appears in their question bank, available for use in any test

### 3.4 Contributor — Build a test
**4-step wizard:**
1. **Test Info** — name, description, category, duration, difficulty
2. **Select Questions** — pick from the shared question bank with topic/difficulty filters
3. **Scoring Rules** — set points for correct/wrong/blank (e.g. `+4 / 0 / 0` for TKA, `+1 / -0.25 / 0` for SAT-style)
4. **Review & Publish** — verify summary, then save as draft or publish

### 3.5 Super Admin — Approve contributor
1. New user requests Contributor role during registration → status = `pending`
2. Super Admin sees pending requests in admin dashboard
3. Approves or rejects → status flips to `active` or `rejected`
4. On approve, user can access contributor features

---

## 4. Functional Requirements

### 4.1 Authentication & User Management
- **FR-001** System MUST support user registration with email + password
- **FR-002** System MUST issue JWT tokens on successful login
- **FR-003** System MUST hash passwords using bcrypt before storage
- **FR-004** System MUST support three roles: `student`, `contributor`, `admin`
- **FR-005** New users registering as `contributor` MUST start with status `pending` and require admin approval
- **FR-006** Students MAY self-register and become `active` immediately
- **FR-007** Super Admin accounts MUST be created via seed/migration, not self-registration
- **FR-008** System MUST enforce role-based access control on all protected endpoints
- **FR-009** System MUST support user statuses: `active`, `inactive`, `suspended`, `pending`

### 4.2 Question Bank
- **FR-010** Contributors MUST be able to create, read, update, and delete their own questions
- **FR-011** Each question MUST have exactly 5 options (A, B, C, D, E)
- **FR-012** Each question MUST have exactly one correct option
- **FR-013** Each question MUST be assigned to one topic and one difficulty (easy/medium/hard)
- **FR-014** Questions MAY include an explanation text shown in post-test review
- **FR-015** System MUST track which contributor authored each question (`contributor_id`)
- **FR-016** All questions MUST be visible to all contributors for use in test building (v1 — no visibility rules)
- **FR-017** Question Bank UI MUST support search by text, filter by topic, filter by difficulty
- **FR-018** Deleting a question that is used in any published test MUST be blocked

### 4.3 Test Builder
- **FR-019** Contributors MUST be able to create tests with: title, description, category, duration, difficulty
- **FR-020** Test category MUST be one of: `tka_saintek`, `tka_soshum`, `smbt`
- **FR-021** Test status MUST be one of: `draft`, `published`
- **FR-022** Contributors MUST be able to select multiple questions from the shared pool into a test
- **FR-023** Question order within a test MUST be preserved (`order_index`)
- **FR-024** Every test MUST have exactly one scoring configuration with three values: `correct_points`, `wrong_points`, `blank_points` (decimals, can be negative or zero)
- **FR-025** Tests MUST be publishable and unpublishable
- **FR-026** Only published tests MUST be visible to students

### 4.4 Test Session (Student Taking Test)
- **FR-027** Starting a test MUST create a `test_session` with `started_at` timestamp and `status = in_progress`
- **FR-028** Server MUST enforce the test duration; sessions exceeding duration MUST auto-submit (`status = expired` or `status = submitted` based on whether student finalized)
- **FR-029** Students MUST be able to save an answer for any question (POST `/sessions/:id/answers`, idempotent upsert by question)
- **FR-030** Students MUST be able to flag/unflag any question for review (`is_flagged` boolean per answer)
- **FR-031** Students MUST be able to navigate between questions in any order
- **FR-032** Students MUST be able to resume an `in_progress` session if they close the browser
- **FR-033** Submit MUST require explicit confirmation when blank questions exist
- **FR-034** On submit, system MUST compute the result using the test's scoring formula and create a `test_result` record

### 4.5 Scoring
- **FR-035** Result computation MUST apply the test's specific scoring formula:
  ```
  total_score = (correct_count × correct_points) + (wrong_count × wrong_points) + (blank_count × blank_points)
  ```
- **FR-036** System MUST store the raw counts (correct, wrong, blank) alongside `total_score` for auditing
- **FR-037** Result computation MUST happen on submit, server-side, and be immutable thereafter

### 4.6 Results & Review
- **FR-038** Students MUST see their result immediately after submission, including: total score, percentage, correct/wrong/blank counts, per-topic breakdown
- **FR-039** Students MUST be able to view answer review: for each question, show their answer, the correct answer, and the explanation
- **FR-040** Answer review MUST support filtering by status: all, correct, wrong, blank
- **FR-041** Students MUST have access to a Score History showing all their past completed sessions

### 4.7 Leaderboard
- **FR-042** Leaderboard MUST display the top 100 students ranked by score, scoped by: global, TKA only, SMBT only, or this week
- **FR-043** Personal rank for users outside top 100 MUST be computed and displayed on the student dashboard
- **FR-044** Rank MUST be computed live from `test_results` using SQL `RANK() OVER (ORDER BY score DESC)`
- **FR-045** System MUST NOT store rank as a column (it is always derived)
- **FR-046** An index on `test_results(score DESC)` MUST exist from day one to keep rank queries fast

### 4.8 Super Admin Functions
- **FR-047** Super Admin MUST be able to view a list of all users with search and status filter
- **FR-048** Super Admin MUST be able to suspend or reactivate any user
- **FR-049** Super Admin MUST be able to view pending contributor applications and approve or reject them
- **FR-050** Super Admin MUST be able to view all tests across all contributors with their attempt counts
- **FR-051** Super Admin MUST be able to unpublish any test
- **FR-052** Super Admin dashboard MUST show platform-wide counters: total students, total contributors, total tests, total questions, recent activity, pending approvals

---

## 5. Tech Stack

| Layer | Choice | Reason |
|---|---|---|
| **Frontend** | Vue 3 (Composition API) | User preference |
| **Backend** | Golang | User preference |
| **HTTP Router** | gin or chi (TBD by implementer) | Stdlib-friendly, well-supported |
| **Database** | PostgreSQL 15+ | Relational data, RANK() window function support |
| **DB Driver** | pgx | Modern, performant pg driver for Go |
| **Migrations** | golang-migrate | Standard tooling |
| **Auth** | JWT (HS256) | Stateless, scales horizontally |
| **Password Hashing** | bcrypt | Industry standard |

---

## 6. Data Model

Ten entities, all with `uuid` primary keys and `timestamp` audit columns where indicated. Foreign keys are enforced.

### 6.1 `users`
| Field | Type | Notes |
|---|---|---|
| id | uuid (PK) | |
| name | string | |
| email | string | unique index |
| password_hash | string | bcrypt |
| role | enum | `student` \| `contributor` \| `admin` |
| status | enum | `active` \| `inactive` \| `suspended` \| `pending` |
| created_at, updated_at | timestamp | |

### 6.2 `topics`
| Field | Type | Notes |
|---|---|---|
| id | uuid (PK) | |
| name | string | unique |
| description | string | nullable |
| created_at | timestamp | |

### 6.3 `questions`
| Field | Type | Notes |
|---|---|---|
| id | uuid (PK) | |
| contributor_id | uuid (FK → users.id) | |
| topic_id | uuid (FK → topics.id) | |
| text | text | |
| explanation | text | nullable |
| difficulty | enum | `easy` \| `medium` \| `hard` |
| created_at, updated_at | timestamp | |

### 6.4 `question_options`
| Field | Type | Notes |
|---|---|---|
| id | uuid (PK) | |
| question_id | uuid (FK → questions.id) | cascade delete |
| label | char(1) | `A`–`E` |
| text | string | |
| is_correct | boolean | exactly one TRUE per question (DB constraint) |

### 6.5 `tests`
| Field | Type | Notes |
|---|---|---|
| id | uuid (PK) | |
| contributor_id | uuid (FK → users.id) | |
| title | string | |
| description | text | nullable |
| category | enum | `tka_saintek` \| `tka_soshum` \| `smbt` |
| duration_minutes | int | |
| difficulty | enum | `easy` \| `medium` \| `hard` |
| status | enum | `draft` \| `published` |
| created_at, published_at | timestamp | published_at nullable |

### 6.6 `test_questions` (junction)
| Field | Type | Notes |
|---|---|---|
| id | uuid (PK) | |
| test_id | uuid (FK → tests.id) | cascade delete |
| question_id | uuid (FK → questions.id) | |
| order_index | int | unique within a test |

### 6.7 `scoring_configs` (1:1 with tests)
| Field | Type | Notes |
|---|---|---|
| id | uuid (PK) | |
| test_id | uuid (FK → tests.id) | unique, cascade delete |
| correct_points | decimal(5,2) | e.g. 4.00 |
| wrong_points | decimal(5,2) | e.g. 0.00 or -1.00 |
| blank_points | decimal(5,2) | e.g. 0.00 |

### 6.8 `test_sessions`
| Field | Type | Notes |
|---|---|---|
| id | uuid (PK) | |
| student_id | uuid (FK → users.id) | |
| test_id | uuid (FK → tests.id) | |
| started_at | timestamp | |
| submitted_at | timestamp | nullable |
| status | enum | `in_progress` \| `submitted` \| `expired` |
| time_remaining_seconds | int | snapshot on save, used to resume |

### 6.9 `session_answers`
| Field | Type | Notes |
|---|---|---|
| id | uuid (PK) | |
| session_id | uuid (FK → test_sessions.id) | cascade delete |
| question_id | uuid (FK → questions.id) | |
| selected_option_id | uuid (FK → question_options.id) | nullable (blank answer) |
| is_flagged | boolean | default false |
| answered_at | timestamp | |
| Unique constraint | (session_id, question_id) | one answer per question per session |

### 6.10 `test_results` (denormalized for performance)
| Field | Type | Notes |
|---|---|---|
| id | uuid (PK) | |
| session_id | uuid (FK → test_sessions.id) | unique |
| student_id | uuid (FK → users.id) | indexed |
| test_id | uuid (FK → tests.id) | indexed |
| total_score | decimal(7,2) | indexed DESC for leaderboard |
| correct_count | int | |
| wrong_count | int | |
| blank_count | int | |
| completed_at | timestamp | |

**Required indexes:**
- `users(email)` unique
- `questions(contributor_id)`, `questions(topic_id)`
- `tests(contributor_id)`, `tests(status)`
- `test_sessions(student_id)`, `test_sessions(test_id, status)`
- `test_results(student_id)`, `test_results(test_id)`, `test_results(total_score DESC)`

---

## 7. API Contract

REST over HTTPS. Base path: `/api/v1`. Authenticated routes require `Authorization: Bearer <jwt>`. All collection endpoints support `?page=N&limit=M` pagination.

### 7.1 Auth
| Method | Path | Purpose | Auth |
|---|---|---|---|
| POST | `/auth/register` | Create user | Public |
| POST | `/auth/login` | Issue JWT | Public |
| POST | `/auth/logout` | Invalidate session (optional, JWT is stateless) | Any |
| GET | `/auth/me` | Current user profile | Any |

### 7.2 Topics
| Method | Path | Purpose | Auth |
|---|---|---|---|
| GET | `/topics` | List | Any |
| POST | `/topics` | Create | Admin |
| PATCH | `/topics/:id` | Update | Admin |
| DELETE | `/topics/:id` | Delete | Admin |

### 7.3 Questions
| Method | Path | Purpose | Auth |
|---|---|---|---|
| GET | `/questions?search=&topic_id=&difficulty=` | List with filters | Contributor, Admin |
| POST | `/questions` | Create | Contributor |
| GET | `/questions/:id` | Detail | Contributor, Admin |
| PATCH | `/questions/:id` | Update (own only) | Contributor, Admin |
| DELETE | `/questions/:id` | Delete (own only, blocked if in published test) | Contributor, Admin |

### 7.4 Tests
| Method | Path | Purpose | Auth |
|---|---|---|---|
| GET | `/tests?category=&difficulty=&status=` | List (scoped: student→published, contributor→own, admin→all) | Any |
| POST | `/tests` | Create | Contributor |
| GET | `/tests/:id` | Detail | Any (visibility-scoped) |
| PATCH | `/tests/:id` | Update (own only) | Contributor, Admin |
| DELETE | `/tests/:id` | Delete (own only, draft only) | Contributor, Admin |
| POST | `/tests/:id/publish` | Publish | Contributor (own), Admin |
| POST | `/tests/:id/unpublish` | Unpublish | Contributor (own), Admin |

### 7.5 Test Sessions
| Method | Path | Purpose | Auth |
|---|---|---|---|
| POST | `/tests/:id/sessions` | Start a new session | Student |
| GET | `/sessions/:id` | Resume / fetch state | Student (own) |
| POST | `/sessions/:id/answers` | Upsert answer `{question_id, selected_option_id?}` | Student (own) |
| POST | `/sessions/:id/flag` | Toggle flag `{question_id, is_flagged}` | Student (own) |
| POST | `/sessions/:id/submit` | Finalize → produces result | Student (own) |

### 7.6 Results
| Method | Path | Purpose | Auth |
|---|---|---|---|
| GET | `/results` | List own results (student) or all (admin) | Student, Admin |
| GET | `/results/:id` | Score breakdown + per-topic stats | Student (own), Admin |
| GET | `/results/:id/review?status=all\|correct\|wrong\|blank` | Per-question review with correct answer + explanation | Student (own), Admin |

### 7.7 Leaderboard
| Method | Path | Purpose | Auth |
|---|---|---|---|
| GET | `/leaderboard?scope=global\|tka\|smbt\|week` | Top 100 ranked students | Any |
| GET | `/leaderboard/me` | Caller's personal rank | Student |

### 7.8 Admin
| Method | Path | Purpose | Auth |
|---|---|---|---|
| GET | `/admin/users?search=&status=` | List users | Admin |
| PATCH | `/admin/users/:id` | Update status/role | Admin |
| GET | `/admin/contributors?status=` | List contributors (filter by status) | Admin |
| POST | `/admin/contributors/:id/approve` | Approve pending | Admin |
| POST | `/admin/contributors/:id/reject` | Reject pending | Admin |
| GET | `/admin/tests` | All tests across contributors | Admin |
| GET | `/admin/stats` | Platform counters | Admin |

---

## 8. Backend Project Structure (Go)

```
tkaprep-backend/
├── cmd/server/main.go              # entrypoint, wiring
├── internal/
│   ├── config/                     # env loader
│   ├── domain/                     # pure entities, zero deps
│   │   ├── user/
│   │   ├── question/
│   │   ├── test/
│   │   ├── session/
│   │   └── result/
│   ├── repository/
│   │   ├── interfaces.go           # repo contracts
│   │   └── postgres/               # pgx implementations
│   ├── service/                    # business logic
│   │   ├── auth/
│   │   ├── question/
│   │   ├── test/
│   │   ├── session/
│   │   ├── result/
│   │   └── leaderboard/
│   ├── handler/
│   │   ├── http/                   # HTTP handlers
│   │   └── middleware/             # auth, RBAC, logging
│   └── pkg/                        # internal utilities
│       ├── jwt/
│       ├── validator/
│       └── errors/
├── migrations/                     # golang-migrate SQL files
├── api/openapi.yaml                # generated OpenAPI spec
├── Makefile
├── docker-compose.yml              # postgres for local dev
├── go.mod
└── README.md
```

### Layer responsibilities
- **domain** — entities, value objects, domain-specific errors. No imports from other internal packages.
- **repository** — `interfaces.go` defines repo contracts consumed by services. `postgres/` implements them.
- **service** — orchestrates domain + repositories. All business rules live here.
- **handler/http** — translates HTTP ↔ service calls. No business logic.
- **handler/middleware** — JWT validation, role checks, request logging.

---

## 9. Frontend Project Structure (Vue 3)

```
tkaprep-frontend/
├── src/
│   ├── views/                      # route-level components
│   │   ├── auth/                   # Landing, Login, Register
│   │   ├── student/                # Dashboard, BrowseTests, TestSession, Result, AnswerReview, ScoreHistory, Leaderboard
│   │   ├── contributor/            # Dashboard, QuestionBank, TestBuilder, MyTests, TestAnalytics
│   │   └── admin/                  # Dashboard, Users, Contributors, AllTests
│   ├── components/                 # shared UI primitives (Card, Badge, Button, Modal, etc.)
│   ├── composables/                # useAuth, useApi, useSession
│   ├── stores/                     # Pinia stores: auth, session, leaderboard
│   ├── router/                     # vue-router with role-based guards
│   ├── api/                        # API client wrappers per module
│   └── main.ts
├── public/
├── vite.config.ts
└── package.json
```

---

## 10. Cross-Cutting Concerns

### 10.1 Validation
- Backend MUST validate all incoming payloads (required fields, types, enum values)
- 400 Bad Request on validation failure with structured error body

### 10.2 Error Format
Standard error response:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable description",
    "details": [{ "field": "email", "issue": "must be valid email" }]
  }
}
```

### 10.3 Rate Limiting
- Auth endpoints (login, register): 5 requests per minute per IP
- All other endpoints: 100 requests per minute per authenticated user
- Implemented via middleware (Redis-backed if available, in-memory otherwise for v1)

### 10.4 Logging
- Structured JSON logs (zerolog or zap)
- Log every request with: method, path, status, duration, user_id (if authed)

### 10.5 Configuration
All via env vars: `DATABASE_URL`, `JWT_SECRET`, `JWT_EXPIRY`, `PORT`, `LOG_LEVEL`.

---

## 11. Acceptance Criteria — v1 Definition of Done

A v1 release is complete when:
1. A new student can register, log in, take a test, submit, see the result, review answers, and view their score in history
2. A contributor can register (pending), be approved by an admin, log in, create questions, build a test with custom scoring formula, and publish it
3. A super admin can log in via seeded credentials, view all users, approve/reject pending contributors, and unpublish a test
4. Leaderboard correctly shows top 100 by score and a student outside top 100 sees their personal rank on the dashboard
5. All 10 database tables exist with the specified columns, foreign keys, and indexes
6. All API endpoints listed in Section 7 are implemented and return appropriate responses for unauthorized/forbidden access
7. Timer is enforced server-side; submitting after duration expires is rejected or marks the session expired
8. Scoring formula is applied correctly per test (verified by integration test with known inputs)

---

## 12. Future Iterations (Backlog)

- Question visibility rules (private/public flag, contributor ownership)
- Test visibility scoping (per-class, per-group)
- Focus mode / anti-cheat (tab lock, fullscreen enforcement)
- Freemium gating (free tier limits, paid plans)
- Discussion threads on questions
- AI-assisted question authoring
- Mobile apps (React Native or Flutter)
- Multi-language support (English in addition to Bahasa Indonesia)
- Performance: introduce `leaderboard_snapshots` table when user base exceeds ~50k

---

## 13. Review Checklist

- [x] User scenarios cover all three roles
- [x] All functional requirements have unique IDs
- [x] Data model is complete with constraints and indexes
- [x] API endpoints cover all functional requirements
- [x] Out-of-scope items are explicit
- [x] Acceptance criteria are testable
- [x] No implementation details leak into requirements (e.g. no specific gin/chi choice in FRs)
- [x] Pagination addressed for all collection endpoints
- [x] Error format defined
- [x] Authentication and RBAC model defined
