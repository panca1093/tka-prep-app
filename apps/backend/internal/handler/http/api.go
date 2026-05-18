package http

import (
	adminsvc "github.com/yourorg/tkaprep/apps/backend/internal/service/admin"
	"github.com/yourorg/tkaprep/apps/backend/internal/service/auth"
	categorysvc "github.com/yourorg/tkaprep/apps/backend/internal/service/category"
	leadersvc "github.com/yourorg/tkaprep/apps/backend/internal/service/leaderboard"
	"github.com/yourorg/tkaprep/apps/backend/internal/service/question"
	resultsvc "github.com/yourorg/tkaprep/apps/backend/internal/service/result"
	sessionsvc "github.com/yourorg/tkaprep/apps/backend/internal/service/session"
	testsvc "github.com/yourorg/tkaprep/apps/backend/internal/service/test"
	"github.com/yourorg/tkaprep/apps/backend/internal/service/topic"
)

// APIServer implements api.StrictServerInterface.
// Each domain area has its methods defined in a separate file (health.go, auth.go, topics.go, questions.go, tests.go, sessions.go, results.go, leaderboard.go, admin.go, category.go).
type APIServer struct {
	authSvc        *auth.Service
	topicSvc       *topic.Service
	questionSvc    *question.Service
	testSvc        *testsvc.Service
	sessionSvc     *sessionsvc.Service
	resultSvc      *resultsvc.Service
	leaderboardSvc *leadersvc.Service
	adminSvc       *adminsvc.Service
	categorySvc    *categorysvc.Service
}

func NewAPIServer(
	authSvc *auth.Service,
	topicSvc *topic.Service,
	questionSvc *question.Service,
	testSvc *testsvc.Service,
	sessionSvc *sessionsvc.Service,
	resultSvc *resultsvc.Service,
	leaderboardSvc *leadersvc.Service,
	adminSvc *adminsvc.Service,
	categorySvc *categorysvc.Service,
) *APIServer {
	return &APIServer{
		authSvc:        authSvc,
		topicSvc:       topicSvc,
		questionSvc:    questionSvc,
		testSvc:        testSvc,
		sessionSvc:     sessionSvc,
		resultSvc:      resultSvc,
		leaderboardSvc: leaderboardSvc,
		adminSvc:       adminSvc,
		categorySvc:    categorySvc,
	}
}
