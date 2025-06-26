package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"sync"
)

type Vector struct {
	Name   string    `json:"name"`
	Vector []float64 `json:"vector"`
}

var (
	dataFile = "data.jsonl"
	vectors  []Vector
	mu       sync.Mutex
)

func cosineSimilarity(a, b []float64) float64 {
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

func permuteToward(original, bias []float64, alpha float64) []float64 {
	out := make([]float64, len(original))
	for i := range original {
		out[i] = (1-alpha)*original[i] + alpha*bias[i]
	}
	return out
}

func permuteVector(original []float64, noiseLevel float64, minSimilarity float64, maxTries int) []float64 {
	for try := 0; try < maxTries; try++ {
		noisy := make([]float64, len(original))
		for i, val := range original {
			noise := (rand.Float64()*2 - 1) * noiseLevel // noise in [-noiseLevel, +noiseLevel]
			noisy[i] = val + noise
		}

		if cosineSimilarity(original, noisy) >= minSimilarity {
			return noisy
		}
	}
	return original // fallback: return original if no good permutation found
}

func loadVectors() {
	file, err := os.Open(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return // First time, no file yet
		}
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var v Vector
		if err := json.Unmarshal(scanner.Bytes(), &v); err == nil {
			vectors = append(vectors, v)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Println("Failed reading file:", err)
	}
}

func appendVectorToFile(v Vector) {
	f, err := os.OpenFile(dataFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening file for append:", err)
		return
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(v); err != nil {
		log.Println("Error writing vector to file:", err)
	}
}

func addVector(w http.ResponseWriter, r *http.Request) {
	var v Vector
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Overwrite by name
	for i, existing := range vectors {
		if existing.Name == v.Name {
			vectors[i] = v
			saveAllVectorsToFile()
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	vectors = append(vectors, v)
	appendVectorToFile(v)
	w.WriteHeader(http.StatusOK)
}

func saveAllVectorsToFile() {
	f, err := os.Create(dataFile)
	if err != nil {
		log.Println("Error overwriting data file:", err)
		return
	}
	defer f.Close()

	for _, v := range vectors {
		data, _ := json.Marshal(v)
		f.Write(append(data, '\n'))
	}
}

func searchVector(w http.ResponseWriter, r *http.Request) {
	var input Vector
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	var best Vector
	bestScore := -1.0
	for _, v := range vectors {
		score := cosineSimilarity(input.Vector, v.Vector)
		if score > bestScore {
			best = v
			bestScore = score
		}
	}
	json.NewEncoder(w).Encode(best)
}

func main() {
	loadVectors()

	http.HandleFunc("/add", addVector)
	http.HandleFunc("/search", searchVector)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
