package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cyperx84/lattice/internal/index"
	"github.com/spf13/cobra"
)

var suggestCount int

var suggestCmd = &cobra.Command{
	Use:   "suggest <situation>",
	Short: "Recommend which mental models to use for a situation",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSuggest,
}

func init() {
	rootCmd.AddCommand(suggestCmd)
	suggestCmd.Flags().IntVar(&suggestCount, "count", 5, "Number of suggestions (default 5)")
}

type suggestion struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Category string `json:"category"`
	Summary  string `json:"summary"`
	Why      string `json:"why"`
}

func runSuggest(cmd *cobra.Command, args []string) error {
	situation := strings.Join(args, " ")

	idx, _, err := loadAllData()
	if err != nil {
		return err
	}

	matches := idx.TopNForQuery(situation, suggestCount)

	if len(matches) == 0 {
		fmt.Printf("No relevant models found for: %s\n", situation)
		return nil
	}

	suggestions := make([]suggestion, len(matches))
	for i, m := range matches {
		suggestions[i] = suggestion{
			Name:     m.Name,
			Slug:     m.Slug,
			Category: m.Category,
			Summary:  m.Summary,
			Why:      buildWhy(m, situation),
		}
	}

	if jsonOutput {
		data, err := json.MarshalIndent(suggestions, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Suggested models for: \"%s\"\n\n", situation)
	for i, s := range suggestions {
		fmt.Printf("%d. %s (%s)\n", i+1, s.Name, s.Category)
		fmt.Printf("   %s\n", s.Summary)
		fmt.Printf("   → Why: %s\n\n", s.Why)
	}
	fmt.Printf("→ Run: lattice think \"%s\" to apply these models\n", situation)

	return nil
}

// buildWhy generates a relevance explanation from keyword/summary match context.
func buildWhy(m index.ModelEntry, situation string) string {
	sit := strings.ToLower(situation)
	words := strings.Fields(sit)

	// Check which keywords matched
	var matchedKW []string
	for _, kw := range m.Keywords {
		kwLower := strings.ToLower(kw)
		for _, w := range words {
			if strings.Contains(kwLower, w) || strings.Contains(w, kwLower) {
				matchedKW = append(matchedKW, kw)
				break
			}
		}
	}

	if len(matchedKW) > 0 {
		return fmt.Sprintf("Relevant to: %s", strings.Join(matchedKW, ", "))
	}

	// Check name match
	nameLower := strings.ToLower(m.Name)
	for _, w := range words {
		if strings.Contains(nameLower, w) {
			return fmt.Sprintf("Model directly addresses %s", w)
		}
	}

	// Fallback: extract relevant fragment from summary
	summaryLower := strings.ToLower(m.Summary)
	for _, w := range words {
		if pos := strings.Index(summaryLower, w); pos >= 0 {
			// Extract a window around the match
			start := pos
			if start > 20 {
				start = pos - 20
			}
			end := pos + len(w) + 60
			if end > len(m.Summary) {
				end = len(m.Summary)
			}
			fragment := m.Summary[start:end]
			if start > 0 {
				fragment = "..." + fragment
			}
			if end < len(m.Summary) {
				fragment = fragment + "..."
			}
			return fragment
		}
	}

	return "Broadly applicable to this type of situation"
}
