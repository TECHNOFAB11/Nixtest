package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"
)

type JUnitReport struct {
	XMLName  xml.Name         `xml:"testsuite"`
	Name     string           `xml:"name,attr"`
	Tests    int              `xml:"tests,attr"`
	Failures int              `xml:"failures,attr"`
	Time     string           `xml:"time,attr"` // in seconds
	Suites   []JUnitTestSuite `xml:"testsuite"`
}

type JUnitTestSuite struct {
	XMLName   xml.Name    `xml:"testsuite"`
	Name      string      `xml:"name,attr"`
	Tests     int         `xml:"tests,attr"`
	Failures  int         `xml:"failures,attr"`
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
}

func GenerateJUnitReport(name string, results Results) (string, error) {
	report := JUnitReport{
		Name:     name,
		Tests:    0,
		Failures: 0,
		Suites:   []JUnitTestSuite{},
	}

	totalDuration := time.Duration(0)

	for suiteName, suiteResults := range results {
		suite := JUnitTestSuite{
			Name:      suiteName,
			Tests:     len(suiteResults),
			Failures:  0,
			TestCases: []JUnitCase{},
		}

		suiteDuration := time.Duration(0)

		for _, result := range suiteResults {
			durationSeconds := fmt.Sprintf("%.3f", result.Duration.Seconds())
			totalDuration += result.Duration
			suiteDuration += result.Duration

			testCase := JUnitCase{
				Name:      result.Name,
				Classname: suiteName, // Use suite name as classname
				Time:      durationSeconds,
			}

			if result.Pos != "" {
				pos := strings.Split(result.Pos, ":")
				testCase.File = pos[0]
				testCase.Line = pos[1]
			}

			if !result.Success {
				suite.Failures++
				report.Failures++
				testCase.Failure = &result.Error
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
