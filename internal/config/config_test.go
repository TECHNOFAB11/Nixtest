package config

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	originalArgs := os.Args
	oldFlagSet := pflag.CommandLine
	defer func() {
		os.Args = originalArgs
		pflag.CommandLine = oldFlagSet
	}()

	// for Load() to not call log.Fatal(), a tests file must be provided
	os.Args = []string{"cmd", "-f", "dummy.json"}
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError) // reset flags

	cfg := Load()

	if cfg.NumWorkers != 4 {
		t.Errorf("Default NumWorkers: got %d, want 4", cfg.NumWorkers)
	}
	if cfg.SnapshotDir != "./snapshots" {
		t.Errorf("Default SnapshotDir: got %s, want ./snapshots", cfg.SnapshotDir)
	}
}

func TestLoad_Fatal(t *testing.T) {
	originalArgs := os.Args
	oldFlagSet := pflag.CommandLine
	defer func() {
		os.Args = originalArgs
		pflag.CommandLine = oldFlagSet
	}()

	// Load() should panic without tests file
	os.Args = []string{
		"cmd",
	}
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError) // Reset flags

	assert.Panics(t, func() { _ = Load() }, "Load should panic withot tests file")
}

func TestLoad_CustomValues(t *testing.T) {
	originalArgs := os.Args
	oldFlagSet := pflag.CommandLine
	defer func() {
		os.Args = originalArgs
		pflag.CommandLine = oldFlagSet
	}()

	// simulate cli args for Load()
	os.Args = []string{
		"cmd", // dummy command name
		"-f", "mytests.json",
		"--workers", "8",
		"--snapshot-dir", "/tmp/snaps",
		"--junit", "report.xml",
		"-u",
		"--skip", "specific-test",
		"--impure",
		"--no-color",
	}
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError) // Reset flags

	cfg := Load()

	if cfg.TestsFile != "mytests.json" {
		t.Errorf("TestsFile: got %s, want mytests.json", cfg.TestsFile)
	}
	if cfg.NumWorkers != 8 {
		t.Errorf("NumWorkers: got %d, want 8", cfg.NumWorkers)
	}
	if !cfg.UpdateSnapshots {
		t.Errorf("UpdateSnapshots: got %v, want true", cfg.UpdateSnapshots)
	}
	if cfg.SkipPattern != "specific-test" {
		t.Errorf("SkipPattern: got %s, want specific-test", cfg.SkipPattern)
	}
	if !cfg.ImpureEnv {
		t.Errorf("ImpureEnv: got %v, want true", cfg.ImpureEnv)
	}
}
