package domain

type PlatformStats struct {
	TotalStudents     int
	TotalContributors int
	TotalTests        int
	TotalQuestions    int
	PendingApprovals  int
}

type TestWithAttempts struct {
	Test
	AttemptCount int
}
