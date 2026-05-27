package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// irtDB is a minimal interface satisfied by both *pgxpool.Pool and pgx.Tx,
// so IRT operations can run inside or outside a transaction.
type irtDB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// IRTParams holds the IRT difficulty parameters for a question.
type IRTParams struct {
	QuestionID   uuid.UUID
	CorrectCount int
	SampleSize   int
	DifficultyB  float64
	UpdatedAt    time.Time
}

// UpsertIRTParam increments sample_size (and correct_count if wasCorrect) for a
// question, then recomputes difficulty_b from the new totals.
// Pass the transaction so it participates in the submit transaction and so that
// the subsequent GetIRTParams call sees the updated values.
func UpsertIRTParam(ctx context.Context, db irtDB, questionID uuid.UUID, wasCorrect bool) error {
	correctDelta := 0
	if wasCorrect {
		correctDelta = 1
	}

	// Step 1: increment counters.
	if _, err := db.Exec(ctx, `
		INSERT INTO question_irt_params (question_id, correct_count, sample_size, difficulty_b, updated_at)
		VALUES ($1, $2, 1, 0.0, now())
		ON CONFLICT (question_id) DO UPDATE
		SET correct_count = question_irt_params.correct_count + $2,
		    sample_size   = question_irt_params.sample_size + 1,
		    updated_at    = now()
	`, questionID, correctDelta); err != nil {
		return err
	}

	// Step 2: recompute difficulty_b from the updated totals using Postgres ln().
	_, err := db.Exec(ctx, `
		UPDATE question_irt_params
		SET difficulty_b = CASE
			WHEN sample_size = 0 THEN 0
			ELSE ln(
				GREATEST(0.001, 1 - LEAST(0.999, correct_count::DECIMAL / sample_size))
				/ GREATEST(0.001, LEAST(0.999, correct_count::DECIMAL / sample_size))
			)
		END
		WHERE question_id = $1
	`, questionID)
	return err
}

// GetIRTParams fetches IRT params for a slice of question IDs using the given db
// (pool or tx). Questions with no params yet are omitted from the result map.
func GetIRTParams(ctx context.Context, db irtDB, questionIDs []uuid.UUID) (map[uuid.UUID]IRTParams, error) {
	rows, err := db.Query(ctx, `
		SELECT question_id, correct_count, sample_size, difficulty_b, updated_at
		FROM question_irt_params
		WHERE question_id = ANY($1)
	`, questionIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]IRTParams, len(questionIDs))
	for rows.Next() {
		var p IRTParams
		if err := rows.Scan(&p.QuestionID, &p.CorrectCount, &p.SampleSize, &p.DifficultyB, &p.UpdatedAt); err != nil {
			return nil, err
		}
		result[p.QuestionID] = p
	}
	return result, rows.Err()
}
