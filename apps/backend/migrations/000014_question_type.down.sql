ALTER TABLE questions DROP COLUMN question_type;
DROP TYPE question_type;

CREATE UNIQUE INDEX idx_question_options_correct
    ON question_options(question_id)
    WHERE is_correct = true;
