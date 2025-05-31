package main

import (
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/rs/zerolog/log"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func printErrors(results Results) {
	for _, suiteResults := range results {
		for _, result := range suiteResults {
			if result.Status == StatusSuccess || result.Status == StatusSkipped {
				continue
			}
			fmt.Println(text.FgRed.Sprintf("⚠ Test \"%s\" failed:", result.Spec.Name))
			var message string = result.ErrorMessage
			// if ErrorMessage is set, prefer that
			if result.Status == StatusFailure && message == "" {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(result.Expected, result.Actual, false)
				message = fmt.Sprintf("Diff:\n%s", dmp.DiffPrettyText(diffs))
			}

			endedWithNewline := false
			// handle multi-line colored changes
			colorState := ""
			colorRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`) // Match any escape sequence
			for line := range strings.Lines(message) {
				coloredLine := colorState + line
				endedWithNewline = strings.HasSuffix(coloredLine, "\n")
				fmt.Printf("%s %s", text.FgRed.Sprint("|"), coloredLine)

				matches := colorRegex.FindAllString(line, -1)

				// determine last color code, to copy to next line
				if len(matches) > 0 {
					lastMatch := matches[len(matches)-1]
					if lastMatch == "\x1b[0m" {
						colorState = "" // reset color state
					} else {
						colorState = lastMatch // save color state for next line
					}
				}
			}
			if endedWithNewline {
				fmt.Printf("%s", text.FgRed.Sprint("|"))
			}
			if message == "" {
				fmt.Printf("- no output -")
			}
			fmt.Printf("\n\n")
		}
	}
}

func printSummary(results Results, successCount int, totalCount int) {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Test", "Duration", "Pass", "File"})

	log.Info().Msg("Summary:")

	suitesSorted := slices.Sorted(maps.Keys(results))
	for _, suite := range suitesSorted {
		suiteResults := results[suite]
		suiteTotal := len(suiteResults)
		suiteSuccess := 0

		for _, res := range suiteResults {
			if res.Status == StatusSuccess || res.Status == StatusSkipped {
				suiteSuccess++
			}
		}

		t.AppendRow(table.Row{
			text.Bold.Sprint(suite),
			"",
			fmt.Sprintf("%d/%d", suiteSuccess, suiteTotal),
			"",
		})
		sort.Slice(suiteResults, func(i, j int) bool {
			return suiteResults[i].Spec.Name < suiteResults[j].Spec.Name
		})
		for _, res := range suiteResults {
			symbol := "❌"
			if res.Status == StatusSuccess {
				symbol = "✅"
			} else if res.Status == StatusError {
				symbol = "error"
			} else if res.Status == StatusSkipped {
				symbol = "skipped"
			}

			t.AppendRow([]any{
				res.Spec.Name,
				fmt.Sprintf("%s", res.Duration),
				symbol,
				res.Spec.Pos,
			})
		}
		t.AppendSeparator()
	}

	t.AppendFooter(table.Row{
		text.Bold.Sprint("TOTAL"),
		"",
		fmt.Sprintf("%d/%d", successCount, totalCount),
		"",
	})
	t.Render()
}
