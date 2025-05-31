package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
)

type SuiteSpec struct {
	Name  string     `json:"name"`
	Tests []TestSpec `json:"tests"`
}

type TestSpec struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Expected    any    `json:"expected,omitempty"`
	Actual      any    `json:"actual,omitempty"`
	ActualDrv   string `json:"actualDrv,omitempty"`
	Script      string `json:"script,omitempty"`
	Pos         string `json:"pos,omitempty"`

	Suite string
}

type TestStatus int

const (
	StatusSuccess TestStatus = iota
	StatusFailure
	StatusError
	StatusSkipped
)

type TestResult struct {
	Spec         TestSpec
	Status       TestStatus
	Duration     time.Duration
	ErrorMessage string
	Expected     string
	Actual       string
}

type Results map[string][]TestResult

func buildDerivation(derivation string) (string, error) {
	cmd := exec.Command(
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
		return "", fmt.Errorf("failed to run nix build: %v, %s", err, stderr.String())
	}

	path := strings.TrimSpace(stdout.String())
	return path, nil
}

func buildAndParse(derivation string) (any, error) {
	path, err := buildDerivation(derivation)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result any
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// builds derivation and runs it
func buildAndRun(derivation string) (exitCode int, stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	exitCode = -1
	path, err := buildDerivation(derivation)
	if err != nil {
		return
	}

	cmd := exec.Command("bash", path)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err = cmd.Start(); err != nil {
		return
	}

	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			exitCode = exiterr.ExitCode()
			err = nil
		}
		return
	}

	return 0, stdout, stderr, nil
}

func PrefixLines(input string) string {
	lines := strings.Split(input, "\n")
	for i := range lines {
		lines[i] = "| " + lines[i]
	}
	return strings.Join(lines, "\n")
}

func shouldSkip(input string, pattern string) bool {
	if pattern == "" {
		return false
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		log.Panic().Err(err).Msg("failed to compile skip regex")
	}

	return regex.MatchString(input)
}

func isString(value any) bool {
	switch value.(type) {
	case string:
		return true
	default:
		return false
	}
}

func runTest(spec TestSpec) TestResult {
	startTime := time.Now()
	result := TestResult{
		Spec:   spec,
		Status: StatusSuccess,
	}

	if shouldSkip(spec.Name, *skipPattern) {
		result.Status = StatusSkipped
		return result
	}

	var actual any
	var expected any

	if spec.ActualDrv != "" {
		var err error
		actual, err = buildAndParse(spec.ActualDrv)
		if err != nil {
			result.Status = StatusError
			result.ErrorMessage = fmt.Sprintf("[system] failed to parse drv output: %v", err.Error())
			goto end
		}
	} else {
		actual = spec.Actual
	}
	if spec.Type == "snapshot" {
		filePath := path.Join(
			*snapshotDir,
			fmt.Sprintf("%s.snap.json", strings.ToLower(spec.Name)),
		)

		if *updateSnapshots {
			createSnapshot(filePath, actual)
		}

		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			result.Status = StatusError
			result.ErrorMessage = "No Snapshot exists yet"
			goto end
		}

		var err error
		expected, err = ParseFile[any](filePath)
		if err != nil {
			result.Status = StatusError
			result.ErrorMessage = fmt.Sprintf("[system] failed to parse snapshot: %v", err.Error())
			goto end
		}
	} else if spec.Type == "unit" {
		expected = spec.Expected
	} else if spec.Type == "script" {
		exitCode, stdout, stderr, err := buildAndRun(spec.Script)
		if err != nil {
			result.Status = StatusError
			result.ErrorMessage = fmt.Sprintf("[system] failed to run script: %v", err.Error())
		}
		if exitCode != 0 {
			result.Status = StatusFailure
			result.ErrorMessage = fmt.Sprintf("[exit code %d]\n[stdout]\n%s\n[stderr]\n%s", exitCode, stdout.String(), stderr.String())
		}
		// no need for equality checking with "script"
		goto end
	} else {
		log.Panic().Str("type", spec.Type).Msg("Invalid test type")
	}

	if reflect.DeepEqual(actual, expected) {
		result.Status = StatusSuccess
	} else {
		var text1, text2 string

		// just keep strings as is, only json marshal if any of them is not a string
		if isString(actual) && isString(expected) {
			text1 = actual.(string)
			text2 = expected.(string)
		} else {
			bytes1, err := json.MarshalIndent(expected, "", "  ")
			if err != nil {
				result.Status = StatusError
				result.ErrorMessage = fmt.Sprintf("[system] failed to json marshal 'expected': %v", err.Error())
				goto end
			}
			bytes2, err := json.MarshalIndent(actual, "", "  ")
			if err != nil {
				result.Status = StatusError
				result.ErrorMessage = fmt.Sprintf("[system] failed to json marshal 'actual': %v", err.Error())
				goto end
			}
			text1 = string(bytes1)
			text2 = string(bytes2)
		}

		result.Status = StatusFailure
		result.Expected = text1
		result.Actual = text2
	}

