package session

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// ─── TestEstimateDifficulty ───────────────────────────────────────────────────

func TestEstimateDifficulty(t *testing.T) {
	tests := []struct {
		name         string
		correctCount int
		sampleSize   int
		wantApprox   float64
		tolerance    float64
	}{
		// p=0.70 → b = logit(0.30) = log(0.30/0.70) ≈ -0.847
		{"I-01: easy question (70% correct)", 70, 100, -0.847, 0.01},
		// p=0.30 → b = logit(0.70) = log(0.70/0.30) ≈ +0.847
		{"I-02: hard question (30% correct)", 30, 100, 0.847, 0.01},
		// p=0.50 → b = 0
		{"I-03: neutral question (50% correct)", 50, 100, 0.0, 0.01},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := estimateDifficulty(tc.correctCount, tc.sampleSize)
			require.InDelta(t, tc.wantApprox, got, tc.tolerance)
		})
	}
}

// ─── TestComputeIRTTheta ─────────────────────────────────────────────────────

func TestComputeIRTTheta(t *testing.T) {
	t.Run("I-04: returns nil when any question has sample_size < 5", func(t *testing.T) {
		qs := []irtQuestion{
			{difficulty: 0.0, correct: true, sampleSize: 10},
			{difficulty: 0.5, correct: false, sampleSize: 4}, // below threshold
		}
		got := computeIRTTheta(qs)
		require.Nil(t, got)
	})

	t.Run("I-05: returns nil for empty question list", func(t *testing.T) {
		got := computeIRTTheta(nil)
		require.Nil(t, got)
	})

	t.Run("I-06: same proportion correct, harder questions → higher theta estimate", func(t *testing.T) {
		// Same 2/3 correct — but set A has easy questions (b=-1.0), set B has hard questions (b=+1.0).
		// Getting 2/3 right on hard questions implies higher ability.
		setA := []irtQuestion{ // easy: b=-1.0; gets 2 right, 1 wrong
			{difficulty: -1.0, correct: true, sampleSize: 20},
			{difficulty: -1.0, correct: true, sampleSize: 20},
			{difficulty: -1.0, correct: false, sampleSize: 20},
		}
		setB := []irtQuestion{ // hard: b=+1.0; gets 2 right, 1 wrong
			{difficulty: 1.0, correct: true, sampleSize: 20},
			{difficulty: 1.0, correct: true, sampleSize: 20},
			{difficulty: 1.0, correct: false, sampleSize: 20},
		}
		thetaA := computeIRTTheta(setA)
		thetaB := computeIRTTheta(setB)
		require.NotNil(t, thetaA)
		require.NotNil(t, thetaB)
		// Answering hard questions correctly → higher ability estimate.
		require.Greater(t, *thetaB, *thetaA)
	})

	t.Run("I-07: student with all correct has higher theta than student with all wrong", func(t *testing.T) {
		difficulties := []float64{0.0, 0.0, 0.0}
		allCorrect := make([]irtQuestion, len(difficulties))
		allWrong := make([]irtQuestion, len(difficulties))
		for i, d := range difficulties {
			allCorrect[i] = irtQuestion{difficulty: d, correct: true, sampleSize: 10}
			allWrong[i] = irtQuestion{difficulty: d, correct: false, sampleSize: 10}
		}
		thetaHigh := computeIRTTheta(allCorrect)
		thetaLow := computeIRTTheta(allWrong)
		require.NotNil(t, thetaHigh)
		require.NotNil(t, thetaLow)
		require.Greater(t, *thetaHigh, *thetaLow)
	})

	t.Run("I-08: result is finite (not NaN or Inf)", func(t *testing.T) {
		qs := []irtQuestion{
			{difficulty: 0.5, correct: true, sampleSize: 10},
			{difficulty: -0.5, correct: false, sampleSize: 10},
		}
		got := computeIRTTheta(qs)
		require.NotNil(t, got)
		require.False(t, math.IsNaN(*got))
		require.False(t, math.IsInf(*got, 0))
	})
}
