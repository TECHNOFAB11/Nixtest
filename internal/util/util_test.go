package util

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	apperrors "gitlab.com/technofab/nixtest/internal/errors"
)

func TestComputeDiff(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
		actual   string
		wantDiff string
		wantErr  bool
	}{
		{
			name:     "identical strings",
			expected: "line1\nline2\n",
			actual:   "line1\nline2\n",
			wantDiff: "",
		},
		{
			name:     "simple change",
			expected: "hello\nworld\n",
			actual:   "hello\nuniverse\n",
			wantDiff: `--- expected
+++ actual
@@ -1,2 +1,2 @@
 hello
-world
+universe
`,
		},
		{
			name:     "addition",
			expected: "line1\nline3\n",
			actual:   "line1\nline2\nline3\n",
			wantDiff: `--- expected
+++ actual
@@ -1,2 +1,3 @@
 line1
+line2
 line3
`,
		},
		{
			name:     "deletion",
			expected: "line1\nline2\nline3\n",
			actual:   "line1\nline3\n",
			wantDiff: `--- expected
+++ actual
@@ -1,3 +1,2 @@
 line1
-line2
 line3
`,
		},
		{
			name:     "empty strings",
			expected: "",
			actual:   "",
			wantDiff: "",
		},
		{
			name:     "expected empty, actual has content",
			expected: "",
			actual:   "new content\n",
			wantDiff: `--- expected
+++ actual
@@ -0,0 +1 @@
+new content
`,
		},
		{
			name:     "expected has content, actual empty",
			expected: "old content\n",
			actual:   "",
			wantDiff: `--- expected
+++ actual
@@ -1 +0,0 @@
-old content
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotDiff, err := ComputeDiff(tc.expected, tc.actual)

			if (err != nil) != tc.wantErr {
				t.Errorf("ComputeDiff() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if err != nil {
				return
			}

			normalizedGotDiff := strings.ReplaceAll(gotDiff, "\r\n", "\n")
			normalizedWantDiff := strings.ReplaceAll(tc.wantDiff, "\r\n", "\n")

			if normalizedGotDiff != normalizedWantDiff {
				t.Errorf("ComputeDiff() mismatch:\n--- GOT DIFF ---\n%s\n--- WANT DIFF ---\n%s", normalizedGotDiff, normalizedWantDiff)
				metaDiff, _ := ComputeDiff(normalizedWantDiff, normalizedGotDiff)
				if metaDiff != "" {
					t.Errorf("--- DIFF OF DIFFS ---\n%s", metaDiff)
				}
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	type sampleStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	validJSON := `{"name": "test", "value": 123}`
	invalidJSON := `{"name": "test", value: 123}`

	tempDir := t.TempDir()

	validFilePath := filepath.Join(tempDir, "valid.json")
	if err := os.WriteFile(validFilePath, []byte(validJSON), 0644); err != nil {
		t.Fatalf("Failed to write valid temp file: %v", err)
	}

	invalidJSONFilePath := filepath.Join(tempDir, "invalid_content.json")
	if err := os.WriteFile(invalidJSONFilePath, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("Failed to write invalid temp file: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		want        sampleStruct
		wantErr     bool
		wantErrType any // expected error type, e.g., (*apperrors.FileReadError)(nil)
		errContains string
	}{
		{
			name:     "Valid JSON file",
			filePath: validFilePath,
			want:     sampleStruct{Name: "test", Value: 123},
			wantErr:  false,
		},
		{
			name:        "File not found",
			filePath:    filepath.Join(tempDir, "nonexistent.json"),
			want:        sampleStruct{},
			wantErr:     true,
			wantErrType: (*apperrors.FileReadError)(nil),
			errContains: "failed to open",
		},
		{
			name:        "Invalid JSON content",
			filePath:    invalidJSONFilePath,
			want:        sampleStruct{},
			wantErr:     true,
			wantErrType: (*apperrors.JSONUnmarshalError)(nil),
			errContains: "failed to decode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFile[sampleStruct](tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if tt.wantErrType != nil && !errors.As(err, &tt.wantErrType) {
					t.Errorf("ParseFile() error type = %T, want type %T", err, tt.wantErrType)
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseFile() error = %v, want error containing %v", err, tt.errContains)
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrefixLines(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		prefix string
		want   string
	}{
		{"Empty input", "", "| ", "| "},
		{"Single line", "hello", "| ", "| hello"},
		{"Multiple lines", "hello\nworld", "> ", "> hello\n> world"},
		{"Line with trailing newline", "hello\n", "- ", "- hello\n- "},
		{"Prefix with space", "line", "PREFIX ", "PREFIX line"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrefixLines(tt.input, tt.prefix); got != tt.want {
				t.Errorf("PrefixLines() = %q, want %q", got, tt.want)
			}
		})
	}
}
func TestIsString(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"String value", "hello", true},
		{"Empty string", "", true},
		{"Integer value", 123, false},
		{"Boolean value", true, false},
		{"Nil value", nil, false},
		{"Struct value", struct{}{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsString(tt.value); got != tt.want {
				t.Errorf("IsString() = %v, want %v for value %v", got, tt.want, tt.value)
			}
		})
	}
}
