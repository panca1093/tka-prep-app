CREATE TABLE scoring_configs (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id        UUID         NOT NULL UNIQUE REFERENCES tests(id) ON DELETE CASCADE,
    correct_points DECIMAL(5,2) NOT NULL DEFAULT 4.00,
    wrong_points   DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    blank_points   DECIMAL(5,2) NOT NULL DEFAULT 0.00
);
