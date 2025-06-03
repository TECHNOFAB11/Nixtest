package console

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/rs/zerolog/log"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gitlab.com/technofab/nixtest/internal/types"
	"gitlab.com/technofab/nixtest/internal/util"
)

// PrintErrors prints error messages for failed tests
func PrintErrors(results types.Results, noColor bool) {
	for _, suiteResults := range results {
		for _, result := range suiteResults {
			if result.Status == types.StatusSuccess || result.Status == types.StatusSkipped {
				continue
			}
			fmt.Println(text.FgRed.Sprintf("⚠ Test \"%s/%s\" failed:", result.Spec.Suite, result.Spec.Name))
			message := result.ErrorMessage
			if result.Status == types.StatusFailure && message == "" {
				if noColor {
					var err error
					message, err = util.ComputeDiff(result.Expected, result.Actual)
					if err != nil {
						log.Panic().Err(err).Msg("failed to compute diff")
					}
				} else {
					dmp := diffmatchpatch.New()
					diffs := dmp.DiffMain(result.Expected, result.Actual, true)
					message = fmt.Sprintf("Diff:\n%s", dmp.DiffPrettyText(diffs))
				}
			}

			if message == "" {
				message = "- no output -"
			}

			for _, line := range strings.Split(strings.TrimRight(message, "\n"), "\n") {
				fmt.Printf("%s %s\n", text.FgRed.Sprint("|"), line)
			}
			fmt.Println()
		}
	}
}

// PrintSummary prints a table summarizing test results
func PrintSummary(results types.Results, totalSuccessCount int, totalTestCount int) {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Suite / Test", "Duration", "Status", "File:Line"})

	log.Info().Msg("Summary:")

	suiteNames := make([]string, 0, len(results))
	for name := range results {
		suiteNames = append(suiteNames, name)
	}
	sort.Strings(suiteNames)

	for _, suiteName := range suiteNames {
		suiteResults := results[suiteName]
		suiteTotal := len(suiteResults)
		suiteSuccess := 0
		suiteSkipped := 0

		for _, res := range suiteResults {
			if res.Status == types.StatusSuccess {
				suiteSuccess++
			} else if res.Status == types.StatusSkipped {
				suiteSkipped++
			}
		}

		statusStr := fmt.Sprintf("%d/%d", suiteSuccess, suiteTotal)
		if suiteSkipped > 0 {
			statusStr += fmt.Sprintf(" (%d skipped)", suiteSkipped)
		}

		t.AppendRow(table.Row{
			text.Bold.Sprint(suiteName),
			"",
			statusStr,
			"",
		})

		sort.Slice(suiteResults, func(i, j int) bool {
			return suiteResults[i].Spec.Name < suiteResults[j].Spec.Name
		})

		for _, res := range suiteResults {
			var symbol string
			switch res.Status {
			case types.StatusSuccess:
				symbol = text.FgGreen.Sprint("✅ PASS")
			case types.StatusFailure:
				symbol = text.FgRed.Sprint("❌ FAIL")
			case types.StatusError:
				symbol = text.FgYellow.Sprint("❗ ERROR")
			case types.StatusSkipped:
				symbol = text.FgBlue.Sprint("⏭️ SKIP")
			default:
				symbol = "UNKNOWN"
			}

			t.AppendRow([]any{
				"  " + res.Spec.Name,
				fmt.Sprintf("%s", res.Duration.Round(time.Millisecond)),
				symbol,
				res.Spec.Pos,
			})
		}
		t.AppendSeparator()
	}

	overallStatusStr := fmt.Sprintf("%d/%d", totalSuccessCount, totalTestCount)
	totalSkipped := 0
	for _, suiteResults := range results {
		for _, res := range suiteResults {
			if res.Status == types.StatusSkipped {
				totalSkipped++
			}
		}
	}
	if totalSkipped > 0 {
		overallStatusStr += fmt.Sprintf(" (%d skipped)", totalSkipped)
	}

	t.AppendFooter(table.Row{
		text.Bold.Sprint("TOTAL"),
		"",
		text.Bold.Sprint(overallStatusStr),
		"",
	})
	t.Render()
}
