package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cyperx84/lattice/internal/index"
	"github.com/spf13/cobra"
)

var listCategory string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all models or filter by category",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listCategory, "category", "", "Filter by category name")
}

func runList(cmd *cobra.Command, args []string) error {
	idx, _, err := loadEmbeddedData()
	if err != nil {
		return err
	}

	var models []index.ModelEntry
	if listCategory != "" {
		models = idx.FilterByCategory(listCategory)
		if len(models) == 0 {
			fmt.Printf("No models found in category: %s\n\nAvailable categories:\n", listCategory)
			for cat := range idx.Categories {
				fmt.Printf("  - %s\n", cat)
			}
			return nil
		}
	} else {
		models = idx.Models
	}

	if jsonOutput {
		data, err := json.MarshalIndent(models, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	if listCategory != "" {
		fmt.Printf("%s (%d models):\n\n", listCategory, len(models))
	} else {
		fmt.Printf("All models (%d total):\n\n", len(models))
	}

	currentCat := ""
	for _, m := range models {
		if m.Category != currentCat {
			currentCat = m.Category
			if listCategory == "" {
				fmt.Printf("\n  %s\n", currentCat)
			}
		}
		fmt.Printf("    %-4s %-40s %s\n", m.ID, m.Name, m.Slug)
	}

	return nil
}
