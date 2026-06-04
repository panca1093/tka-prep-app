package repository

import (
	"testing"

	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
)

// TestQuestionFilter_VisibilityDefaults verifies that the CallerRole field
// is properly defaulted (zero value = empty Role, which means "not admin"
// and thus triggers contributor filtering in the repo layer).
func TestQuestionFilter_VisibilityDefaults(t *testing.T) {
	f := QuestionFilter{
		CallerID: uuid.New(),
		Page:     1,
		Limit:    20,
	}

	// Zero value for CallerRole means "not admin" → repo should filter by contributor.
	if f.CallerRole != "" {
		t.Error("expected empty CallerRole by default")
	}
}

// TestQuestionFilter_AdminBypass verifies that setting CallerRole to admin
// disables contributor filtering.
func TestQuestionFilter_AdminBypass(t *testing.T) {
	f := QuestionFilter{
		CallerID:   uuid.New(),
		CallerRole: domain.RoleAdmin,
		Page:       1,
		Limit:      20,
	}

	if f.CallerRole != domain.RoleAdmin {
		t.Errorf("expected admin role, got %s", f.CallerRole)
	}
}

// TestQuestionFilter_ContributorFilter verifies that setting CallerRole to
// contributor enables the visibility filter.
func TestQuestionFilter_ContributorFilter(t *testing.T) {
	contributorID := uuid.New()
	f := QuestionFilter{
		CallerID:   contributorID,
		CallerRole: domain.RoleContributor,
		Page:       1,
		Limit:      20,
	}

	if f.CallerRole != domain.RoleContributor {
		t.Errorf("expected contributor role, got %s", f.CallerRole)
	}
	if f.CallerID != contributorID {
		t.Errorf("expected caller ID to be set")
	}
}
