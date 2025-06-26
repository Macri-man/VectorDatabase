package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
)

func LoadVectors(filename string) []Vector {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		log.Fatal(err)
	}
	defer file.Close()

	var vectors []Vector
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
	return vectors
}

func AppendVectorToFile(filename string, v Vector) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(v); err != nil {
		log.Println("Error writing vector:", err)
	}
}

func SaveAllVectorsToFile(filename string, vectors []Vector) {
	f, err := os.Create(filename)
	if err != nil {
		log.Println("Error overwriting file:", err)
		return
	}
	defer f.Close()

	for _, v := range vectors {
		data, _ := json.Marshal(v)
		f.Write(append(data, '\n'))
	}
}
