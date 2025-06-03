package types

import "testing"

func TestTestStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status TestStatus
		want   string
	}{
		{"Success", StatusSuccess, "SUCCESS"},
		{"Failure", StatusFailure, "FAILURE"},
		{"Error", StatusError, "ERROR"},
		{"Skipped", StatusSkipped, "SKIPPED"},
		{"Unknown", TestStatus(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("TestStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
