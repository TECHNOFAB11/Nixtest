package snapshot

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"

	apperrors "gitlab.com/technofab/nixtest/internal/errors"
	"gitlab.com/technofab/nixtest/internal/util"
)

// Service defines operations related to test snapshots
type Service interface {
	GetPath(snapshotDir string, testName string) string
	CreateFile(filePath string, data any) error
	LoadFile(filePath string) (any, error)
	Stat(name string) (os.FileInfo, error)
}

type DefaultService struct{}

func NewDefaultService() *DefaultService {
	return &DefaultService{}
}

// GetPath generates the canonical path for a snapshot file
func (s *DefaultService) GetPath(snapshotDir string, testName string) string {
	fileName := filepath.ToSlash(
		strings.ToLower(strings.ReplaceAll(testName, " ", "_")) + ".snap.json",
	)
	return path.Join(snapshotDir, fileName)
}

// CreateFile creates or updates a snapshot file with the given data
func (s *DefaultService) CreateFile(filePath string, data any) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return &apperrors.SnapshotCreateError{
			FilePath: filePath,
			Err:      &apperrors.JSONUnmarshalError{Source: "snapshot data for " + filePath, Err: err},
		}
	}

	err = os.MkdirAll(path.Dir(filePath), 0777)
	if err != nil {
		return &apperrors.SnapshotCreateError{FilePath: filePath, Err: err}
	}

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return &apperrors.SnapshotCreateError{FilePath: filePath, Err: err}
	}
	return nil
}

// LoadFile loads a snapshot file.
func (s *DefaultService) LoadFile(filePath string) (any, error) {
	result, err := util.ParseFile[any](filePath)
	if err != nil {
		return nil, &apperrors.SnapshotLoadError{FilePath: filePath, Err: err}
	}
	return result, nil
}

// Stat just wraps os.Stat
func (s *DefaultService) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
