package session

import "math"

const irtMinSamples = 5 // minimum sample_size per question before theta is estimated

// irtQuestion holds per-question data needed to estimate student ability.
type irtQuestion struct {
	difficulty float64
	correct    bool
	sampleSize int
}

// estimateDifficulty returns the Rasch difficulty parameter b for a question
// given its historical correct_count and sample_size.
// b = logit(failure_rate) = log((1-p)/p) where p = correct_count/sample_size.
// Clamped to avoid log(0): p is clipped to [0.001, 0.999].
func estimateDifficulty(correctCount, sampleSize int) float64 {
	if sampleSize == 0 {
		return 0.0
	}
	p := math.Max(0.001, math.Min(0.999, float64(correctCount)/float64(sampleSize)))
	return math.Log((1 - p) / p)
}

// computeIRTTheta estimates student ability theta via Rasch 1PL MLE (Newton-Raphson).
// Returns nil when:
//   - qs is empty
//   - any question has sampleSize < irtMinSamples (cold-start guard)
func computeIRTTheta(qs []irtQuestion) *float64 {
	if len(qs) == 0 {
		return nil
	}
	for _, q := range qs {
		if q.sampleSize < irtMinSamples {
			return nil
		}
	}

	// Starting estimate: logit of proportion correct (clipped to avoid ±Inf).
	nCorrect := 0
	for _, q := range qs {
		if q.correct {
			nCorrect++
		}
	}
	p0 := math.Max(0.01, math.Min(0.99, float64(nCorrect)/float64(len(qs))))
	theta := math.Log(p0 / (1 - p0))

	// Newton-Raphson: max 30 iterations, stop when gradient is tiny.
	for iter := 0; iter < 30; iter++ {
		grad, hess := 0.0, 0.0
		for _, q := range qs {
			pi := 1.0 / (1.0 + math.Exp(q.difficulty-theta))
			x := 0.0
			if q.correct {
				x = 1.0
			}
			grad += x - pi
			hess -= pi * (1 - pi)
		}
		if math.Abs(hess) < 1e-9 {
			break
		}
		step := grad / hess
		theta -= step
		if math.Abs(step) < 1e-4 {
			break
		}
	}

	// Clamp to a reasonable range to avoid extreme estimates with sparse data.
	theta = math.Max(-6.0, math.Min(6.0, theta))
	v := math.Round(theta*10000) / 10000 // 4 decimal places, matches DECIMAL(6,4)
	return &v
}
