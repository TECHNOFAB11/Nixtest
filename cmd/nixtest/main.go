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
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sergi/go-diff/diffmatchpatch"
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
	Pos         string `json:"pos,omitempty"`

	Suite string
}

type TestResult struct {
	Name     string
	Success  bool
	Error    string
	Duration time.Duration
	Pos      string
	Suite    string
}

type Results map[string][]TestResult

func buildAndParse(derivation string) (any, error) {
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
		return nil, fmt.Errorf("failed to run nix build: %v, %s", err, stderr.String())
	}

	path := strings.TrimSpace(stdout.String())

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

func runTest(spec TestSpec) TestResult {
	startTime := time.Now()
	result := TestResult{
		Name:    spec.Name,
		Pos:     spec.Pos,
		Suite:   spec.Suite,
		Success: false,
		Error:   "",
	}

	var actual any
	var expected any

	if spec.ActualDrv != "" {
		var err error
		actual, err = buildAndParse(spec.ActualDrv)
		if err != nil {
			result.Error = fmt.Sprintf("[system] failed to parse drv output: %v", err.Error())
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
			result.Error = "No Snapshot exists yet"
			goto end
		}

		var err error
		expected, err = ParseFile[any](filePath)
		if err != nil {
			result.Error = fmt.Sprintf("[system] failed to parse snapshot: %v", err.Error())
			goto end
		}
	} else if spec.Type == "unit" {
		expected = spec.Expected
	} else {
		log.Panic().Str("type", spec.Type).Msg("Invalid test type")
	}

	if reflect.DeepEqual(actual, expected) {
		result.Success = true
	} else {
		dmp := diffmatchpatch.New()
		text1, err := json.MarshalIndent(actual, "", "  ")
		if err != nil {
			result.Error = fmt.Sprintf("[system] failed to json marshal 'actual': %v", err.Error())
			goto end
		}
		text2, err := json.MarshalIndent(expected, "", "  ")
		if err != nil {
			result.Error = fmt.Sprintf("[system] failed to json marshal 'expected': %v", err.Error())
			goto end
		}
		diffs := dmp.DiffMain(string(text1), string(text2), false)
		result.Error = fmt.Sprintf("Mismatch:\n%s", dmp.DiffPrettyText(diffs))
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
	updateSnapshots *bool = flag.Bool("update-snapshots", false, "Update all snapshots")
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
		results[r.Suite] = append(results[r.Suite], r)
		if r.Success {
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
