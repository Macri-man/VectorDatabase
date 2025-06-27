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
		log.Println("Decode error:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Received vector:", newVec)

	mu.Lock()
	defer mu.Unlock()

	existingVectors := LoadVectors(dataFile)
	log.Printf("Loaded %d vectors from file", len(existingVectors))

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
		log.Println("Appended new vector")
	} else {
		log.Println("Updated existing vector")
	}

	SaveAllVectorsToFile(dataFile, existingVectors)
	log.Println("Saved vectors to file")

	w.WriteHeader(http.StatusOK)
}

func searchVector(w http.ResponseWriter, r *http.Request) {
	var input Vector
	if err := DecodeVector(r.Body, &input); err != nil {
		log.Println("Failed to decode input vector:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(input.Vector) == 0 {
		http.Error(w, "Search vector is empty", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	vectors := LoadVectors(dataFile)
	if len(vectors) == 0 {
		http.Error(w, "No vectors available for search", http.StatusNotFound)
		return
	}

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
