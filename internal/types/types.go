package types

import "time"

type TestType string

const (
	TestTypeScript   TestType = "script"
	TestTypeUnit     TestType = "unit"
	TestTypeSnapshot TestType = "snapshot"
)

type SuiteSpec struct {
	Name  string     `json:"name"`
	Tests []TestSpec `json:"tests"`
}

type TestSpec struct {
	Type        TestType `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Expected    any      `json:"expected,omitempty"`
	Actual      any      `json:"actual,omitempty"`
	ActualDrv   string   `json:"actualDrv,omitempty"`
	Script      string   `json:"script,omitempty"`
	Pos         string   `json:"pos,omitempty"`

	Suite string
}

type TestStatus int

const (
	StatusSuccess TestStatus = iota
	StatusFailure
	StatusError
	StatusSkipped
)

func (ts TestStatus) String() string {
	switch ts {
	case StatusSuccess:
		return "SUCCESS"
	case StatusFailure:
		return "FAILURE"
	case StatusError:
		return "ERROR"
	case StatusSkipped:
		return "SKIPPED"
	default:
		return "UNKNOWN"
	}
}

type TestResult struct {
	Spec         TestSpec
	Status       TestStatus
	Duration     time.Duration
	ErrorMessage string
	Expected     string
	Actual       string
}

type Results map[string][]TestResult
