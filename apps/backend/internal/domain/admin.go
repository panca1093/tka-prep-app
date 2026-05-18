package domain

import "time"

type AttemptsByTopic struct {
	Topic string
	SD    int
	SMP   int
	SMA   int
	SMK   int
}

type TopicPerformance struct {
	Topic     string
	AvgScore  float64
	BestScore float64
}

type TopicQuestionCount struct {
	Topic  string
	Total  int
	Used   int
	Unused int
}

type QuestionCountGroup struct {
	EducationLevel string
	Topics         []TopicQuestionCount
}

type ContributorProductivity struct {
	Name          string
	Email         string
	QuestionCount int
	TestCount     int
	OutputScore   int
}

type TestCompletion struct {
	TestTitle     string
	Started       int
	Submitted     int
	Expired       int
	CompletionPct float64
}

type ActivityFeedEntry struct {
	EventType string
	ActorName string
	Detail    string
	Timestamp time.Time
}

type PlatformStats struct {
	TotalStudents           int
	TotalContributors       int
	TotalTests              int
	TotalQuestions          int
	PendingApprovals        int
	StudentsThisWeek        int
	ContributorsThisWeek    int
	AttemptsByTopic         []AttemptsByTopic
	TopicPerformance        []TopicPerformance
	QuestionCounts          []QuestionCountGroup
	ContributorProductivity []ContributorProductivity
	TestCompletion          []TestCompletion
	ActivityFeed            []ActivityFeedEntry
}

type TestWithAttempts struct {
	Test
	AttemptCount int
}
