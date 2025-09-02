package runner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	apperrors "gitlab.com/technofab/nixtest/internal/errors"
	"gitlab.com/technofab/nixtest/internal/types"
)

// --- Mock Service Implementations ---

type mockNixService struct {
	BuildDerivationFunc   func(derivation string) (string, error)
	BuildAndParseJSONFunc func(derivation string) (any, error)
	BuildAndRunScriptFunc func(derivation string, impureEnv bool) (exitCode int, stdout string, stderr string, err error)
}

func (m *mockNixService) BuildDerivation(d string) (string, error) {
	if m.BuildDerivationFunc == nil {
		panic("mockNixService.BuildDerivationFunc not set")
	}
	return m.BuildDerivationFunc(d)
}
func (m *mockNixService) BuildAndParseJSON(d string) (any, error) {
	if m.BuildAndParseJSONFunc == nil {
		panic("mockNixService.BuildAndParseJSONFunc not set")
	}
	return m.BuildAndParseJSONFunc(d)
}
func (m *mockNixService) BuildAndRunScript(d string, p bool) (int, string, string, error) {
	if m.BuildAndRunScriptFunc == nil {
		panic("mockNixService.BuildAndRunScriptFunc not set")
	}
	return m.BuildAndRunScriptFunc(d, p)
}

type mockSnapshotService struct {
	GetPathFunc    func(snapshotDir string, testName string) string
	CreateFileFunc func(filePath string, data any) error
	LoadFileFunc   func(filePath string) (any, error)
	StatFunc       func(name string) (os.FileInfo, error)
}

func (m *mockSnapshotService) GetPath(sDir string, tName string) string {
	if m.GetPathFunc == nil { // provide a default if not overridden
		return filepath.Join(sDir, strings.ToLower(strings.ReplaceAll(tName, " ", "_"))+".snap.json")
	}
	return m.GetPathFunc(sDir, tName)
}
func (m *mockSnapshotService) CreateFile(fp string, d any) error {
	if m.CreateFileFunc == nil {
		panic("mockSnapshotService.CreateFileFunc not set")
	}
	return m.CreateFileFunc(fp, d)
}
func (m *mockSnapshotService) LoadFile(fp string) (any, error) {
	if m.LoadFileFunc == nil {
		panic("mockSnapshotService.LoadFileFunc not set")
	}
	return m.LoadFileFunc(fp)
}
func (m *mockSnapshotService) Stat(n string) (os.FileInfo, error) {
	if m.StatFunc == nil {
		panic("mockSnapshotService.StatFunc not set")
	}
	return m.StatFunc(n)
}

// mockFileInfo for snapshot.Stat
type mockFileInfo struct {
	name    string
	isDir   bool
	modTime time.Time
	size    int64
	mode    os.FileMode
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m mockFileInfo) Sys() any           { return nil }

// --- Test Cases ---

func TestNewRunner(t *testing.T) {
	mockNix := &mockNixService{}
	mockSnap := &mockSnapshotService{}
	tests := []struct {
		name        string
		cfg         Config
		wantErr     bool
		skipPattern string
	}{
		{"Valid config, no skip", Config{NumWorkers: 1}, false, ""},
		{"Valid config, valid skip", Config{NumWorkers: 1, SkipPattern: "Test.*"}, false, "Test.*"},
		{"Invalid skip pattern", Config{NumWorkers: 1, SkipPattern: "[invalid"}, true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(tt.cfg, mockNix, mockSnap)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if tt.skipPattern == "" && r.skipRegex != nil {
					t.Errorf("Expected nil skipRegex, got %v", r.skipRegex)
				}
				if tt.skipPattern != "" && (r.skipRegex == nil || r.skipRegex.String() != tt.skipPattern) {
					t.Errorf("Expected skipRegex %q, got %v", tt.skipPattern, r.skipRegex)
				}
			}
		})
	}
}

