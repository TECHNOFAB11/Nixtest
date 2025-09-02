package runner

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/technofab/nixtest/internal/nix"
	"gitlab.com/technofab/nixtest/internal/snapshot"
	"gitlab.com/technofab/nixtest/internal/types"
	"gitlab.com/technofab/nixtest/internal/util"
)

// Runner executes tests based on provided specifications and configuration
type Runner struct {
	config      Config
	nixService  nix.Service
	snapService snapshot.Service
	skipRegex   *regexp.Regexp
	resultsChan chan types.TestResult
	jobsChan    chan types.TestSpec
	wg          sync.WaitGroup
}

// Config holds configuration for Runner
type Config struct {
	NumWorkers      int
	SnapshotDir     string
	UpdateSnapshots bool
	SkipPattern     string
	ImpureEnv       bool
}

func New(cfg Config, nixService nix.Service, snapService snapshot.Service) (*Runner, error) {
	r := &Runner{
		config:      cfg,
		nixService:  nixService,
		snapService: snapService,
	}
	if cfg.SkipPattern != "" {
		var err error
		r.skipRegex, err = regexp.Compile(cfg.SkipPattern)
		if err != nil {
			return nil, fmt.Errorf("failed to compile skip regex: %w", err)
		}
	}
	return r, nil
}

func (r *Runner) shouldSkip(name string) bool {
	if r.skipRegex == nil {
		return false
	}
	return r.skipRegex.MatchString(name)
}

// RunTests executes all tests from the given suites
func (r *Runner) RunTests(suites []types.SuiteSpec) types.Results {
	totalTests := 0
	for _, suite := range suites {
		totalTests += len(suite.Tests)
	}

	r.jobsChan = make(chan types.TestSpec, totalTests)
	r.resultsChan = make(chan types.TestResult, totalTests)

	for i := 1; i <= r.config.NumWorkers; i++ {
		r.wg.Add(1)
		go r.worker()
	}

	for _, suite := range suites {
		for _, test := range suite.Tests {
			test.Suite = suite.Name
			r.jobsChan <- test
		}
	}
	close(r.jobsChan)

	r.wg.Wait()
	close(r.resultsChan)

	results := make(types.Results)
	for res := range r.resultsChan {
		results[res.Spec.Suite] = append(results[res.Spec.Suite], res)
	}
	return results
}

func (r *Runner) worker() {
	defer r.wg.Done()
	for spec := range r.jobsChan {
		r.resultsChan <- r.executeTest(spec)
	}
}

// executeTest -> main test execution logic
func (r *Runner) executeTest(spec types.TestSpec) types.TestResult {
	startTime := time.Now()
	result := types.TestResult{
		Spec:   spec,
		Status: types.StatusSuccess,
	}

	if r.shouldSkip(spec.Name) {
		result.Status = types.StatusSkipped
		result.Duration = time.Since(startTime)
		return result
	}

	var actual any
	var err error

	if spec.ActualDrv != "" {
		actual, err = r.nixService.BuildAndParseJSON(spec.ActualDrv)
		if err != nil {
			result.Status = types.StatusError
			result.ErrorMessage = fmt.Sprintf("[system] failed to build/parse actualDrv %s: %v", spec.ActualDrv, err)
			goto end
		}
	} else {
		actual = spec.Actual
	}

	switch spec.Type {
	case types.TestTypeSnapshot:
		r.handleSnapshotTest(&result, spec, actual)
	case types.TestTypeUnit:
		r.handleUnitTest(&result, spec, actual)
	case types.TestTypeScript:
		r.handleScriptTest(&result, spec)
	default:
		result.Status = types.StatusError
		result.ErrorMessage = fmt.Sprintf("Invalid test type: %s", spec.Type)
	}

end:
	result.Duration = time.Since(startTime)
	return result
}

// handleSnapshotTest processes snapshot type tests
func (r *Runner) handleSnapshotTest(result *types.TestResult, spec types.TestSpec, actual any) {
	snapPath := r.snapService.GetPath(r.config.SnapshotDir, spec.Name)

	if r.config.UpdateSnapshots {
		if err := r.snapService.CreateFile(snapPath, actual); err != nil {
			result.Status = types.StatusError
			result.ErrorMessage = fmt.Sprintf("[system] failed to update snapshot %s: %v", snapPath, err)
			return
		}
		log.Info().Str("test", spec.Name).Str("path", snapPath).Msg("Snapshot updated")
	}

	_, statErr := r.snapService.Stat(snapPath)
	if statErr != nil {
		result.ErrorMessage = fmt.Sprintf("[system] failed to stat snapshot %s: %v", snapPath, statErr)
		result.Status = types.StatusError
		return
	}

	expected, err := r.snapService.LoadFile(snapPath)
	if err != nil {
		result.Status = types.StatusError
		result.ErrorMessage = fmt.Sprintf("[system] failed to parse snapshot %s: %v", snapPath, err)
		return
	}

	r.compareActualExpected(result, actual, expected)
}

// handleUnitTest processes unit type tests
func (r *Runner) handleUnitTest(result *types.TestResult, spec types.TestSpec, actual any) {
	expected := spec.Expected
	r.compareActualExpected(result, actual, expected)
}

// handleScriptTest processes script type tests
func (r *Runner) handleScriptTest(result *types.TestResult, spec types.TestSpec) {
	exitCode, stdout, stderrStr, err := r.nixService.BuildAndRunScript(spec.Script, r.config.ImpureEnv)
	if err != nil {
		result.Status = types.StatusError
		result.ErrorMessage = fmt.Sprintf("[system] failed to run script derivation %s: %v", spec.Script, err)
		return
	}
	if exitCode != 0 {
		result.Status = types.StatusFailure
		result.ErrorMessage = fmt.Sprintf("[exit code %d]\n[stdout]\n%s\n[stderr]\n%s", exitCode, stdout, stderrStr)
	}
}

// compareActualExpected performs the deep equality check and formats diffs
func (r *Runner) compareActualExpected(result *types.TestResult, actual, expected any) {
	if reflect.DeepEqual(actual, expected) {
		// if we already have an error don't overwrite it
		if result.Status != types.StatusError {
			result.Status = types.StatusSuccess
		}
	} else {
		result.Status = types.StatusFailure

		var actualStr, expectedStr string
		var marshalErr error

		if util.IsString(actual) && util.IsString(expected) {
			actualStr = actual.(string)
			expectedStr = expected.(string)
		} else {
			expectedBytes, err := json.MarshalIndent(expected, "", "  ")
			if err != nil {
				marshalErr = fmt.Errorf("[system] failed to marshal 'expected' for diff: %w", err)
			} else {
				expectedStr = string(expectedBytes)
			}

			actualBytes, err := json.MarshalIndent(actual, "", "  ")
			if err != nil && marshalErr == nil {
				marshalErr = fmt.Errorf("[system] failed to marshal 'actual' for diff: %w", err)
			} else if err == nil && marshalErr == nil {
				actualStr = string(actualBytes)
			}
		}

		if marshalErr != nil {
			result.Status = types.StatusError
			result.ErrorMessage = marshalErr.Error()
			return
		}

		result.Expected = expectedStr
		result.Actual = actualStr
	}
}
