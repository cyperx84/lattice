package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cyperx84/lattice/internal/color"
	"github.com/cyperx84/lattice/internal/journal"
	"github.com/spf13/cobra"
)

var (
	journalDue     bool
	journalAll     bool
	journalProject bool
	journalLimit   int
)

var journalCmd = &cobra.Command{
	Use:   "journal",
	Short: "View and manage your decision journal",
	RunE:  runJournalList,
}

var journalReviewCmd = &cobra.Command{
	Use:   "review <id>",
	Short: "Review a past decision and record the outcome",
	Args:  cobra.ExactArgs(1),
	RunE:  runJournalReview,
}

func init() {
	rootCmd.AddCommand(journalCmd)
	journalCmd.AddCommand(journalReviewCmd)

	journalCmd.Flags().BoolVar(&journalDue, "due", false, "Show only decisions due for review")
	journalCmd.Flags().BoolVar(&journalAll, "all", false, "Show all decisions")
	journalCmd.Flags().BoolVar(&journalProject, "project", false, "Read from ./decisions/ instead of global journal")
	journalCmd.Flags().IntVar(&journalLimit, "limit", 20, "Number of entries to show")
}

func runJournalList(cmd *cobra.Command, args []string) error {
	setupColor()

	dir := journal.JournalDir()
	if journalProject {
		dir = journal.ProjectJournalDir()
	}

	limit := journalLimit
	if journalAll || journalDue {
		limit = 0
	}

	entries, err := journal.List(dir, limit)
	if err != nil {
		return err
	}

	if journalDue {
		entries = journal.DueForReview(entries)
	}

	if jsonOutput {
		color.Disable()
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	if len(entries) == 0 {
		if journalDue {
			fmt.Println("No decisions due for review.")
		} else {
			fmt.Println("No decisions recorded yet.")
			fmt.Println("Run 'lattice decide \"<decision>\"' to create your first entry.")
		}
		return nil
	}

	if journalDue {
		fmt.Printf("%s\n\n", color.Bold("Decisions due for review:"))
	} else {
		fmt.Printf("%s (%s):\n\n", color.Bold("Decision journal"), color.Bold(fmt.Sprintf("%d", len(entries))))
	}

	for i, e := range entries {
		status := color.Green("open")
		if e.Status == "resolved" {
			status = color.Dim("resolved")
		}

		fmt.Printf("%s. %s [%s] %s\n",
			color.Bold(fmt.Sprintf("%d", i+1)),
			color.Dim(e.Date),
			status,
			e.Decision,
		)
		if len(e.Models) > 0 {
			fmt.Printf("   %s Models: %s\n", color.Arrow(), color.Dim(strings.Join(e.Models, ", ")))
		}
		if e.Prediction != "" && e.Prediction != "_No prediction made_" {
			pred := e.Prediction
			if len(pred) > 60 {
				pred = pred[:57] + "..."
			}
			fmt.Printf("   %s Prediction: %s\n", color.Arrow(), color.Dim(pred))
		}
		if e.Status == "open" && len(e.ReviewDates) > 0 {
			fmt.Printf("   %s Review: %s\n", color.Arrow(), color.Dim(strings.Join(e.ReviewDates, ", ")))
		}
		fmt.Println()
	}

	return nil
}

func runJournalReview(cmd *cobra.Command, args []string) error {
	setupColor()

	id := args[0]

	dir := journal.JournalDir()
	if journalProject {
		dir = journal.ProjectJournalDir()
	}

	// Find the entry file
	path := findEntryPath(dir, id)
	if path == "" {
		return fmt.Errorf("decision not found: %s", id)
	}

	entry, err := journal.LoadEntry(path)
	if err != nil {
		return err
	}

	// Show the original decision
	fmt.Printf("\n%s %s\n", color.Bold("Decision:"), entry.Decision)
	fmt.Printf("%s %s\n", color.Bold("Date:"), entry.Date)
	fmt.Printf("%s %s\n", color.Bold("Models:"), strings.Join(entry.Models, ", "))
	fmt.Printf("%s %s\n\n", color.Bold("Prediction:"), entry.Prediction)

	reader := bufio.NewReader(os.Stdin)

	// Ask what happened
	fmt.Printf("%s ", color.Bold("What actually happened?"))
	outcome, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read outcome: %w", err)
	}
	outcome = strings.TrimSpace(outcome)

	// Ask if it was a good decision
	fmt.Printf("%s ", color.Bold("Was this a good decision? [yes/no/mixed/too-early]"))
	verdict, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read verdict: %w", err)
	}
	verdict = strings.TrimSpace(verdict)

	// Update entry
	entry.Status = "resolved"
	entry.Outcome = fmt.Sprintf("%s (Verdict: %s)", outcome, verdict)

	// Rewrite the file
	if _, err := journal.Save(entry, journalProject); err != nil {
		return err
	}

	fmt.Printf("\n%s Decision %s marked as resolved\n", color.Checkmark(), color.Bold(id))
	return nil
}

func findEntryPath(dir, id string) string {
	// Try direct match: id.md
	path := dir + "/" + id + ".md"
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// Scan directory for matching ID in frontmatter
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		p := dir + "/" + e.Name()
		entry, err := journal.LoadEntry(p)
		if err != nil {
			continue
		}
		if entry.ID == id {
			return p
		}
	}
	return ""
}
