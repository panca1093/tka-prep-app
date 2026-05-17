package domain

type PlatformStats struct {
	TotalStudents        int
	TotalContributors    int
	TotalTests           int
	TotalQuestions       int
	PendingApprovals     int
	StudentsThisWeek     int
	ContributorsThisWeek int
	TotalAttempts        int
	AvgScore             float64
	TopTestTitle         string
	TopTestAttempts      int
	QuestionsTKASaintek  int
	QuestionsTKASoshum   int
	QuestionsSMBT        int
	QuestionsUsed        int
	QuestionsUnused      int
}

type TestWithAttempts struct {
	Test
	AttemptCount int
}
