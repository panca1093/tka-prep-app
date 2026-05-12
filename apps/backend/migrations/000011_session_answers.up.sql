CREATE TABLE session_answers (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id         UUID        NOT NULL REFERENCES test_sessions(id) ON DELETE CASCADE,
    question_id        UUID        NOT NULL REFERENCES questions(id),
    selected_option_id UUID        REFERENCES question_options(id),
    is_flagged         BOOLEAN     NOT NULL DEFAULT false,
    answered_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (session_id, question_id)
);

CREATE INDEX session_answers_session_id_idx ON session_answers(session_id);
