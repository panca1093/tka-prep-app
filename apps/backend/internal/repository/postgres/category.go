package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
)

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.name, c.description, c.created_by, COUNT(t.id)::int AS test_count, c.created_at
		FROM categories c
		LEFT JOIN tests t ON t.category_id = c.id
		GROUP BY c.id, c.name, c.description, c.created_by, c.created_at
		ORDER BY c.name`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var out []*domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedBy, &c.TestCount, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		out = append(out, &c)
	}
	return out, rows.Err()
}

func (r *CategoryRepository) Create(ctx context.Context, name string) (*domain.Category, error) {
	var c domain.Category
	err := r.pool.QueryRow(ctx, `
		INSERT INTO categories (name) VALUES ($1)
		RETURNING id, name, description, created_by, 0, created_at`, name,
	).Scan(&c.ID, &c.Name, &c.Description, &c.CreatedBy, &c.TestCount, &c.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apierr.ErrConflict
		}
		return nil, fmt.Errorf("create category: %w", err)
	}
	return &c, nil
}

func (r *CategoryRepository) Update(ctx context.Context, id uuid.UUID, name string) (*domain.Category, error) {
	var c domain.Category
	err := r.pool.QueryRow(ctx, `
		UPDATE categories SET name = $2 WHERE id = $1
		RETURNING id, name, description, created_by,
		          (SELECT COUNT(*) FROM tests WHERE category_id = $1)::int, created_at`,
		id, name,
	).Scan(&c.ID, &c.Name, &c.Description, &c.CreatedBy, &c.TestCount, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierr.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apierr.ErrConflict
		}
		return nil, fmt.Errorf("update category: %w", err)
	}
	return &c, nil
}

func (r *CategoryRepository) UpdateOwned(ctx context.Context, callerID, id uuid.UUID, name string, description *string) (*domain.Category, error) {
	var c domain.Category
	err := r.pool.QueryRow(ctx, `
		UPDATE categories SET name = $3, description = $4
		WHERE id = $2 AND created_by = $1
		RETURNING id, name, description, created_by,
		          (SELECT COUNT(*) FROM tests WHERE category_id = $2)::int, created_at`,
		callerID, id, name, description,
	).Scan(&c.ID, &c.Name, &c.Description, &c.CreatedBy, &c.TestCount, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Could be: category doesn't exist, or it exists but is not owned by callerID.
			// Distinguish by checking existence.
			var exists bool
			_ = r.pool.QueryRow(ctx,
				`SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)`, id,
			).Scan(&exists)
			if !exists {
				return nil, apierr.ErrNotFound
			}
			return nil, apierr.ErrForbidden
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apierr.ErrConflict
		}
		return nil, fmt.Errorf("update owned category: %w", err)
	}
	return &c, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return fmt.Errorf("category is in use: %w", apierr.ErrConflict)
		}
		return fmt.Errorf("delete category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}
	return nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	var c domain.Category
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, created_by,
		       (SELECT COUNT(*) FROM tests WHERE category_id = $1)::int, created_at
		FROM categories WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.Description, &c.CreatedBy, &c.TestCount, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierr.ErrNotFound
		}
		return nil, fmt.Errorf("find category: %w", err)
	}
	return &c, nil
}

func (r *CategoryRepository) TestCount(ctx context.Context, id uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tests WHERE category_id = $1`, id).Scan(&n)
	return n, err
}
