CREATE TABLE test_results (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id    UUID          NOT NULL UNIQUE REFERENCES test_sessions(id),
    student_id    UUID          NOT NULL REFERENCES users(id),
    test_id       UUID          NOT NULL REFERENCES tests(id),
    total_score   DECIMAL(7,2)  NOT NULL,
    correct_count INT           NOT NULL,
    wrong_count   INT           NOT NULL,
    blank_count   INT           NOT NULL,
    completed_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX test_results_student_id_idx ON test_results(student_id);
CREATE INDEX test_results_test_id_idx    ON test_results(test_id);
CREATE INDEX test_results_score_idx      ON test_results(total_score DESC);
