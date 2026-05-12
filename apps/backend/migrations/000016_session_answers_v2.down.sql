DROP INDEX IF EXISTS session_answers_option_uq;
DROP INDEX IF EXISTS session_answers_stmt_uq;

ALTER TABLE session_answers
    DROP COLUMN IF EXISTS statement_id,
    DROP COLUMN IF EXISTS boolean_answer;

ALTER TABLE session_answers ADD COLUMN is_flagged BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE session_answers ADD CONSTRAINT session_answers_session_id_question_id_key UNIQUE (session_id, question_id);

DROP TABLE IF EXISTS session_question_flags;
