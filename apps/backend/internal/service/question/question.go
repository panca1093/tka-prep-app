package question

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/storage"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

type Service struct {
	questions repository.QuestionRepository
	store     storage.FileStorage
}

func NewService(questions repository.QuestionRepository, store storage.FileStorage) *Service {
	return &Service{questions: questions, store: store}
}

type OptionInput struct {
	Label     string
	Text      string
	IsCorrect bool
	ImageURL  *string
}

type StatementInput struct {
	Text      string
	IsCorrect bool
	Position  int
	ImageURL  *string
}

type CreateInput struct {
	ContributorID  uuid.UUID
	TopicID        uuid.UUID
	QuestionType   domain.QuestionType
	EducationLevel *string
	Text           string
	Explanation    *string
	ImageURL       *string
	Difficulty     domain.Difficulty
	Options        []OptionInput
	Statements     []StatementInput
}

type UpdateInput struct {
	TopicID        *uuid.UUID
	EducationLevel *string
	Text           *string
	Explanation    *string
	ImageURL       *string
	Difficulty     *domain.Difficulty
	Options        []OptionInput
	Statements     []StatementInput
}

type ListFilter struct {
	Search         string
	TopicID        *uuid.UUID
	Difficulty     *domain.Difficulty
	QuestionType   *domain.QuestionType
	EducationLevel *string
	CallerID       uuid.UUID
	CallerRole     domain.Role
	Page           int
	Limit          int
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]*domain.Question, int, error) {
	return s.questions.List(ctx, repository.QuestionFilter{
		Search:         f.Search,
		TopicID:        f.TopicID,
		Difficulty:     f.Difficulty,
		QuestionType:   f.QuestionType,
		EducationLevel: f.EducationLevel,
		CallerID:       f.CallerID,
		CallerRole:     f.CallerRole,
		Page:           f.Page,
		Limit:          f.Limit,
	})
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.Question, error) {
	in.Text = strings.TrimSpace(in.Text)
	if in.Text == "" {
		return nil, fmt.Errorf("%w: question text is required", apierr.ErrValidation)
	}

	qt := in.QuestionType
	if qt == "" {
		qt = domain.QuestionTypeMCQ
	}

	now := time.Now().UTC()
	q := &domain.Question{
		ID:             uuid.New(),
		ContributorID:  in.ContributorID,
		TopicID:        in.TopicID,
		Type:           qt,
		EducationLevel: in.EducationLevel,
		Text:           SanitizeHTML(in.Text),
		Explanation:    sanitizeOptional(in.Explanation),
		ImageURL:       in.ImageURL,
		Difficulty:     in.Difficulty,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	switch qt {
	case domain.QuestionTypeMCQ:
		if err := validateMCQOptions(in.Options); err != nil {
			return nil, err
		}
		for _, o := range in.Options {
			q.Options = append(q.Options, domain.QuestionOption{ID: uuid.New(), Label: o.Label, Text: SanitizeHTML(o.Text), IsCorrect: o.IsCorrect, ImageURL: o.ImageURL})
		}
	case domain.QuestionTypeMultiCorrect:
		if err := validatePGKOptions(in.Options); err != nil {
			return nil, err
		}
		for _, o := range in.Options {
			q.Options = append(q.Options, domain.QuestionOption{ID: uuid.New(), Label: o.Label, Text: SanitizeHTML(o.Text), IsCorrect: o.IsCorrect, ImageURL: o.ImageURL})
		}
	case domain.QuestionTypeTrueFalse:
		if err := validateStatements(in.Statements); err != nil {
			return nil, err
		}
		for _, st := range in.Statements {
			q.Statements = append(q.Statements, domain.QuestionStatement{ID: uuid.New(), Text: SanitizeHTML(st.Text), IsCorrect: st.IsCorrect, Position: st.Position, ImageURL: st.ImageURL})
		}
	}

	if err := s.questions.Create(ctx, q); err != nil {
		return nil, fmt.Errorf("create question: %w", err)
	}
	return q, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole domain.Role) (*domain.Question, error) {
	q, err := s.questions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Contributors can only view their own questions. Admins and students (test-taking) bypass.
	if callerRole == domain.RoleContributor && q.ContributorID != callerID {
		return nil, apierr.ErrNotFound
	}
	return q, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole domain.Role, in UpdateInput) (*domain.Question, error) {
	q, err := s.questions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if callerRole != domain.RoleAdmin && q.ContributorID != callerID {
		return nil, apierr.ErrForbidden
	}

	// Capture old text content for orphan image cleanup.
	oldContent := collectAllText(q)

	if in.Text != nil {
		text := strings.TrimSpace(*in.Text)
		if text == "" {
			return nil, fmt.Errorf("%w: question text cannot be empty", apierr.ErrValidation)
		}
		q.Text = SanitizeHTML(text)
	}
	if in.TopicID != nil {
		q.TopicID = *in.TopicID
	}
	if in.EducationLevel != nil {
		q.EducationLevel = in.EducationLevel
	}
	if in.Explanation != nil {
		q.Explanation = sanitizeOptional(in.Explanation)
	}
	if in.ImageURL != nil {
		q.ImageURL = in.ImageURL
	}
	if in.Difficulty != nil {
		q.Difficulty = *in.Difficulty
	}

	switch q.Type {
	case domain.QuestionTypeMCQ:
		if len(in.Options) > 0 {
			if err := validateMCQOptions(in.Options); err != nil {
				return nil, err
			}
			q.Options = make([]domain.QuestionOption, len(in.Options))
			for i, o := range in.Options {
				q.Options[i] = domain.QuestionOption{ID: uuid.New(), Label: o.Label, Text: SanitizeHTML(o.Text), IsCorrect: o.IsCorrect, ImageURL: o.ImageURL}
			}
		}
	case domain.QuestionTypeMultiCorrect:
		if len(in.Options) > 0 {
			if err := validatePGKOptions(in.Options); err != nil {
				return nil, err
			}
			q.Options = make([]domain.QuestionOption, len(in.Options))
			for i, o := range in.Options {
				q.Options[i] = domain.QuestionOption{ID: uuid.New(), Label: o.Label, Text: SanitizeHTML(o.Text), IsCorrect: o.IsCorrect, ImageURL: o.ImageURL}
			}
		}
	case domain.QuestionTypeTrueFalse:
		if len(in.Statements) > 0 {
			if err := validateStatements(in.Statements); err != nil {
				return nil, err
			}
			q.Statements = make([]domain.QuestionStatement, len(in.Statements))
			for i, st := range in.Statements {
				q.Statements[i] = domain.QuestionStatement{ID: uuid.New(), Text: SanitizeHTML(st.Text), IsCorrect: st.IsCorrect, Position: st.Position, ImageURL: st.ImageURL}
			}
		}
	}

	q.UpdatedAt = time.Now().UTC()
	if err := s.questions.Update(ctx, q); err != nil {
		return nil, fmt.Errorf("update question: %w", err)
	}

	// Clean up orphan images from editing.
	s.cleanupOrphanImages(oldContent, collectAllText(q))
	return q, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole domain.Role) error {
	q, err := s.questions.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if callerRole != domain.RoleAdmin && q.ContributorID != callerID {
		return apierr.ErrForbidden
	}
	used, err := s.questions.IsUsedInPublishedTest(ctx, id)
	if err != nil {
		return fmt.Errorf("check question usage: %w", err)
	}
	if used {
		return apierr.ErrConflict
	}
	return s.questions.Delete(ctx, id)
}

func validateMCQOptions(opts []OptionInput) error {
	if len(opts) < 4 || len(opts) > 5 {
		return fmt.Errorf("%w: 4-5 options required", apierr.ErrValidation)
	}
	labels := map[string]bool{}
	correct := 0
	for _, o := range opts {
		l := strings.ToUpper(strings.TrimSpace(o.Label))
		if len(l) != 1 || !strings.ContainsAny(l, "ABCDE") {
			return fmt.Errorf("%w: option label must be A-E", apierr.ErrValidation)
		}
		if labels[l] {
			return fmt.Errorf("%w: duplicate label %s", apierr.ErrValidation, l)
		}
		labels[l] = true
		if strings.TrimSpace(o.Text) == "" {
			return fmt.Errorf("%w: option text cannot be empty", apierr.ErrValidation)
		}
		if o.IsCorrect {
			correct++
		}
	}
	if correct != 1 {
		return fmt.Errorf("%w: exactly one option must be correct for MCQ", apierr.ErrValidation)
	}
	return nil
}

func validatePGKOptions(opts []OptionInput) error {
	if len(opts) < 4 || len(opts) > 5 {
		return fmt.Errorf("%w: 4-5 options required", apierr.ErrValidation)
	}
	labels := map[string]bool{}
	correct := 0
	for _, o := range opts {
		l := strings.ToUpper(strings.TrimSpace(o.Label))
		if len(l) != 1 || !strings.ContainsAny(l, "ABCDE") {
			return fmt.Errorf("%w: option label must be A-E", apierr.ErrValidation)
		}
		if labels[l] {
			return fmt.Errorf("%w: duplicate label %s", apierr.ErrValidation, l)
		}
		labels[l] = true
		if strings.TrimSpace(o.Text) == "" {
			return fmt.Errorf("%w: option text cannot be empty", apierr.ErrValidation)
		}
		if o.IsCorrect {
			correct++
		}
	}
	if correct < 1 {
		return fmt.Errorf("%w: at least one option must be correct for PGK", apierr.ErrValidation)
	}
	return nil
}

func validateStatements(stmts []StatementInput) error {
	if len(stmts) < 2 || len(stmts) > 6 {
		return fmt.Errorf("%w: true/false questions must have 2–6 statements", apierr.ErrValidation)
	}
	for _, s := range stmts {
		if strings.TrimSpace(s.Text) == "" {
			return fmt.Errorf("%w: statement text cannot be empty", apierr.ErrValidation)
		}
	}
	return nil
}

// EnsureContributorActive is called before create to ensure the contributor is active (not pending).
func EnsureContributorActive(status domain.Status) error {
	if status != domain.StatusActive {
		return errors.New("contributor account is not active")
	}
	return nil
}

var (
	htmlPolicy     *bluemonday.Policy
	uploadURLRegex = regexp.MustCompile(`/uploads/[a-f0-9-]+\.(png|jpe?g|gif|webp)`)
)

func init() {
	htmlPolicy = bluemonday.NewPolicy()
	htmlPolicy.AllowElements("b", "i", "u", "p", "br", "ul", "ol", "li")
	htmlPolicy.AllowAttrs("src").Matching(regexp.MustCompile(`^/uploads/[a-f0-9-]+\.(png|jpe?g|gif|webp)$`)).OnElements("img")
	htmlPolicy.AllowAttrs("data-formula").OnElements("span")
	htmlPolicy.AllowAttrs("class").OnElements("span")
	htmlPolicy.AllowNoAttrs().OnElements("sup", "sub")
}

// SanitizeHTML sanitizes rich text HTML input, allowing only safe formatting,
// formula spans, and uploads images.
func SanitizeHTML(raw string) string {
	return htmlPolicy.Sanitize(raw)
}

func sanitizeOptional(s *string) *string {
	if s == nil {
		return nil
	}
	clean := SanitizeHTML(*s)
	return &clean
}

// sanitize text fields within question data.
func sanitizeQuestion(q *domain.Question) {
	q.Text = htmlPolicy.Sanitize(q.Text)
	if q.Explanation != nil {
		clean := htmlPolicy.Sanitize(*q.Explanation)
		q.Explanation = &clean
	}
	for i := range q.Options {
		q.Options[i].Text = htmlPolicy.Sanitize(q.Options[i].Text)
	}
	for i := range q.Statements {
		q.Statements[i].Text = htmlPolicy.Sanitize(q.Statements[i].Text)
	}
}

// extractUploadURLs finds all /uploads/... paths in an HTML string.
func extractUploadURLs(html string) []string {
	return uploadURLRegex.FindAllString(html, -1)
}

// collectAllText gathers all text fields from a question into a single string
// for URL extraction.
func collectAllText(q *domain.Question) string {
	var b strings.Builder
	b.WriteString(q.Text)
	if q.Explanation != nil {
		b.WriteString(*q.Explanation)
	}
	for _, o := range q.Options {
		b.WriteString(o.Text)
	}
	for _, s := range q.Statements {
		b.WriteString(s.Text)
	}
	return b.String()
}

// cleanupOrphanImages deletes image files that are in the old set but not the new set.
func (s *Service) cleanupOrphanImages(oldHTML, newHTML string) {
	oldURLs := makeSet(extractUploadURLs(oldHTML))
	newURLs := makeSet(extractUploadURLs(newHTML))
	for url := range oldURLs {
		if newURLs[url] {
			continue
		}
		_ = s.store.Delete(context.Background(), url)
	}
}

func (s *Service) ListUsageStats(ctx context.Context, callerID uuid.UUID, limit int) ([]domain.QuestionUsageEntry, error) {
	return s.questions.ListUsageStats(ctx, callerID, limit)
}

func makeSet(urls []string) map[string]bool {
	s := make(map[string]bool, len(urls))
	for _, u := range urls {
		s[u] = true
	}
	return s
}
