CREATE TABLE question_statements (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID        NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    text        TEXT        NOT NULL,
    is_correct  BOOLEAN     NOT NULL,
    position    INT         NOT NULL DEFAULT 0
);

CREATE INDEX idx_question_statements_question_id ON question_statements(question_id);
