package main

import (
	"math/rand/v2"
)

func PermuteVector(original []float64, noiseLevel float64, minSimilarity float64, maxTries int) []float64 {
	for try := 0; try < maxTries; try++ {
		noisy := make([]float64, len(original))
		for i, val := range original {
			noise := (rand.Float64()*2 - 1) * noiseLevel
			noisy[i] = val + noise
		}
		if CosineSimilarity(original, noisy) >= minSimilarity {
			return noisy
		}
	}
	return original
}

func PermuteToward(original, bias []float64, alpha float64) []float64 {
	out := make([]float64, len(original))
	for i := range original {
		out[i] = (1-alpha)*original[i] + alpha*bias[i]
	}
	return out
}
