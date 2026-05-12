CREATE TYPE test_category AS ENUM ('tka_saintek', 'tka_soshum', 'smbt');
CREATE TYPE test_status   AS ENUM ('draft', 'published');

CREATE TABLE tests (
    id               UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    contributor_id   UUID            NOT NULL REFERENCES users(id),
    title            TEXT            NOT NULL,
    description      TEXT,
    category         test_category   NOT NULL,
    duration_minutes INT             NOT NULL CHECK (duration_minutes > 0),
    difficulty       difficulty NOT NULL,
    status           test_status     NOT NULL DEFAULT 'draft',
    created_at       TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    published_at     TIMESTAMPTZ
);

CREATE INDEX tests_contributor_idx ON tests (contributor_id);
CREATE INDEX tests_status_idx      ON tests (status);
