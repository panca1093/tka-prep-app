# Feedback — Quiz/Test Platform

Source: WhatsApp chat, 15 May 2026  
Reporter: Family Ka Deen  

---

## Bugs

### BUG-01 — Unpublish button throws an error
When a test that has already been created is clicked "Unpublish", an error occurs.  
**Expected:** The test status reverts to draft/unpublished without error.

---

## Feature Requests & Changes

### FR-01 — PGK (Multiple-Select) questions must use checkboxes, not A/B/C labels
For question type **PGK** (Pilihan Ganda Kompleks / Multiple Select), the answer choices displayed to students during the test must render as **checkbox inputs**, not alphabetical labels (A, B, C, D).  
Regular multiple-choice (PG) questions keep their current A/B/C display.

---

### FR-02 — Default answer options count = 4, with the ability to add more
For both **PG** (single-select multiple choice) and **PGK** (multiple-select) question types, the default number of answer choices when creating a question must be **4**.  
The contributor must be able to **add extra choices** beyond 4 if needed.

---

### FR-03 — Students may only attempt a test once
Once a student submits a test (or the test session ends), they **cannot retake** that same test.  
The system must block any further attempt by the same student on the same test package.

---

### FR-04 — Prevent screenshots during a test
While a student is taking a test, the app must **block or detect screenshots** to prevent questions from being distributed outside the platform.  
This is especially important for tests that cannot be run simultaneously (non-concurrent exams).

---

### FR-05 — Leaderboard / results must be per-test and per-subject, not combined
Currently the leaderboard averages results across multiple tests.  
**Required behaviour:**
- Show each student's score **per test** separately.
- Show scores **per subject** within each test.
- Do **not** aggregate or average results across different tests.

---

### FR-06 — Add education-level tag on individual questions
When a contributor creates a question in the **bank soal (question bank)**, they must select an **education level**:
- **SD** (Primary)
- **SMP** (Junior High)
- **SMA** (Senior High)

The purpose is to prevent SD-level questions from appearing in SMP/SMA test packages, and vice versa.

---

## Deferred / To Be Clarified Later

- **Question categories** — The reporter mentioned adding categories to questions in the contributor's question-creation screen. Details to be provided separately by the reporter.
- **Test package subject composition (FR-07)** — Structure of subjects per package type and level. To be confirmed by reporter.
