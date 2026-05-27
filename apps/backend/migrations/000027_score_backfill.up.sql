-- Recalculate total_score as percentage for all existing test results.
-- New formula: (correct_count / total_questions) * 100, rounded to 2 decimal places.
-- Rows where total is 0 (should not exist, but guard) are left at 0.
UPDATE test_results
SET total_score = ROUND(
    correct_count::DECIMAL / NULLIF(correct_count + wrong_count + blank_count, 0) * 100,
    2
)
WHERE correct_count + wrong_count + blank_count > 0;
