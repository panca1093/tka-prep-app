CREATE TYPE session_status AS ENUM ('in_progress', 'submitted', 'expired');

CREATE TABLE test_sessions (
    id                     UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id             UUID          NOT NULL REFERENCES users(id),
    test_id                UUID          NOT NULL REFERENCES tests(id),
    started_at             TIMESTAMPTZ   NOT NULL DEFAULT now(),
    submitted_at           TIMESTAMPTZ,
    status                 session_status NOT NULL DEFAULT 'in_progress',
    time_remaining_seconds INT           NOT NULL
);

CREATE INDEX test_sessions_student_id_idx ON test_sessions(student_id);
CREATE INDEX test_sessions_test_status_idx ON test_sessions(test_id, status);
