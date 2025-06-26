package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	vectors  []Vector
	dataFile = "data.json"
	mu       sync.Mutex
)

func addVector(w http.ResponseWriter, r *http.Request) {
	var v Vector
	if err := DecodeVector(r.Body, &v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Overwrite if name exists
	for i, existing := range vectors {
		if existing.Name == v.Name {
			vectors[i] = v
			SaveAllVectorsToFile(dataFile, vectors)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	vectors = append(vectors, v)
	AppendVectorToFile(dataFile, v)
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

func main() {
	vectors = LoadVectors(dataFile)

	http.HandleFunc("/add", addVector)
	http.HandleFunc("/search", searchVector)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
