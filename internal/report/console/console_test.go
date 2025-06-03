package console

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jedib0t/go-pretty/v6/text"
	"gitlab.com/technofab/nixtest/internal/types"
)

// captureOutput captures stdout and stderr during the execution of a function
func captureOutput(f func()) (string, string) {
	// save original stdout and stderr
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	defer func() {
		// restore stdout & stderr
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	}()

	outC := make(chan string)
	errC := make(chan string)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		_, _ = io.Copy(&buf, rOut)
		outC <- buf.String()
	}()

	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		_, _ = io.Copy(&buf, rErr)
		errC <- buf.String()
	}()

	f()

	wOut.Close()
	wErr.Close()
	wg.Wait()

	stdout := <-outC
	stderr := <-errC

	return stdout, stderr
}

func TestPrintErrorsColor(t *testing.T) {
	results := types.Results{
		"Suite1": []types.TestResult{
			{
				Spec:     types.TestSpec{Suite: "Suite1", Name: "TestFailure_Diff"},
				Status:   types.StatusFailure,
				Expected: "line1\nline2 expected\nline3",
				Actual:   "line1\nline2 actual\nline3",
			},
		},
	}
	stdout, _ := captureOutput(func() {
		PrintErrors(results, false)
	})

	ansiEscapePattern := `(?:\\x1b\[[0-9;]*m)*`
	pattern := `.*\n` +
		`A|A Diff:\n` +
		`A|A line1\n` +
		`A|A line2 AexpeAAacAAtedAAualA\n` +
		`A|A line3.*`
	pattern = strings.ReplaceAll(pattern, "A", ansiEscapePattern)

	matched, _ := regexp.MatchString(pattern, stdout)

	if !matched {
		t.Errorf("PrintErrors() TestFailure_Diff diff output mismatch or missing.\nExpected pattern:\n%s\nGot:\n%s", pattern, stdout)
	}
}

func TestPrintErrors(t *testing.T) {
	text.DisableColors()
	defer text.EnableColors()

	results := types.Results{
		"Suite1": []types.TestResult{
			{
				Spec:   types.TestSpec{Suite: "Suite1", Name: "TestSuccess"},
				Status: types.StatusSuccess,
			},
			{
				Spec:     types.TestSpec{Suite: "Suite1", Name: "TestFailure_Diff"},
				Status:   types.StatusFailure,
				Expected: "line1\nline2 expected\nline3",
				Actual:   "line1\nline2 actual\nline3",
			},
			{
				Spec:         types.TestSpec{Suite: "Suite1", Name: "TestFailure_Message"},
				Status:       types.StatusFailure,
				ErrorMessage: "This is a specific failure message.\nWith multiple lines.",
			},
			{
				Spec:         types.TestSpec{Suite: "Suite1", Name: "TestError"},
				Status:       types.StatusError,
				ErrorMessage: "System error occurred.",
			},
			{
				Spec:         types.TestSpec{Suite: "Suite1", Name: "TestEmpty"},
				Status:       types.StatusError,
				ErrorMessage: "",
			},
		},
	}

	stdout, _ := captureOutput(func() {
		PrintErrors(results, true)
	})

	if strings.Contains(stdout, "TestSuccess") {
		t.Errorf("PrintErrors() should not print success cases, but found 'TestSuccess'")
	}

	expectedDiffPattern := `\|\s*--- expected\s*\n` + // matches "| --- expected"
		`\|\s*\+\+\+ actual\s*\n` + // matches "| +++ actual"
		`\|\s*@@ -\d+,\d+ \+\d+,\d+ @@\s*\n` + // matches "| @@ <hunk info> @@"
		`\|\s* line1\s*\n` + // matches "|  line1" (note the leading space for an "equal" line)
		`\|\s*-line2 expected\s*\n` + // matches "| -line2 expected"
		`\|\s*\+line2 actual\s*\n` + // matches "| +line2 actual"
		`\|\s* line3\s*` // matches "|  line3"

	matched, _ := regexp.MatchString(expectedDiffPattern, stdout)

	if !matched {
		t.Errorf("PrintErrors() TestFailure_Diff diff output mismatch or missing.\nExpected pattern:\n%s\nGot:\n%s", expectedDiffPattern, stdout)
	}

	if !strings.Contains(stdout, "⚠ Test \"Suite1/TestFailure_Message\" failed:") {
		t.Errorf("PrintErrors() missing header for TestFailure_Message. Output:\n%s", stdout)
	}

	if !strings.Contains(stdout, "| This is a specific failure message.") ||

		!strings.Contains(stdout, "| With multiple lines.") {
		t.Errorf("PrintErrors() TestFailure_Message message output mismatch or missing. Output:\n%s", stdout)
	}

	if !strings.Contains(stdout, "⚠ Test \"Suite1/TestError\" failed:") {
		t.Errorf("PrintErrors() missing header for TestError. Output:\n%s", stdout)
	}

	if !strings.Contains(stdout, "| System error occurred.") {
		t.Errorf("PrintErrors() TestError message output mismatch or missing. Output:\n%s", stdout)
	}
	if !strings.Contains(stdout, "- no output -") {
		t.Errorf("PrintErrors() missing '- no output -'. Output:\n%s", stdout)
	}
}

