package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
)

type TopicRepository struct {
	pool *pgxpool.Pool
}

func NewTopicRepository(pool *pgxpool.Pool) *TopicRepository {
	return &TopicRepository{pool: pool}
}

func (r *TopicRepository) List(ctx context.Context) ([]*domain.Topic, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, description, created_by, created_at FROM topics ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list topics: %w", err)
	}
	defer rows.Close()

	var topics []*domain.Topic
	for rows.Next() {
		t := &domain.Topic{}
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan topic: %w", err)
		}
		topics = append(topics, t)
	}
	return topics, rows.Err()
}

func (r *TopicRepository) Create(ctx context.Context, t *domain.Topic) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO topics (id, name, description, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		t.ID, t.Name, t.Description, t.CreatedBy, t.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apierr.ErrConflict
		}
		return fmt.Errorf("insert topic: %w", err)
	}
	return nil
}

func (r *TopicRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error) {
	t := &domain.Topic{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, created_by, created_at FROM topics WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.Description, &t.CreatedBy, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierr.ErrNotFound
		}
		return nil, fmt.Errorf("find topic by id: %w", err)
	}
	return t, nil
}

func (r *TopicRepository) Update(ctx context.Context, t *domain.Topic) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE topics SET name = $1, description = $2 WHERE id = $3`,
		t.Name, t.Description, t.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apierr.ErrConflict
		}
		return fmt.Errorf("update topic: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}
	return nil
}

func (r *TopicRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM topics WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete topic: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}
	return nil
}

func (r *TopicRepository) IsUsedByQuestions(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM questions WHERE topic_id = $1 LIMIT 1`, id,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check topic usage: %w", err)
	}
	return count > 0, nil
}
