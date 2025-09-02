package main

import (
	"os"

	"gitlab.com/technofab/nixtest/internal/config"
	appnix "gitlab.com/technofab/nixtest/internal/nix"
	"gitlab.com/technofab/nixtest/internal/report/console"
	"gitlab.com/technofab/nixtest/internal/report/junit"
	"gitlab.com/technofab/nixtest/internal/runner"
	appsnap "gitlab.com/technofab/nixtest/internal/snapshot"
	"gitlab.com/technofab/nixtest/internal/types"
	"gitlab.com/technofab/nixtest/internal/util"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal().Any("r", r).Msg("Panicked")
		}
	}()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFieldFormat})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	appCfg := config.Load()

	log.Info().
		Int("workers", appCfg.NumWorkers).
		Str("testsFile", appCfg.TestsFile).
		Msg("Starting nixtest")

	if _, err := os.Stat(appCfg.TestsFile); os.IsNotExist(err) {
		log.Error().Str("file", appCfg.TestsFile).Msg("Tests file does not exist")
		os.Exit(1)
	}

	suites, err := util.ParseFile[[]types.SuiteSpec](appCfg.TestsFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load tests from file")
		os.Exit(1)
	}

	totalTests := 0
	for _, suite := range suites {
		totalTests += len(suite.Tests)
	}
	log.Info().
		Int("suites", len(suites)).
		Int("tests", totalTests).
		Msg("Discovered suites and tests")

	nixService := appnix.NewDefaultService()
	snapshotService := appsnap.NewDefaultService()

	runnerCfg := runner.Config{
		NumWorkers:      appCfg.NumWorkers,
		SnapshotDir:     appCfg.SnapshotDir,
		UpdateSnapshots: appCfg.UpdateSnapshots,
		SkipPattern:     appCfg.SkipPattern,
		ImpureEnv:       appCfg.ImpureEnv,
	}
	testRunner, err := runner.New(runnerCfg, nixService, snapshotService)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize test runner")
	}

	results := testRunner.RunTests(suites)

	relevantSuccessCount := 0
	for _, suiteResults := range results {
		for _, r := range suiteResults {
			if r.Status == types.StatusSuccess || r.Status == types.StatusSkipped {
				relevantSuccessCount++
			}
		}
	}

	if appCfg.JunitPath != "" {
		err = junit.WriteFile(appCfg.JunitPath, "nixtest", results)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate junit file")
		} else {
			log.Info().Str("path", appCfg.JunitPath).Msg("Generated Junit report")
		}
	}

	// print errors first then summary
	console.PrintErrors(results, appCfg.NoColor)
	console.PrintSummary(results, relevantSuccessCount, totalTests)

	if relevantSuccessCount != totalTests {
		log.Error().Msgf("Test run finished with failures or errors. %d/%d successful (includes skipped).", relevantSuccessCount, totalTests)
		os.Exit(2) // exit 2 on test failures, 1 is for internal errors
	}

	log.Info().Msg("All tests passed successfully!")
}
