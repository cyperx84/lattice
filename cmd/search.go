package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cyperx84/lattice/internal/color"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search model index by keyword",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	setupColor()

	query := strings.Join(args, " ")

	idx, _, err := loadAllData()
	if err != nil {
		return err
	}

	results := idx.Search(query)

	if jsonOutput {
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	if len(results) == 0 {
		fmt.Printf("No models found for: %s\n", query)
		fmt.Println("\nTips: try broader terms like 'decision', 'risk', 'strategy', 'team', or 'growth'")
		fmt.Println("Or use 'lattice suggest' for relevance-ranked recommendations")
		return nil
	}

	fmt.Printf("Found %s matching \"%s\":\n\n", color.Bold(fmt.Sprintf("%d model(s)", len(results))), query)
	for _, m := range results {
		fmt.Printf("  %s %s %s\n",
			color.Cyan(fmt.Sprintf("%-30s", m.Name)),
			color.Dim("["+m.Category+"]"),
			color.Dim(m.Slug))
		if m.Summary != "" {
			fmt.Printf("  %s\n\n", m.Summary)
		}
	}

	return nil
}
