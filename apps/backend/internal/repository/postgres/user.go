package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
	"github.com/yourorg/tkaprep/apps/backend/internal/pkg/apierr"
	"github.com/yourorg/tkaprep/apps/backend/internal/repository"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, name, email, password_hash, role, status, education_level, gender, phone, avatar_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		user.ID, user.Name, user.Email, user.PasswordHash,
		string(user.Role), string(user.Status), user.EducationLevel, user.Gender, user.Phone, user.AvatarURL, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apierr.ErrConflict
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	var role, status string
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, password_hash, role, status, education_level, gender, phone, avatar_url, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &role, &status, &u.EducationLevel, &u.Gender, &u.Phone, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierr.ErrNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	u.Role = domain.Role(role)
	u.Status = domain.Status(status)
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	var role, status string
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, password_hash, role, status, education_level, gender, phone, avatar_url, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &role, &status, &u.EducationLevel, &u.Gender, &u.Phone, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierr.ErrNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	u.Role = domain.Role(role)
	u.Status = domain.Status(status)
	return &u, nil
}

func (r *UserRepository) List(ctx context.Context, f repository.UserAdminFilter) ([]*domain.User, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	where := []string{"1=1"}
	args := []any{}
	n := 1

	if f.Search != "" {
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR email ILIKE $%d)", n, n+1))
		like := "%" + f.Search + "%"
		args = append(args, like, like)
		n += 2
	}
	if f.Role != nil {
		where = append(where, fmt.Sprintf("role = $%d", n))
		args = append(args, string(*f.Role))
		n++
	}
	if f.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", n))
		args = append(args, string(*f.Status))
		n++
	}

	pred := strings.Join(where, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE "+pred, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	args = append(args, f.Limit, offset)
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(`SELECT id, name, email, password_hash, role, status, education_level, gender, phone, avatar_url, created_at, updated_at
		             FROM users WHERE %s
		             ORDER BY created_at DESC
		             LIMIT $%d OFFSET $%d`, pred, n, n+1),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		var role, status string
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &role, &status, &u.EducationLevel, &u.Gender, &u.Phone, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		u.Role = domain.Role(role)
		u.Status = domain.Status(status)
		users = append(users, &u)
	}
	return users, total, rows.Err()
}

func (r *UserRepository) UpdateEducationLevel(ctx context.Context, id uuid.UUID, level *domain.EducationLevel) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET education_level = $1, updated_at = NOW() WHERE id = $2`,
		level, id,
	)
	if err != nil {
		return fmt.Errorf("update education level: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, gender *domain.Gender, phone *string, avatarURL *string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET gender = $1, phone = $2, avatar_url = $3, updated_at = NOW() WHERE id = $4`,
		gender, phone, avatarURL, id,
	)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2`,
		string(status), id,
	)
	if err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.ErrNotFound
	}
	return nil
}
