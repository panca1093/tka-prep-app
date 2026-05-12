package http

import (
	"github.com/yourorg/tkaprep/apps/backend/internal/service/auth"
	"github.com/yourorg/tkaprep/apps/backend/internal/service/question"
	"github.com/yourorg/tkaprep/apps/backend/internal/service/topic"
)

// APIServer implements api.StrictServerInterface.
// Each domain area has its methods defined in a separate file (health.go, auth.go, topics.go, questions.go, …).
type APIServer struct {
	authSvc     *auth.Service
	topicSvc    *topic.Service
	questionSvc *question.Service
}

func NewAPIServer(authSvc *auth.Service, topicSvc *topic.Service, questionSvc *question.Service) *APIServer {
	return &APIServer{
		authSvc:     authSvc,
		topicSvc:    topicSvc,
		questionSvc: questionSvc,
	}
}
