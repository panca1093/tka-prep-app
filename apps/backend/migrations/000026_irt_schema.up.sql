CREATE TABLE question_irt_params (
    question_id   UUID         PRIMARY KEY REFERENCES questions(id) ON DELETE CASCADE,
    correct_count INT          NOT NULL DEFAULT 0,
    sample_size   INT          NOT NULL DEFAULT 0,
    difficulty_b  DECIMAL(6,4) NOT NULL DEFAULT 0.0,
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

ALTER TABLE test_results ADD COLUMN irt_theta DECIMAL(6,4);
