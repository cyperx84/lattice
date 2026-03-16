package cmd

import (
	"fmt"
	"strings"

	"github.com/cyperx84/lattice/internal/color"
	"github.com/cyperx84/lattice/internal/history"
	"github.com/spf13/cobra"
)

var historyLimit int

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View or manage think/apply session history",
	RunE:  runHistoryList,
}

var historyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent session history (default subcommand)",
	RunE:  runHistoryList,
}

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Delete all history entries",
	RunE:  runHistoryClear,
}

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.AddCommand(historyListCmd)
	historyCmd.AddCommand(historyClearCmd)

	historyCmd.Flags().IntVar(&historyLimit, "limit", 20, "Number of entries to show")
	historyListCmd.Flags().IntVar(&historyLimit, "limit", 20, "Number of entries to show")
}

func runHistoryList(cmd *cobra.Command, args []string) error {
	mgr, err := history.NewManager()
	if err != nil {
		return err
	}

	entries, err := mgr.List(historyLimit)
	if err != nil {
		return err
	}

	if jsonOutput {
		// Disable color for JSON output
		color.Disable()
		out, err := history.FormatJSON(entries)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	}

	// Apply colors to output
	fmt.Print(formatHistoryList(entries))
	return nil
}

func runHistoryClear(cmd *cobra.Command, args []string) error {
	mgr, err := history.NewManager()
	if err != nil {
		return err
	}

	if err := mgr.Clear(); err != nil {
		return err
	}

	fmt.Printf("%s History cleared\n", color.Checkmark())
	return nil
}

func formatHistoryList(entries []*history.Entry) string {
	if len(entries) == 0 {
		return "No history entries found.\nRun 'lattice think' or 'lattice apply' to create entries.\n"
	}

	var output string
	output += fmt.Sprintf("Recent sessions (%s):\n\n", color.Bold(fmt.Sprintf("%d", len(entries))))

	for i, e := range entries {
		ts := e.Timestamp
		output += fmt.Sprintf("%s. %s %s\n",
			color.Bold(fmt.Sprintf("%d", i+1)),
			color.Dim("["+ts+"]"),
			color.Cyan(e.Type))
		output += fmt.Sprintf("   %s\n", e.Problem)
		if len(e.Models) > 0 {
			output += fmt.Sprintf("   %s %s\n", color.Arrow(), color.Dim(strings.Join(e.Models, ", ")))
		}
		output += "\n"
	}

	return output
}


