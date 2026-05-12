CREATE TYPE question_type AS ENUM ('mcq', 'true_false', 'multi_correct');

ALTER TABLE questions
    ADD COLUMN question_type question_type NOT NULL DEFAULT 'mcq';

-- The old constraint allowed only one correct option per question.
-- Drop it — multi_correct questions need multiple correct options.
DROP INDEX idx_question_options_correct;
