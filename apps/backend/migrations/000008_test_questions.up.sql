CREATE TABLE test_questions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id     UUID NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id),
    order_index INT  NOT NULL,
    UNIQUE (test_id, order_index),
    UNIQUE (test_id, question_id)
);
