package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/akedrou/textdiff"
	"github.com/akedrou/textdiff/myers"
)

type JUnitReport struct {
	XMLName  xml.Name         `xml:"testsuite"`
	Name     string           `xml:"name,attr"`
	Tests    int              `xml:"tests,attr"`
	Failures int              `xml:"failures,attr"`
	Errors   int              `xml:"errors,attr"`
	Skipped  int              `xml:"skipped,attr"`
	Time     string           `xml:"time,attr"` // in seconds
	Suites   []JUnitTestSuite `xml:"testsuite"`
}

type JUnitTestSuite struct {
	XMLName   xml.Name    `xml:"testsuite"`
	Name      string      `xml:"name,attr"`
	Tests     int         `xml:"tests,attr"`
	Failures  int         `xml:"failures,attr"`
	Errors    int         `xml:"errors,attr"`
	Skipped   int         `xml:"skipped,attr"`
	Time      string      `xml:"time,attr"` // in seconds
	TestCases []JUnitCase `xml:"testcase"`
}

type JUnitCase struct {
	Name      string  `xml:"name,attr"`
	Classname string  `xml:"classname,attr"`
	Time      string  `xml:"time,attr"` // in seconds
	File      string  `xml:"file,attr,omitempty"`
	Line      string  `xml:"line,attr,omitempty"`
	Failure   *string `xml:"failure,omitempty"`
	Error     *string `xml:"error,omitempty"`
	Skipped   *string `xml:"skipped,omitempty"`
}

func GenerateJUnitReport(name string, results Results) (string, error) {
	report := JUnitReport{
		Name:     name,
		Tests:    0,
		Failures: 0,
		Errors:   0,
		Skipped:  0,
		Suites:   []JUnitTestSuite{},
	}

	totalDuration := time.Duration(0)

	for suiteName, suiteResults := range results {
		suite := JUnitTestSuite{
			Name:      suiteName,
			Tests:     len(suiteResults),
			Failures:  0,
			Errors:    0,
			Skipped:   0,
			TestCases: []JUnitCase{},
		}

		suiteDuration := time.Duration(0)

		for _, result := range suiteResults {
			durationSeconds := fmt.Sprintf("%.3f", result.Duration.Seconds())
			totalDuration += result.Duration
			suiteDuration += result.Duration

			testCase := JUnitCase{
				Name:      result.Spec.Name,
				Classname: suiteName, // Use suite name as classname
				Time:      durationSeconds,
			}

			if result.Spec.Pos != "" {
				pos := strings.Split(result.Spec.Pos, ":")
				testCase.File = pos[0]
				testCase.Line = pos[1]
			}

			if result.Status == StatusFailure {
				suite.Failures++
				report.Failures++
				// FIXME: ComputeEdits deprecated
				edits := myers.ComputeEdits(result.Expected, result.Actual)
				diff, err := textdiff.ToUnified("expected", "actual", result.Expected, edits, 3)
				if err != nil {
					return "", err
				}
				// remove newline hint
				diff = strings.ReplaceAll(diff, "\\ No newline at end of file\n", "")
				testCase.Failure = &diff
			} else if result.Status == StatusError {
				suite.Errors++
				report.Errors++
				testCase.Error = &result.ErrorMessage
			} else if result.Status == StatusSkipped {
				suite.Skipped++
				report.Skipped++
				skipped := ""
				testCase.Skipped = &skipped
			}

			suite.TestCases = append(suite.TestCases, testCase)
		}
		suite.Time = fmt.Sprintf("%.3f", suiteDuration.Seconds())
		report.Suites = append(report.Suites, suite)
		report.Tests += len(suiteResults)
	}

	report.Time = fmt.Sprintf("%.3f", totalDuration.Seconds())

	output, err := xml.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal XML: %w", err)
	}

	return xml.Header + string(output), nil
}

func GenerateJunitFile(path string, results Results) error {
	res, err := GenerateJUnitReport("nixtest", results)
	if err != nil {
		return fmt.Errorf("failed to generate junit report: %w", err)
	}
	file, err := os.Create(*junitPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(res)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}
