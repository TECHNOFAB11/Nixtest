package errors

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestNixBuildError(t *testing.T) {
	underlyingErr := errors.New("exec: \"nix\": executable file not found in $PATH")
	err := &NixBuildError{
		Derivation: "test.drv",
		Stderr:     "some stderr output",
		Err:        underlyingErr,
	}

	expectedMsg := "nix build for test.drv failed: exec: \"nix\": executable file not found in $PATH (stderr: some stderr output)"
	if err.Error() != expectedMsg {
		t.Errorf("Error() got %q, want %q", err.Error(), expectedMsg)
	}

	if !errors.Is(err, underlyingErr) {
		t.Errorf("Unwrap() failed, underlying error not found")
	}
}

func TestNixNoOutputPathError(t *testing.T) {
	err := &NixNoOutputPathError{
		Derivation: "empty.drv",
		Stderr:     "build successful, but no paths",
	}
	expectedMsg := "nix build for empty.drv produced no output path (stderr: build successful, but no paths)"
	if err.Error() != expectedMsg {
		t.Errorf("Error() got %q, want %q", err.Error(), expectedMsg)
	}
}

func TestFileReadError(t *testing.T) {
	underlyingErr := os.ErrPermission
	err := &FileReadError{
		Path: "/tmp/file.json",
		Err:  underlyingErr,
	}
	expectedMsg := "failed to read file /tmp/file.json: permission denied"
	if err.Error() != expectedMsg {
		t.Errorf("Error() got %q, want %q", err.Error(), expectedMsg)
	}
	if !errors.Is(err, underlyingErr) {
		t.Errorf("Unwrap() failed, underlying error not found")
	}
}

func TestJSONUnmarshalError(t *testing.T) {
	underlyingErr := errors.New("unexpected end of JSON input")
	err := &JSONUnmarshalError{
		Source: "/tmp/data.json",
		Err:    underlyingErr,
	}
	expectedMsg := "failed to unmarshal JSON from /tmp/data.json: unexpected end of JSON input"
	if err.Error() != expectedMsg {
		t.Errorf("Error() got %q, want %q", err.Error(), expectedMsg)
	}
	if !errors.Is(err, underlyingErr) {
		t.Errorf("Unwrap() failed, underlying error not found")
	}
}

func TestScriptExecutionError(t *testing.T) {
	underlyingErr := errors.New("command timed out")
	err := &ScriptExecutionError{
		Path: "/tmp/script.sh",
		Err:  underlyingErr,
	}
	expectedMsg := "script /tmp/script.sh execution failed: command timed out"
	if err.Error() != expectedMsg {
		t.Errorf("Error() got %q, want %q", err.Error(), expectedMsg)
	}
	if !errors.Is(err, underlyingErr) {
		t.Errorf("Unwrap() failed, underlying error not found")
	}
}

func TestSnapshotCreateError(t *testing.T) {
	underlyingErr := os.ErrExist
	err := &SnapshotCreateError{
		FilePath: "/snapshots/test.snap.json",
		Err:      underlyingErr,
	}
	expectedMsg := "failed to create/update snapshot /snapshots/test.snap.json: file already exists"
	if err.Error() != expectedMsg {
		t.Errorf("Error() got %q, want %q", err.Error(), expectedMsg)
	}
	if !errors.Is(err, underlyingErr) {
		t.Errorf("Unwrap() failed, underlying error not found")
	}
}

func TestSnapshotLoadError(t *testing.T) {
	underlyingErr := &JSONUnmarshalError{Source: "test.snap.json", Err: fmt.Errorf("bad json")}
	err := &SnapshotLoadError{
		FilePath: "/snapshots/test.snap.json",
		Err:      underlyingErr,
	}
	expectedMsg := "failed to load/parse snapshot /snapshots/test.snap.json: failed to unmarshal JSON from test.snap.json: bad json"
	if err.Error() != expectedMsg {
		t.Errorf("Error() got %q, want %q", err.Error(), expectedMsg)
	}
	if !errors.Is(err, underlyingErr) {
		t.Errorf("Unwrap() failed, underlying error not found")
	}
}
