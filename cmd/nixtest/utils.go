package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func ParseFile[T any](filePath string) (result T, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return result, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&result)
	if err != nil {
		return result, fmt.Errorf("failed to decode JSON from file %s: %w", filePath, err)
	}

	return result, nil
}
