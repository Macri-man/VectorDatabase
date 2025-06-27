package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var (
	vectors  []Vector
	dataFile = "data.json"
	mu       sync.Mutex
)

func addVector(w http.ResponseWriter, r *http.Request) {
	var newVec Vector
	if err := DecodeVector(r.Body, &newVec); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	existingVectors := LoadVectors(dataFile)

	updated := false
	for i, existing := range existingVectors {
		if existing.Name == newVec.Name {
			existingVectors[i] = newVec
			updated = true
			break
		}
	}
	if !updated {
		existingVectors = append(existingVectors, newVec)
	}

	SaveAllVectorsToFile(dataFile, existingVectors)
	w.WriteHeader(http.StatusOK)
}

func searchVector(w http.ResponseWriter, r *http.Request) {
	var input Vector
	if err := DecodeVector(r.Body, &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	best, _ := SearchBestMatch(vectors, input)
	EncodeVector(w, best)
}

func listAllVectors(w http.ResponseWriter, r *http.Request) {
	vectors := LoadVectors(dataFile)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(vectors); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	staticDir := filepath.Join(dir, "static")

	vectors = LoadVectors(dataFile)

	http.HandleFunc("/add", addVector)
	http.HandleFunc("/search", searchVector)
	http.HandleFunc("/all", listAllVectors)
	http.Handle("/", http.FileServer(http.Dir(staticDir)))

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
