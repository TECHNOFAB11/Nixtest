package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/rs/zerolog/log"
)

func printErrors(results Results) {
	for _, suiteResults := range results {
		for _, result := range suiteResults {
			if result.Success {
				continue
			}
			fmt.Println(text.FgRed.Sprintf("⚠ Test \"%s\" failed:", result.Name))
			var message string
			if result.Error.Diff != "" {
				message = fmt.Sprintf("Diff:\n%s", result.Error.Diff)
			} else {
				message = result.Error.Message
			}
			for line := range strings.Lines(message) {
				fmt.Printf("%s %s", text.FgRed.Sprint("|"), line)
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

	for suite, suiteResults := range results {
		suiteTotal := len(suiteResults)
		suiteSuccess := 0

		for _, res := range suiteResults {
			if res.Success {
				suiteSuccess++
			}
		}

		t.AppendRow(table.Row{
			text.Bold.Sprint(suite),
			"",
			fmt.Sprintf("%d/%d", suiteSuccess, suiteTotal),
			"",
		})
		for _, res := range suiteResults {
			symbol := "❌"
			if res.Success {
				symbol = "✅"
			}

			t.AppendRow([]any{
				res.Name,
				fmt.Sprintf("%s", res.Duration),
				symbol,
				res.Pos,
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
