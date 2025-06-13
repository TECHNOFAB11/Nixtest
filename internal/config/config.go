package config

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
)

type AppConfig struct {
	NumWorkers      int
	TestsFile       string
	SnapshotDir     string
	JunitPath       string
	UpdateSnapshots bool
	SkipPattern     string
	PureEnv         bool
	NoColor         bool
}

// loads configuration from cli flags
func Load() AppConfig {
	cfg := AppConfig{}
	flag.IntVarP(&cfg.NumWorkers, "workers", "w", 4, "Amount of tests to run in parallel")
	flag.StringVarP(&cfg.TestsFile, "tests", "f", "", "Path to JSON file containing tests (required)")
	flag.StringVar(&cfg.SnapshotDir, "snapshot-dir", "./snapshots", "Directory where snapshots are stored")
	flag.StringVar(&cfg.JunitPath, "junit", "", "Path to generate JUNIT report to, leave empty to disable")
	flag.BoolVarP(&cfg.UpdateSnapshots, "update-snapshots", "u", false, "Update all snapshots")
	flag.StringVarP(&cfg.SkipPattern, "skip", "s", "", "Regular expression to skip tests (e.g., 'test-.*|.*-b')")
	flag.BoolVar(&cfg.PureEnv, "pure", false, "Unset all env vars before running script tests")
	flag.BoolVar(&cfg.NoColor, "no-color", false, "Disable coloring")
	helpRequested := flag.BoolP("help", "h", false, "Show this menu")

	flag.Parse()

	if *helpRequested {
		fmt.Println("Usage of nixtest:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if cfg.TestsFile == "" {
		log.Panic().Msg("Tests file path (-f or --tests) is required.")
	}

	if cfg.NoColor {
		text.DisableColors()
	}

	return cfg
}
