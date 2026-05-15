ALTER TABLE test_results ADD CONSTRAINT test_results_student_test_unique UNIQUE (student_id, test_id);