func TestRunner_executeTest(t *testing.T) {
	tempDir := t.TempDir() // used for snapshotDir in runnerConfig

	tests := []struct {
		name               string
		spec               types.TestSpec
		runnerConfig       Config
		setupMockServices  func(t *testing.T, mockNix *mockNixService, mockSnap *mockSnapshotService, spec types.TestSpec, cfg Config)
		wantStatus         types.TestStatus
		wantErrMsgContains string
		wantActual         string
		wantExpected       string
	}{
		// --- Invalid ---
		{
			name:         "Invalid test type",
			spec:         types.TestSpec{Name: "Invalid", Type: "invalid"},
			runnerConfig: Config{},
			setupMockServices: func(t *testing.T, mNix *mockNixService, mSnap *mockSnapshotService, s types.TestSpec, c Config) {
				// No service calls expected
			},
			wantStatus: types.StatusError,
		},
		// --- Skip ---
		{
			name:         "Skip test due to pattern",
			spec:         types.TestSpec{Name: "SkipThisTest", Type: types.TestTypeUnit},
			runnerConfig: Config{SkipPattern: "SkipThis.*"},
			setupMockServices: func(t *testing.T, mNix *mockNixService, mSnap *mockSnapshotService, s types.TestSpec, c Config) {
				// No service calls expected
			},
			wantStatus: types.StatusSkipped,
		},
		// --- Unit Tests ---
		{
			name:              "Unit test success",
			spec:              types.TestSpec{Name: "UnitSuccess", Type: types.TestTypeUnit, Expected: "hello", Actual: "hello"},
			runnerConfig:      Config{},
			setupMockServices: func(t *testing.T, mNix *mockNixService, mSnap *mockSnapshotService, s types.TestSpec, c Config) {},
			wantStatus:        types.StatusSuccess,
		},
		{
			name:              "Unit test failure",
			spec:              types.TestSpec{Name: "UnitFail", Type: types.TestTypeUnit, Expected: map[string]int{"a": 1}, Actual: map[string]int{"a": 2}},
			runnerConfig:      Config{},
			setupMockServices: func(t *testing.T, mNix *mockNixService, mSnap *mockSnapshotService, s types.TestSpec, c Config) {},
			wantStatus:        types.StatusFailure,
			wantExpected:      "{\n  \"a\": 1\n}",
			wantActual:        "{\n  \"a\": 2\n}",
		},
		{
			name:         "Unit test success with ActualDrv",
			spec:         types.TestSpec{Name: "UnitActualDrvSuccess", Type: types.TestTypeUnit, Expected: map[string]any{"key": "val"}, ActualDrv: "drv.actual"},
			runnerConfig: Config{},
			setupMockServices: func(t *testing.T, mNix *mockNixService, mSnap *mockSnapshotService, s types.TestSpec, c Config) {
				mNix.BuildAndParseJSONFunc = func(derivation string) (any, error) {
					if derivation == "drv.actual" {
						return map[string]any{"key": "val"}, nil
					}
					return nil, fmt.Errorf("unexpected drv: %s", derivation)
				}
			},
			wantStatus: types.StatusSuccess,
		},
		{
			name:         "Unit test error (ActualDrv build fail)",
			spec:         types.TestSpec{Name: "UnitActualDrvError", Type: types.TestTypeUnit, Expected: "any", ActualDrv: "drv.actual.fail"},
			runnerConfig: Config{},
			setupMockServices: func(t *testing.T, mNix *mockNixService, mSnap *mockSnapshotService, s types.TestSpec, c Config) {
				mNix.BuildAndParseJSONFunc = func(derivation string) (any, error) {
					return nil, &apperrors.NixBuildError{Derivation: "drv.actual.fail", Err: errors.New("build failed")}
				}
			},
			wantStatus:         types.StatusError,
			wantErrMsgContains: "failed to build/parse actualDrv drv.actual.fail: nix build for drv.actual.fail failed: build failed",
		},
		// --- Snapshot Tests ---
		{
			name:         "Snapshot test success (existing snapshot match)",
			spec:         types.TestSpec{Name: "SnapSuccess", Type: types.TestTypeSnapshot, Actual: map[string]any{"data": "match"}},
			runnerConfig: Config{SnapshotDir: tempDir},
			setupMockServices: func(t *testing.T, mNix *mockNixService, mSnap *mockSnapshotService, s types.TestSpec, c Config) {
				snapPath := mSnap.GetPath(c.SnapshotDir, s.Name)
				mSnap.StatFunc = func(name string) (os.FileInfo, error) {
					if name == snapPath {
						return mockFileInfo{name: filepath.Base(snapPath)}, nil
					}
					return nil, os.ErrNotExist
				}
				mSnap.LoadFileFunc = func(filePath string) (any, error) {
					if filePath == snapPath {
						return map[string]any{"data": "match"}, nil
					}
					return nil, os.ErrNotExist
				}
			},
			wantStatus: types.StatusSuccess,
		},
		{
			name:         "Snapshot test update (snapshot created, no prior)",
			spec:         types.TestSpec{Name: "SnapUpdateNew", Type: types.TestTypeSnapshot, Actual: map[string]any{"data": "new"}},
			runnerConfig: Config{SnapshotDir: tempDir, UpdateSnapshots: true},
			setupMockServices: func(t *testing.T, mNix *mockNixService, mSnap *mockSnapshotService, s types.TestSpec, c Config) {
				snapPath := mSnap.GetPath(c.SnapshotDir, s.Name)
				mSnap.CreateFileFunc = func(filePath string, data any) error {
					if filePath == snapPath {
						return nil
					}
					return fmt.Errorf("unexpected create path: %s", filePath)
				}
				mSnap.StatFunc = func(name string) (os.FileInfo, error) {
					if name == snapPath {
						return mockFileInfo{name: filepath.Base(snapPath)}, nil
					}
					return nil, os.ErrNotExist
				}
				mSnap.LoadFileFunc = func(filePath string) (any, error) {
					if filePath == snapPath {
						return s.Actual, nil
					}
					return nil, os.ErrNotExist
				}
			},
			wantStatus: types.StatusSuccess,
		},
		// --- Script Tests ---
		{
			name:         "Script test success (exit 0)",
			spec:         types.TestSpec{Name: "ScriptSuccess", Type: types.TestTypeScript, Script: "script.sh"},
			runnerConfig: Config{},
			setupMockServices: func(t *testing.T, mNix *mockNixService, mSnap *mockSnapshotService, s types.TestSpec, c Config) {
				mNix.BuildAndRunScriptFunc = func(derivation string, impureEnv bool) (int, string, string, error) {
					return 0, "stdout", "stderr", nil
				}
			},
			wantStatus: types.StatusSuccess,
		},
		{
			name:         "Script test failure (exit non-0)",
			spec:         types.TestSpec{Name: "ScriptFail", Type: types.TestTypeScript, Script: "script.sh"},
			runnerConfig: Config{},
			setupMockServices: func(t *testing.T, mNix *mockNixService, mSnap *mockSnapshotService, s types.TestSpec, c Config) {
				mNix.BuildAndRunScriptFunc = func(derivation string, impureEnv bool) (int, string, string, error) {
					return 1, "out on fail", "err on fail", nil
				}
			},
			wantStatus:         types.StatusFailure,
			wantErrMsgContains: "[exit code 1]\n[stdout]\nout on fail\n[stderr]\nerr on fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNixSvc := &mockNixService{}
			mockSnapSvc := &mockSnapshotService{}

			tt.setupMockServices(t, mockNixSvc, mockSnapSvc, tt.spec, tt.runnerConfig)

			r, err := New(tt.runnerConfig, mockNixSvc, mockSnapSvc)
			if err != nil {
				t.Fatalf("New() failed: %v", err)
			}

			result := r.executeTest(tt.spec)

			if result.Status != tt.wantStatus {
				t.Errorf("executeTest() status = %s, want %s. ErrorMsg: %s", result.Status, tt.wantStatus, result.ErrorMessage)
			}
			if tt.wantErrMsgContains != "" && !strings.Contains(result.ErrorMessage, tt.wantErrMsgContains) {
				t.Errorf("executeTest() ErrorMessage = %q, want to contain %q", result.ErrorMessage, tt.wantErrMsgContains)
			}
			if result.Status == types.StatusFailure {
				if tt.wantExpected != "" && result.Expected != tt.wantExpected {
					t.Errorf("executeTest() Expected diff string mismatch.\nGot:\n%s\nWant:\n%s", result.Expected, tt.wantExpected)
				}
				if tt.wantActual != "" && result.Actual != tt.wantActual {
					t.Errorf("executeTest() Actual diff string mismatch.\nGot:\n%s\nWant:\n%s", result.Actual, tt.wantActual)
				}
			}
			if result.Duration <= 0 && result.Status != types.StatusSkipped {
				t.Errorf("executeTest() Duration = %v, want > 0", result.Duration)
			}
		})
	}
}

