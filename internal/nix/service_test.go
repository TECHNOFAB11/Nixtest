package nix

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	apperrors "gitlab.com/TECHNOFAB/nixtest/internal/errors"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 && args[0] != "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "No command after --")
		os.Exit(1)
	}
	args = args[1:]

	cmd, params := args[0], args[1:]

	switch cmd {
	case "nix":
		if len(params) > 0 && params[0] == "build" {
			mockOutput := os.Getenv("MOCK_NIX_BUILD_OUTPUT")
			mockError := os.Getenv("MOCK_NIX_BUILD_ERROR")
			mockExitCode := os.Getenv("MOCK_NIX_BUILD_EXIT_CODE")

			if mockError != "" {
				fmt.Fprintln(os.Stderr, mockError)
			}
			if mockExitCode != "" && mockExitCode != "0" {
				os.Exit(1) // simplified exit for helper
			}
			if mockError == "" && (mockExitCode == "" || mockExitCode == "0") {
				fmt.Fprintln(os.Stdout, mockOutput)
			}
		}
	case "bash", "env":
		scriptPath := params[0]
		if cmd == "env" && len(params) > 2 {
			scriptPath = params[2]
		}
		if _, err := os.Stat(scriptPath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "mocked script: script path %s could not be statted: %v\n", scriptPath, err)
			os.Exit(3)
		}
		fmt.Fprint(os.Stdout, os.Getenv("MOCK_SCRIPT_STDOUT"))
		fmt.Fprint(os.Stderr, os.Getenv("MOCK_SCRIPT_STDERR"))
		if code := os.Getenv("MOCK_SCRIPT_EXIT_CODE"); code != "" && code != "0" {
			os.Exit(5) // custom exit for script failure
		}
	default:
		fmt.Fprintf(os.Stderr, "mocked command: unknown command %s\n", cmd)
		os.Exit(126)
	}
}

// mockExecCommand configures the DefaultService to use the test helper
func mockExecCommandForService(service *DefaultService) {
	service.commandExecutor = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		cmd.Env = append(cmd.Env, os.Environ()...)
		return cmd
	}
}