end:
	result.Duration = time.Since(startTime)
	return result
}

func createSnapshot(filePath string, actual any) error {
	jsonData, err := json.Marshal(actual)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(filePath), 0777)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// worker to process TestSpec items
func worker(jobs <-chan TestSpec, results chan<- TestResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for spec := range jobs {
		results <- runTest(spec)
	}
}

var (
	numWorkers  *int    = flag.Int("workers", 4, "Amount of tests to run in parallel")
	testsFile   *string = flag.String("tests", "", "Path to JSON file containing tests")
	snapshotDir *string = flag.String(
		"snapshot-dir", "./snapshots", "Directory where snapshots are stored",
	)
	junitPath *string = flag.String(
		"junit", "", "Path to generate JUNIT report to, leave empty to disable",
	)
	updateSnapshots *bool   = flag.Bool("update-snapshots", false, "Update all snapshots")
	skipPattern     *string = flag.String("skip", "", "Regular expression to skip (e.g., 'test-.*|.*-b')")
)

func main() {
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().
		Int("workers", *numWorkers).
		Msg("Starting nixtest")

	if _, err := os.Stat(*testsFile); errors.Is(err, os.ErrNotExist) {
		log.Error().Str("file", *testsFile).Msg("Tests file does not exist")
		os.Exit(1)
	}

	parsedSpecs, err := ParseFile[[]SuiteSpec](*testsFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load tests from file")
		os.Exit(1)
	}

	totalTests := 0
	for _, suite := range parsedSpecs {
		totalTests += len(suite.Tests)
	}
	log.Info().
		Int("suites", len(parsedSpecs)).
		Int("tests", totalTests).
		Msg("Discovered suites")

	jobsChan := make(chan TestSpec, totalTests)
	resultsChan := make(chan TestResult, totalTests)

	var wg sync.WaitGroup

	for i := 1; i <= *numWorkers; i++ {
		wg.Add(1)
		go worker(jobsChan, resultsChan, &wg)
	}

	for _, suite := range parsedSpecs {
		for _, test := range suite.Tests {
			test.Suite = suite.Name
			jobsChan <- test
		}
	}
	close(jobsChan)

	wg.Wait()
	close(resultsChan)

	results := map[string][]TestResult{}

	successCount := 0

	for r := range resultsChan {
		results[r.Spec.Suite] = append(results[r.Spec.Suite], r)
		if r.Status == StatusSuccess || r.Status == StatusSkipped {
			successCount++
		}
	}

	if *junitPath != "" {
		err = GenerateJunitFile(*junitPath, results)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate junit file")
		} else {
			log.Info().Str("path", *junitPath).Msg("Generated Junit report")
		}
	}

	// print errors/logs of failed tests
	printErrors(results)

	// show table summary
	printSummary(results, successCount, totalTests)

	if successCount != totalTests {
		os.Exit(2)
	}
}
