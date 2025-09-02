package nix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	apperrors "gitlab.com/technofab/nixtest/internal/errors"
)

// Service defines operations related to Nix
type Service interface {
	BuildDerivation(derivation string) (string, error)
	BuildAndParseJSON(derivation string) (any, error)
	BuildAndRunScript(derivation string, impureEnv bool) (exitCode int, stdout string, stderr string, err error)
}

type DefaultService struct {
	commandExecutor func(command string, args ...string) *exec.Cmd
}

func NewDefaultService() *DefaultService {
	return &DefaultService{commandExecutor: exec.Command}
}

// BuildDerivation builds a Nix derivation and returns the output path
func (s *DefaultService) BuildDerivation(derivation string) (string, error) {
	cmd := s.commandExecutor(
		"nix",
		"build",
		derivation+"^*",
		"--print-out-paths",
		"--no-link",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", &apperrors.NixBuildError{Derivation: derivation, Stderr: stderr.String(), Err: err}
	}

	path := strings.TrimSpace(stdout.String())
	if path == "" {
		return "", &apperrors.NixNoOutputPathError{Derivation: derivation, Stderr: stderr.String()}
	}
	return path, nil
}

// BuildAndParseJSON builds a derivation and parses its output file as JSON
func (s *DefaultService) BuildAndParseJSON(derivation string) (any, error) {
	path, err := s.BuildDerivation(derivation)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &apperrors.FileReadError{Path: path, Err: err}
	}

	var result any
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, &apperrors.JSONUnmarshalError{Source: path, Err: err}
	}

	return result, nil
}

// BuildAndRunScript builds a derivation and runs it as a script
func (s *DefaultService) BuildAndRunScript(derivation string, impureEnv bool) (exitCode int, stdout string, stderr string, err error) {
	exitCode = -1
	path, err := s.BuildDerivation(derivation)
	if err != nil {
		return exitCode, "", "", err
	}

	// run scripts in a temporary directory
	tempDir, err := os.MkdirTemp("", "nixtest-script-")
	if err != nil {
		return exitCode, "", "", &apperrors.ScriptExecutionError{Path: path, Err: fmt.Errorf("failed to create temporary directory: %w", err)}
	}
	defer os.RemoveAll(tempDir)

	var cmdArgs []string
	if impureEnv {
		cmdArgs = []string{"bash", path}
	} else {
		cmdArgs = append([]string{"env", "-i"}, "bash", path)
	}

	cmd := s.commandExecutor(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = tempDir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err = cmd.Start(); err != nil {
		return exitCode, "", "", &apperrors.ScriptExecutionError{Path: path, Err: err}
	}

	runErr := cmd.Wait()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			return exitErr.ExitCode(), stdout, stderr, nil
		}
		return exitCode, stdout, stderr, &apperrors.ScriptExecutionError{Path: path, Err: runErr}
	}

	return 0, stdout, stderr, nil
}