func TestRunner_RunTests(t *testing.T) {
	mockNixSvc := &mockNixService{}
	mockSnapSvc := &mockSnapshotService{}

	mockNixSvc.BuildAndParseJSONFunc = func(derivation string) (any, error) { return "parsed", nil }
	mockNixSvc.BuildAndRunScriptFunc = func(derivation string, impureEnv bool) (int, string, string, error) { return 0, "", "", nil }
	mockSnapSvc.StatFunc = func(name string) (os.FileInfo, error) { return mockFileInfo{}, nil }
	mockSnapSvc.LoadFileFunc = func(filePath string) (any, error) { return "snapshot", nil }
	mockSnapSvc.CreateFileFunc = func(filePath string, data any) error { return nil }

	suites := []types.SuiteSpec{
		{Name: "Suite1", Tests: []types.TestSpec{
			{Name: "S1_Test1_Pass", Type: types.TestTypeUnit, Actual: "a", Expected: "a"},
			{Name: "S1_Test2_Fail", Type: types.TestTypeUnit, Actual: "a", Expected: "b"},
		}},
		{Name: "Suite2", Tests: []types.TestSpec{
			{Name: "S2_Test1_Pass", Type: types.TestTypeUnit, Actual: "c", Expected: "c"},
			{Name: "S2_Test2_SkipThis", Type: types.TestTypeUnit, Actual: "d", Expected: "d"},
		}},
	}

	runnerCfg := Config{NumWorkers: 2, SkipPattern: ".*SkipThis.*"}
	testRunner, err := New(runnerCfg, mockNixSvc, mockSnapSvc)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	results := testRunner.RunTests(suites)

	totalTestsProcessed := 0
	suite1Results, ok1 := results["Suite1"]
	if !ok1 {
		t.Fatalf("Missing results for Suite1")
	}
	totalTestsProcessed += len(suite1Results)

	suite2Results, ok2 := results["Suite2"]
	if !ok2 {
		t.Fatalf("Missing results for Suite2")
	}
	totalTestsProcessed += len(suite2Results)

	if totalTestsProcessed != 4 {
		t.Errorf("RunTests() processed %d tests, want 4", totalTestsProcessed)
	}

	// Check statuses
	foundS1T1, foundS1T2, foundS2T1, foundS2T2 := false, false, false, false
	for _, res := range suite1Results {
		if res.Spec.Name == "S1_Test1_Pass" {
			foundS1T1 = true
			if res.Status != types.StatusSuccess {
				t.Errorf("S1_Test1_Pass status %s, want Success", res.Status)
			}
		}
		if res.Spec.Name == "S1_Test2_Fail" {
			foundS1T2 = true
			if res.Status != types.StatusFailure {
				t.Errorf("S1_Test2_Fail status %s, want Failure", res.Status)
			}
		}
	}
	for _, res := range suite2Results {
		if res.Spec.Name == "S2_Test1_Pass" {
			foundS2T1 = true
			if res.Status != types.StatusSuccess {
				t.Errorf("S2_Test1_Pass status %s, want Success", res.Status)
			}
		}
		if res.Spec.Name == "S2_Test2_SkipThis" {
			foundS2T2 = true
			if res.Status != types.StatusSkipped {
				t.Errorf("S2_Test2_SkipThis status %s, want Skipped", res.Status)
			}
		}
	}
	if !foundS1T1 || !foundS1T2 || !foundS2T1 || !foundS2T2 {
		t.Errorf("Not all tests were found in results map. S1T1:%v, S1T2:%v, S2T1:%v, S2T2:%v", foundS1T1, foundS1T2, foundS2T1, foundS2T2)
	}
}