func TestDefaultService_BuildDerivation(t *testing.T) {
	service := NewDefaultService()
	mockExecCommandForService(service) // configure service to use helper

	tests := []struct {
		name               string
		derivation         string
		mockOutput         string
		mockError          string
		mockExitCode       string
		wantPath           string
		wantErr            bool
		wantErrType        any
		wantErrMsgContains string
	}{
		{"Success", "some.drv#attr", "/nix/store/mock-path", "", "0", "/nix/store/mock-path", false, nil, ""},
		{
			"Nix command error", "error.drv#attr", "", "nix error details", "1", "", true,
			(*apperrors.NixBuildError)(nil), "nix error details",
		},
		{
			"Nix command success but no output path", "empty.drv#attr", "", "", "0", "", true,
			(*apperrors.NixNoOutputPathError)(nil), "produced no output path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("MOCK_NIX_BUILD_OUTPUT", tt.mockOutput)
			os.Setenv("MOCK_NIX_BUILD_ERROR", tt.mockError)
			os.Setenv("MOCK_NIX_BUILD_EXIT_CODE", tt.mockExitCode)
			defer func() {
				os.Unsetenv("MOCK_NIX_BUILD_OUTPUT")
				os.Unsetenv("MOCK_NIX_BUILD_ERROR")
				os.Unsetenv("MOCK_NIX_BUILD_EXIT_CODE")
			}()

			gotPath, err := service.BuildDerivation(tt.derivation)

			if (err != nil) != tt.wantErr {
				t.Fatalf("BuildDerivation() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.wantErrType != nil && !errors.As(err, &tt.wantErrType) {
					t.Errorf("BuildDerivation() error type = %T, want %T", err, tt.wantErrType)
				}
				if tt.wantErrMsgContains != "" && !strings.Contains(err.Error(), tt.wantErrMsgContains) {
					t.Errorf("BuildDerivation() error = %q, want error containing %q", err.Error(), tt.wantErrMsgContains)
				}
			}
			if !tt.wantErr && gotPath != tt.wantPath {
				t.Errorf("BuildDerivation() gotPath = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestDefaultService_BuildAndParseJSON(t *testing.T) {
	service := NewDefaultService()
	mockExecCommandForService(service)

	tempDir := t.TempDir()
	mockDrvOutputPath := filepath.Join(tempDir, "drv_output.json")

	tests := []struct {
		name               string
		derivation         string
		mockBuildOutput    string
		mockJSONContent    string
		mockBuildError     string
		mockBuildExitCode  string
		want               any
		wantErr            bool
		wantErrType        any
		wantErrMsgContains string
	}{
		{
			"Success", "some.drv#json", mockDrvOutputPath, `{"key": "value"}`, "", "0",
			map[string]any{"key": "value"}, false, nil, "",
		},
		{
			"BuildDerivation fails", "error.drv#json", "", "", "nix build error", "1",
			nil, true, (*apperrors.NixBuildError)(nil), "nix build error",
		},
		{
			"ReadFile fails", "readfail.drv#json", "/nonexistent/path/output.json", "", "", "0",
			nil, true, (*apperrors.FileReadError)(nil), "failed to read file",
		},
		{
			"Unmarshal fails", "badjson.drv#json", mockDrvOutputPath, `{"key": "value"`, "", "0",
			nil, true, (*apperrors.JSONUnmarshalError)(nil), "failed to unmarshal JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("MOCK_NIX_BUILD_OUTPUT", tt.mockBuildOutput)
			os.Setenv("MOCK_NIX_BUILD_ERROR", tt.mockBuildError)
			os.Setenv("MOCK_NIX_BUILD_EXIT_CODE", tt.mockBuildExitCode)

			if tt.mockJSONContent != "" && tt.mockBuildOutput == mockDrvOutputPath {
				if err := os.WriteFile(mockDrvOutputPath, []byte(tt.mockJSONContent), 0644); err != nil {
					t.Fatalf("Failed to write mock JSON content: %v", err)
				}
				defer os.Remove(mockDrvOutputPath)
			}

			got, err := service.BuildAndParseJSON(tt.derivation)

			if (err != nil) != tt.wantErr {
				t.Fatalf("BuildAndParseJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.wantErrType != nil && !errors.As(err, &tt.wantErrType) {
					t.Errorf("BuildAndParseJSON() error type = %T, want %T", err, tt.wantErrType)
				}
				if tt.wantErrMsgContains != "" && !strings.Contains(err.Error(), tt.wantErrMsgContains) {
					t.Errorf("BuildAndParseJSON() error = %q, want error containing %q", err.Error(), tt.wantErrMsgContains)
				}
			}
			if !tt.wantErr && !jsonDeepEqual(got, tt.want) {
				t.Errorf("BuildAndParseJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func jsonDeepEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	jsonA, _ := json.Marshal(a)
	jsonB, _ := json.Marshal(b)
	return string(jsonA) == string(jsonB)
}

func TestDefaultService_BuildAndRunScript(t *testing.T) {
	service := NewDefaultService()
	mockExecCommandForService(service)

	tempDir := t.TempDir()
	mockScriptPath := filepath.Join(tempDir, "mock_script.sh")
	if err := os.WriteFile(mockScriptPath, []byte("#!/bin/bash\necho hello"), 0755); err != nil {
		t.Fatalf("Failed to create dummy mock script: %v", err)
	}

	tests := []struct {
		name                 string
		derivation           string
		impureEnv            bool
		mockBuildDrvOutput   string
		mockBuildDrvError    string
		mockBuildDrvExitCode string
		mockScriptStdout     string
		mockScriptStderr     string
		mockScriptExitCode   string
		wantExitCode         int
		wantStdout           string
		wantStderr           string
		wantErr              bool
		wantErrType          any
		wantErrMsgContains   string
	}{
		{
			"Success", "script.drv#sh", false, mockScriptPath, "", "0",
			"Hello", "ErrOut", "0",
			0, "Hello", "ErrOut", false, nil, "",
		},
		{
			"Success impure", "script.drv#sh", true, mockScriptPath, "", "0",
			"Hello", "ErrOut", "0",
			0, "Hello", "ErrOut", false, nil, "",
		},
		{
			"Script fails (non-zero exit)", "fail.drv#sh", false, mockScriptPath, "", "0",
			"Out", "Err", "custom", // helper uses 5 for script failure
			5, "Out", "Err", false, nil, "", // error is nil, non-zero exit code
		},
		{
			"BuildDerivation fails", "buildfail.drv#sh", false, "", "nix error", "1",
			"", "", "",
			-1, "", "", true, (*apperrors.NixBuildError)(nil), "nix error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("MOCK_NIX_BUILD_OUTPUT", tt.mockBuildDrvOutput)
			os.Setenv("MOCK_NIX_BUILD_ERROR", tt.mockBuildDrvError)
			os.Setenv("MOCK_NIX_BUILD_EXIT_CODE", tt.mockBuildDrvExitCode)
			os.Setenv("MOCK_SCRIPT_STDOUT", tt.mockScriptStdout)
			os.Setenv("MOCK_SCRIPT_STDERR", tt.mockScriptStderr)
			os.Setenv("MOCK_SCRIPT_EXIT_CODE", tt.mockScriptExitCode)

			exitCode, stdout, stderr, err := service.BuildAndRunScript(tt.derivation, tt.impureEnv)

			if (err != nil) != tt.wantErr {
				t.Fatalf("BuildAndRunScript() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.wantErrType != nil && !errors.As(err, &tt.wantErrType) {
					t.Errorf("BuildAndRunScript() error type = %T, want %T", err, tt.wantErrType)
				}
				if tt.wantErrMsgContains != "" && !strings.Contains(err.Error(), tt.wantErrMsgContains) {
					t.Errorf("BuildAndRunScript() error = %q, want error containing %q", err.Error(), tt.wantErrMsgContains)
				}
			} else {
				if exitCode != tt.wantExitCode {
					t.Errorf("BuildAndRunScript() exitCode = %v, want %v", exitCode, tt.wantExitCode)
				}
				if stdout != tt.wantStdout {
					t.Errorf("BuildAndRunScript() stdout = %q, want %q", stdout, tt.wantStdout)
				}
				if stderr != tt.wantStderr {
					t.Errorf("BuildAndRunScript() stderr = %q, want %q", stderr, tt.wantStderr)
				}
			}
		})
	}
}
