# Brief: TKAPrep — Scoring Overhaul + Contributor UX Improvements

## Problem

Current scoring uses a weighted points system (`correct_points`, `wrong_points`, `blank_points` per test config) which is opaque to students and inconsistently configured by contributors. There is no standardized baseline score, no statistical insight into question quality, and contributors can't preview how their questions render to students or easily fix miscategorized questions.

## Idea

**Four improvements in one branch:**

1. **Standardized base scoring** — Replace the freeform weighted config with a simple percentage score: `score = (correct / total) × 100`. Each correct answer = 1 point; score is a percentage. Recalculate all existing `test_results` rows with this formula.

2. **Optional IRT score display** — After a student submits a test, compute an IRT-estimated ability score (theta) alongside the base score, visible only to contributors on the result analytics panel. IRT parameters (difficulty `b`, discrimination `a`, guessing `c`) are auto-estimated from accumulated student answer data (item-level frequency stats). Start with 1PL/Rasch (difficulty only) since data is sparse early on; upgrade to 3PL when enough data exists.

3. **Question preview modal** — On the contributor question detail/list page, a "Preview" button opens a modal that renders the question exactly as a student would see it during a test session: question text, answer options (with image/formula support), no correct answer revealed.

4. **Contributor category management** — Contributors can: (a) edit category metadata (name, description) for categories they own, and (b) reassign a question to a different category.

## Users

- **Students** — see a cleaner, standardized percentage score after submitting a test.
- **Contributors** — see the IRT-estimated ability score on their test result analytics; can preview questions before publishing; can fix category mistakes.
- **Super Admin** — no direct interaction, but benefits from better data quality.

## Success Criteria

- Submitting a test returns `score` as a percentage (0–100), stored in `test_results.total_score`.
- All existing `test_results` rows are backfilled with the new formula (migration).
- `scoring_configs` table is either removed or its `correct_points/wrong_points/blank_points` columns are deprecated (kept for audit, ignored in calculation).
- On the contributor result analytics page, an "IRT Score (θ)" column is visible showing the estimated theta for each student who took the test, computed from item-level response data.
- Clicking "Preview" on any question in the contributor UI opens a modal showing the student-facing question view.
- Contributor can edit a category's name/description from the category list/detail page.
- Contributor can change a question's category from the question edit form.

## Constraints

- Backend: Go 1.22 + chi + pgx; clean layer architecture (handler → service → repository → domain).
- Frontend: Vue 3 + TypeScript + Vite + Pinia.
- DB: PostgreSQL 16; all schema changes via golang-migrate numbered migrations.
- IRT: Start simple — 1PL Rasch model (`theta = difficulty` estimate from % correct). Full 3PL is future work. The IRT score is computed server-side, never client-side.
- Existing `test_results` must be recalculated — this is a data migration, not just a schema change.
- `scoring_configs` fields are not deleted (audit trail) but ignored by the new formula.
- Category editing: contributors can only edit categories they own (created by them or assigned to them).
- No breaking changes to the student-facing test-taking flow UX — only the score display changes.
