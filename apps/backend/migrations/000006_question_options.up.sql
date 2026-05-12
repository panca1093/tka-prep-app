CREATE TABLE question_options (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    label       CHAR(1) NOT NULL CHECK (label IN ('A','B','C','D','E')),
    text        TEXT NOT NULL,
    is_correct  BOOLEAN NOT NULL DEFAULT false,
    UNIQUE (question_id, label)
);

-- Enforce exactly one correct option per question
CREATE UNIQUE INDEX idx_question_options_correct
    ON question_options(question_id)
    WHERE is_correct = true;
