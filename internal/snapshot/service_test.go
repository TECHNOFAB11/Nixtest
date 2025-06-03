package snapshot

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	apperrors "gitlab.com/technofab/nixtest/internal/errors"
)

func TestDefaultService_GetPath(t *testing.T) {
	service := NewDefaultService()
	tests := []struct {
		name        string
		snapshotDir string
		testName    string
		want        string
	}{
		{"Simple name", "/tmp/snapshots", "TestSimple", "/tmp/snapshots/testsimple.snap.json"},
		{"Name with spaces", "snaps", "Test With Spaces", "snaps/test_with_spaces.snap.json"},
		{"Name with mixed case", "./data", "MyTestSERVICE", "data/mytestservice.snap.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := filepath.ToSlash(tt.want) // Normalize for comparison
			got := service.GetPath(tt.snapshotDir, tt.testName)
			if got != want {
				t.Errorf("GetPath() = %v, want %v", got, want)
			}
		})
	}
}

func TestDefaultService_CreateFileAndLoadFile(t *testing.T) {
	service := NewDefaultService()
	tempDir := t.TempDir()
	filePath := service.GetPath(tempDir, "Test Snapshot Content")

	dataToWrite := map[string]any{
		"name":   "test snapshot",
		"value":  float64(42),
		"nested": map[string]any{"active": true},
	}

	t.Run("CreateFile", func(t *testing.T) {
		err := service.CreateFile(filePath, dataToWrite)
		if err != nil {
			t.Fatalf("CreateFile() failed: %v", err)
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read created snapshot file: %v", err)
		}

		var readData map[string]any
		if err := json.Unmarshal(content, &readData); err != nil {
			t.Fatalf("Failed to unmarshal created snapshot content: %v", err)
		}
		if !reflect.DeepEqual(readData, dataToWrite) {
			t.Errorf("CreateFile() content mismatch. Got %v, want %v", readData, dataToWrite)
		}
	})

	t.Run("LoadFile", func(t *testing.T) {
		// ensure file exists
		if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
			if errCreate := service.CreateFile(filePath, dataToWrite); errCreate != nil {
				t.Fatalf("Prerequisite CreateFile for LoadFile test failed: %v", errCreate)
			}
		}

		loadedData, err := service.LoadFile(filePath)
		if err != nil {
			t.Fatalf("LoadFile() failed: %v", err)
		}
		loadedMap, ok := loadedData.(map[string]any)
		if !ok {
			t.Fatalf("LoadFile() did not return a map, got %T", loadedData)
		}
		if !reflect.DeepEqual(loadedMap, dataToWrite) {
			t.Errorf("LoadFile() content mismatch. Got %v, want %v", loadedMap, dataToWrite)
		}
	})

	t.Run("LoadFile_NotExist", func(t *testing.T) {
		nonExistentPath := service.GetPath(tempDir, "NonExistentSnapshot")
		_, err := service.LoadFile(nonExistentPath)
		if err == nil {
			t.Error("LoadFile() expected error for non-existent file, got nil")
		} else {
			var loadErr *apperrors.SnapshotLoadError
			if !errors.As(err, &loadErr) {
				t.Errorf("LoadFile() wrong error type, got %T, want *apperrors.SnapshotLoadError", err)
			}
			var fileReadErr *apperrors.FileReadError
			if !errors.As(loadErr.Err, &fileReadErr) {
				t.Errorf("SnapshotLoadError wrong wrapped error type, got %T, want *apperrors.FileReadError", loadErr.Err)
			}
			if !strings.Contains(err.Error(), "failed to open") {
				t.Errorf("LoadFile() error = %v, want error containing 'failed to open'", err)
			}
		}
	})

	t.Run("CreateFile_MarshalError", func(t *testing.T) {
		err := service.CreateFile(filepath.Join(tempDir, "marshal_error.snap.json"), make(chan int))
		if err == nil {
			t.Fatal("CreateFile expected error for unmarshalable data, got nil")
		}
		var createErr *apperrors.SnapshotCreateError
		if !errors.As(err, &createErr) {
			t.Fatalf("Wrong error type for marshal error, got %T", err)
		}
		var marshalErr *apperrors.JSONUnmarshalError
		if !errors.As(createErr.Err, &marshalErr) {
			t.Errorf("SnapshotCreateError did not wrap JSONUnmarshalError for marshal failure, got %T", createErr.Err)
		}
	})
}

func TestDefaultService_Stat(t *testing.T) {
	service := NewDefaultService()
	tempDir := t.TempDir()

	t.Run("File exists", func(t *testing.T) {
		p := filepath.Join(tempDir, "exists.txt")
		if err := os.WriteFile(p, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
		fi, err := service.Stat(p)
		if err != nil {
			t.Errorf("Stat() for existing file failed: %v", err)
		}
		if fi == nil || fi.Name() != "exists.txt" {
			t.Errorf("Stat() returned incorrect FileInfo")
		}
	})

	t.Run("File does not exist", func(t *testing.T) {
		p := filepath.Join(tempDir, "notexists.txt")
		_, err := service.Stat(p)
		if !os.IsNotExist(err) {
			t.Errorf("Stat() for non-existing file: got %v, want os.ErrNotExist", err)
		}
	})
}
