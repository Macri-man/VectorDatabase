package main

import (
	"encoding/json"
	"io"
	"math"
)

type Vector struct {
	Name   string    `json:"name"`
	Vector []float64 `json:"vector"`
}

func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func DecodeVector(r io.Reader, v *Vector) error {
	return json.NewDecoder(r).Decode(v)
}

func EncodeVector(w io.Writer, v Vector) error {
	return json.NewEncoder(w).Encode(v)
}

func SearchBestMatch(all []Vector, input Vector) (Vector, float64) {
	var best Vector
	bestScore := -1.0

	for _, v := range all {
		if len(v.Vector) != len(input.Vector) {
			continue
		}
		score := CosineSimilarity(v.Vector, input.Vector)
		if score > bestScore {
			bestScore = score
			best = v
		}
	}
	return best, bestScore
}
