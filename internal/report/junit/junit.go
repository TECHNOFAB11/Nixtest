package junit

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"

	"gitlab.com/technofab/nixtest/internal/types"
	"gitlab.com/technofab/nixtest/internal/util"
)

type JUnitReport struct {
	XMLName  xml.Name         `xml:"testsuites"`
	Name     string           `xml:"name,attr,omitempty"`
	Tests    int              `xml:"tests,attr"`
	Failures int              `xml:"failures,attr"`
	Errors   int              `xml:"errors,attr"`
	Skipped  int              `xml:"skipped,attr"`
	Time     string           `xml:"time,attr"`
	Suites   []JUnitTestSuite `xml:"testsuite"`
}

type JUnitTestSuite struct {
	XMLName   xml.Name    `xml:"testsuite"`
	Name      string      `xml:"name,attr"`
	Tests     int         `xml:"tests,attr"`
	Failures  int         `xml:"failures,attr"`
	Errors    int         `xml:"errors,attr"`
	Skipped   int         `xml:"skipped,attr"`
	Time      string      `xml:"time,attr"`
	TestCases []JUnitCase `xml:"testcase"`
}

type JUnitCase struct {
	XMLName   xml.Name      `xml:"testcase"`
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
	File      string        `xml:"file,attr,omitempty"`
	Line      string        `xml:"line,attr,omitempty"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
	Error     *JUnitError   `xml:"error,omitempty"`
	Skipped   *JUnitSkipped `xml:"skipped,omitempty"`
}

type JUnitFailure struct {
	XMLName xml.Name `xml:"failure"`
	Message string   `xml:"message,attr,omitempty"`
	Data    string   `xml:",cdata"`
}

type JUnitError struct {
	XMLName xml.Name `xml:"error"`
	Message string   `xml:"message,attr,omitempty"`
	Data    string   `xml:",cdata"`
}

type JUnitSkipped struct {
	XMLName xml.Name `xml:"skipped"`
	Message string   `xml:"message,attr,omitempty"`
}

// GenerateReport generates the Junit XML content as a string
func GenerateReport(reportName string, results types.Results) (string, error) {
	report := JUnitReport{
		Name:   reportName,
		Suites: []JUnitTestSuite{},
	}
	totalDuration := time.Duration(0)

	for suiteName, suiteResults := range results {
		suite := JUnitTestSuite{
			Name:      suiteName,
			Tests:     len(suiteResults),
			TestCases: []JUnitCase{},
		}
		suiteDuration := time.Duration(0)

		for _, result := range suiteResults {
			durationSeconds := fmt.Sprintf("%.3f", result.Duration.Seconds())
			totalDuration += result.Duration
			suiteDuration += result.Duration

			testCase := JUnitCase{
				Name:      result.Spec.Name,
				Classname: suiteName,
				Time:      durationSeconds,
			}

			if result.Spec.Pos != "" {
				parts := strings.SplitN(result.Spec.Pos, ":", 2)
				testCase.File = parts[0]
				if len(parts) > 1 {
					testCase.Line = parts[1]
				}
			}

			switch result.Status {
			case types.StatusFailure:
				suite.Failures++
				report.Failures++
				var failureContent string
				if result.ErrorMessage != "" {
					failureContent = result.ErrorMessage
				} else {
					var err error
					failureContent, err = util.ComputeDiff(result.Expected, result.Actual)
					if err != nil {
						return "", fmt.Errorf("failed to compute diff")
					}
				}
				testCase.Failure = &JUnitFailure{Message: "Test failed", Data: failureContent}
			case types.StatusError:
				suite.Errors++
				report.Errors++
				testCase.Error = &JUnitError{Message: "Test errored", Data: result.ErrorMessage}
			case types.StatusSkipped:
				suite.Skipped++
				report.Skipped++
				testCase.Skipped = &JUnitSkipped{Message: "Test skipped"}
			}
			report.Tests++
			suite.TestCases = append(suite.TestCases, testCase)
		}
		suite.Time = fmt.Sprintf("%.3f", suiteDuration.Seconds())
		report.Suites = append(report.Suites, suite)
	}

	report.Time = fmt.Sprintf("%.3f", totalDuration.Seconds())

	output, err := xml.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal XML: %w", err)
	}
	return xml.Header + string(output), nil
}

// WriteFile generates a Junit report and writes it to the specified path
func WriteFile(filePath string, reportName string, results types.Results) error {
	xmlContent, err := GenerateReport(reportName, results)
	if err != nil {
		return fmt.Errorf("failed to generate junit report content: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create junit file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = file.WriteString(xmlContent)
	if err != nil {
		return fmt.Errorf("failed to write junit report to %s: %w", filePath, err)
	}
	return nil
}
