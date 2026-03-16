package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

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
	query := strings.Join(args, " ")

	idx, _, err := loadEmbeddedData()
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
		return nil
	}

	fmt.Printf("Found %d model(s) matching \"%s\":\n\n", len(results), query)
	for _, m := range results {
		fmt.Printf("  %-30s [%s] %s\n", m.Name, m.Category, m.Slug)
		if m.Summary != "" {
			fmt.Printf("  %s\n\n", m.Summary)
		}
	}

	return nil
}
