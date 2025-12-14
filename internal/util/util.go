package util

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/akedrou/textdiff"
	"github.com/akedrou/textdiff/myers"
	apperrors "gitlab.com/TECHNOFAB/nixtest/internal/errors"
)

func ComputeDiff(expected, actual string) (string, error) {
	// FIXME: ComputeEdits deprecated
	edits := myers.ComputeEdits(expected, actual)
	diff, err := textdiff.ToUnified("expected", "actual", expected, edits, 3)
	if err != nil {
		return "", err
	}
	// remove newline hint
	diff = strings.ReplaceAll(diff, "\\ No newline at end of file\n", "")
	return diff, nil
}

// ParseFile reads and decodes a JSON file into the provided type
func ParseFile[T any](filePath string) (result T, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return result, &apperrors.FileReadError{Path: filePath, Err: fmt.Errorf("failed to open: %w", err)}
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&result)
	if err != nil {
		return result, &apperrors.JSONUnmarshalError{Source: filePath, Err: fmt.Errorf("failed to decode: %w", err)}
	}
	return result, nil
}

// PrefixLines adds a prefix to each line of the input string
func PrefixLines(input string, prefix string) string {
	lines := strings.Split(input, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

func IsString(value any) bool {
	_, ok := value.(string)
	return ok
}
