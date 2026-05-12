CREATE TYPE difficulty AS ENUM ('easy', 'medium', 'hard');

CREATE TABLE questions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contributor_id UUID NOT NULL REFERENCES users(id),
    topic_id       UUID NOT NULL REFERENCES topics(id),
    text           TEXT NOT NULL,
    explanation    TEXT,
    difficulty     difficulty NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_questions_contributor_id ON questions(contributor_id);
CREATE INDEX idx_questions_topic_id ON questions(topic_id);