func TestPrintSummary(t *testing.T) {
	text.DisableColors()
	defer text.EnableColors()

	results := types.Results{
		"AlphaSuite": []types.TestResult{
			{Spec: types.TestSpec{Suite: "AlphaSuite", Name: "TestA", Pos: "alpha.nix:1"}, Status: types.StatusSuccess, Duration: 100 * time.Millisecond},
			{Spec: types.TestSpec{Suite: "AlphaSuite", Name: "TestB", Pos: "alpha.nix:2"}, Status: types.StatusFailure, Duration: 200 * time.Millisecond},
		},
		"BetaSuite": []types.TestResult{
			{Spec: types.TestSpec{Suite: "BetaSuite", Name: "TestC", Pos: "beta.nix:1"}, Status: types.StatusSkipped, Duration: 50 * time.Millisecond},
			{Spec: types.TestSpec{Suite: "BetaSuite", Name: "TestD", Pos: "beta.nix:2"}, Status: types.StatusError, Duration: 150 * time.Millisecond},
			{Spec: types.TestSpec{Suite: "BetaSuite", Name: "TestE", Pos: "beta.nix:3"}, Status: 123, Duration: 150 * time.Millisecond},
		},
	}
	totalSuccessCount := 2
	totalTestCount := 4

	r, w, _ := os.Pipe()
	originalStdout := os.Stdout
	os.Stdout = w

	PrintSummary(results, totalSuccessCount, totalTestCount)

	w.Close()
	os.Stdout = originalStdout

	var summaryTable bytes.Buffer
	_, _ = io.Copy(&summaryTable, r)
	stdout := summaryTable.String()

	if !strings.Contains(stdout, "AlphaSuite") || !strings.Contains(stdout, "BetaSuite") {
		t.Errorf("PrintSummary() missing suite names. Output:\n%s", stdout)
	}

	// check for test names and statuses
	if !strings.Contains(stdout, "TestA") || !strings.Contains(stdout, "PASS") {
		t.Errorf("PrintSummary() missing TestA or its PASS status. Output:\n%s", stdout)
	}
	if !strings.Contains(stdout, "TestB") || !strings.Contains(stdout, "FAIL") {
		t.Errorf("PrintSummary() missing TestB or its FAIL status. Output:\n%s", stdout)
	}
	if !strings.Contains(stdout, "TestE") || !strings.Contains(stdout, "UNKNOWN") {
		t.Errorf("PrintSummary() missing TestE or its UNKNOWN status. Output:\n%s", stdout)
	}

	// check for total summary
	expectedTotalSummary := fmt.Sprintf("%d/%d (1 SKIPPED)", totalSuccessCount, totalTestCount)
	if !strings.Contains(stdout, expectedTotalSummary) {
		t.Errorf("PrintSummary() total summary incorrect. Expected to contain '%s'. Output:\n%s", expectedTotalSummary, stdout)
	}
}
