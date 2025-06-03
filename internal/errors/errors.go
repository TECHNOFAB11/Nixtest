package errors

import (
	"fmt"
)

// NixBuildError indicates an error during `nix build`
type NixBuildError struct {
	Derivation string
	Stderr     string
	Err        error // underlying error from exec.Cmd or similar
}

func (e *NixBuildError) Error() string {
	return fmt.Sprintf("nix build for %s failed: %v (stderr: %s)", e.Derivation, e.Err, e.Stderr)
}
func (e *NixBuildError) Unwrap() error { return e.Err }

// NixNoOutputPathError indicates `nix build` succeeded but produced no output path
type NixNoOutputPathError struct {
	Derivation string
	Stderr     string
}

func (e *NixNoOutputPathError) Error() string {
	return fmt.Sprintf("nix build for %s produced no output path (stderr: %s)", e.Derivation, e.Stderr)
}

// FileReadError indicates an error reading a file, often a derivation output
type FileReadError struct {
	Path string
	Err  error
}

func (e *FileReadError) Error() string {
	return fmt.Sprintf("failed to read file %s: %v", e.Path, e.Err)
}
func (e *FileReadError) Unwrap() error { return e.Err }

// JSONUnmarshalError indicates an error unmarshalling JSON data
type JSONUnmarshalError struct {
	Source string // e.g. file path or "derivation output"
	Err    error
}

func (e *JSONUnmarshalError) Error() string {
	return fmt.Sprintf("failed to unmarshal JSON from %s: %v", e.Source, e.Err)
}
func (e *JSONUnmarshalError) Unwrap() error { return e.Err }

// ScriptExecutionError indicates an error starting or waiting for a script
type ScriptExecutionError struct {
	Path string // path to script that was attempted to run
	Err  error
}

func (e *ScriptExecutionError) Error() string {
	return fmt.Sprintf("script %s execution failed: %v", e.Path, e.Err)
}
func (e *ScriptExecutionError) Unwrap() error { return e.Err }

// SnapshotCreateError indicates an error during snapshot creation
type SnapshotCreateError struct {
	FilePath string
	Err      error
}

func (e *SnapshotCreateError) Error() string {
	return fmt.Sprintf("failed to create/update snapshot %s: %v", e.FilePath, e.Err)
}
func (e *SnapshotCreateError) Unwrap() error { return e.Err }

// SnapshotLoadError indicates an error loading a snapshot file
type SnapshotLoadError struct {
	FilePath string
	Err      error
}

func (e *SnapshotLoadError) Error() string {
	return fmt.Sprintf("failed to load/parse snapshot %s: %v", e.FilePath, e.Err)
}
func (e *SnapshotLoadError) Unwrap() error { return e.Err }
