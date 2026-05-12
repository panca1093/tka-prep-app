-- Separate flag storage so one flag row exists per question regardless of answer type.
CREATE TABLE session_question_flags (
    session_id  UUID    NOT NULL REFERENCES test_sessions(id) ON DELETE CASCADE,
    question_id UUID    NOT NULL REFERENCES questions(id),
    is_flagged  BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (session_id, question_id)
);

-- Migrate existing flags before altering session_answers.
INSERT INTO session_question_flags (session_id, question_id, is_flagged)
SELECT session_id, question_id, is_flagged
FROM session_answers
WHERE is_flagged = true
ON CONFLICT DO NOTHING;

-- Remove flag-only rows (blank answer + no option chosen) — they only existed for flag tracking.
DELETE FROM session_answers WHERE selected_option_id IS NULL;

-- Drop old single-answer-per-question constraint.
ALTER TABLE session_answers DROP CONSTRAINT session_answers_session_id_question_id_key;
ALTER TABLE session_answers DROP COLUMN is_flagged;

-- New columns for multi_correct / true_false answers.
ALTER TABLE session_answers
    ADD COLUMN statement_id    UUID REFERENCES question_statements(id),
    ADD COLUMN boolean_answer  BOOLEAN;

-- Unique constraints: one row per (session, question, option) and per (session, question, statement).
CREATE UNIQUE INDEX session_answers_option_uq
    ON session_answers(session_id, question_id, selected_option_id)
    WHERE selected_option_id IS NOT NULL;

CREATE UNIQUE INDEX session_answers_stmt_uq
    ON session_answers(session_id, question_id, statement_id)
    WHERE statement_id IS NOT NULL;
