package junit

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gitlab.com/technofab/nixtest/internal/types"
)

func formatDurationSeconds(d time.Duration) string {
	return fmt.Sprintf("%.3f", d.Seconds())
}

func TestGenerateReport(t *testing.T) {
	results := types.Results{
		"Suite1": []types.TestResult{
			{
				Spec:     types.TestSpec{Name: "Test1_Success", Suite: "Suite1", Pos: "file1.nix:10"},
				Status:   types.StatusSuccess,
				Duration: 123 * time.Millisecond,
			},
			{
				Spec:     types.TestSpec{Name: "Test2_Failure", Suite: "Suite1", Pos: "file1.nix:20"},
				Status:   types.StatusFailure,
				Duration: 234 * time.Millisecond,
				Expected: "hello",
				Actual:   "world",
			},
		},
		"Suite2": []types.TestResult{
			{
				Spec:         types.TestSpec{Name: "Test3_Error", Suite: "Suite2"},
				Status:       types.StatusError,
				Duration:     345 * time.Millisecond,
				ErrorMessage: "Something went very wrong",
			},
			{
				Spec:         types.TestSpec{Name: "Test4_Failure_Message", Suite: "Suite2"},
				Status:       types.StatusFailure,
				Duration:     456 * time.Millisecond,
				ErrorMessage: "hello world",
			},
			{
				Spec:     types.TestSpec{Name: "Test5_Skipped", Suite: "Suite2"},
				Status:   types.StatusSkipped,
				Duration: 567 * time.Millisecond,
			},
		},
	}

	totalDuration := (123 + 234 + 345 + 456 + 567) * time.Millisecond
	reportName := "MyNixtestReport"
	xmlString, err := GenerateReport(reportName, results)
	if err != nil {
		t.Fatalf("GenerateReport() failed: %v", err)
	}

	if !strings.HasPrefix(xmlString, xml.Header) {
		t.Error("GenerateReport() output missing XML header")
	}
	if !strings.Contains(xmlString, "<testsuites name=\"MyNixtestReport\"") {
		t.Errorf("GenerateReport() missing root <testsuites>. Got: %s", xmlString)
	}
	if !strings.Contains(xmlString, "tests=\"5\"") {
		t.Errorf("GenerateReport() incorrect total tests count. Got: %s", xmlString)
	}
	if !strings.Contains(xmlString, "failures=\"2\"") {
		t.Errorf("GenerateReport() incorrect total failures count. Got: %s", xmlString)
	}
	if !strings.Contains(xmlString, "errors=\"1\"") {
		t.Errorf("GenerateReport() incorrect total errors count. Got: %s", xmlString)
	}
	if !strings.Contains(xmlString, "time=\""+formatDurationSeconds(totalDuration)+"\"") {
		t.Errorf("GenerateReport() incorrect total time. Expected %s. Got part: %s", formatDurationSeconds(totalDuration), xmlString)
	}

	var report JUnitReport
	if err := xml.Unmarshal([]byte(strings.TrimPrefix(xmlString, xml.Header)), &report); err != nil {
		t.Fatalf("Failed to unmarshal generated XML: %v\nXML:\n%s", err, xmlString)
	}

	if report.Name != reportName {
		t.Errorf("Report.Name = %q, want %q", report.Name, reportName)
	}
	if len(report.Suites) != 2 {
		t.Fatalf("Report.Suites length = %d, want 2", len(report.Suites))
	}
}

func TestWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "junit_report.xml")
	results := types.Results{
		"SuiteSimple": []types.TestResult{
			{Spec: types.TestSpec{Name: "SimpleTest", Suite: "SuiteSimple"}, Status: types.StatusSuccess, Duration: 1 * time.Second},
		},
	}

	err := WriteFile(filePath, "TestReport", results)
	if err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read written JUnit file: %v", err)
	}
	if !strings.Contains(string(content), "<testsuites name=\"TestReport\"") {
		t.Error("Written JUnit file content seems incorrect.")
	}
}
