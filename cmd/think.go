package cmd

import (
	"fmt"
	"strings"

	"github.com/cyperx84/lattice/internal/apply"
	"github.com/cyperx84/lattice/internal/color"
	"github.com/cyperx84/lattice/internal/config"
	"github.com/cyperx84/lattice/internal/history"
	"github.com/cyperx84/lattice/internal/think"
	"github.com/spf13/cobra"
)

var thinkModels string

var thinkCmd = &cobra.Command{
	Use:   "think <problem>",
	Short: "Surface top models and apply thinking steps to a problem",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runThinkCmd,
}

func init() {
	rootCmd.AddCommand(thinkCmd)
	thinkCmd.Flags().StringVar(&thinkModels, "models", "", "Comma-separated model slugs (e.g., inversion,second-order-thinking)")
}

func runThinkCmd(cmd *cobra.Command, args []string) error {
	setupColor()

	problem := strings.Join(args, " ")
	cfg := config.Load()

	idx, modelFiles, err := loadAllData()
	if err != nil {
		return err
	}

	effectiveLLM := llmCmd
	if effectiveLLM == "" {
		effectiveLLM = cfg.LLMCmd
	}
	if noLLM {
		effectiveLLM = ""
	}

	n := cfg.DefaultModels
	var specific []string
	if thinkModels != "" {
		specific = strings.Split(thinkModels, ",")
		for i := range specific {
			specific[i] = strings.TrimSpace(specific[i])
		}
	}

	result, err := think.Think(problem, idx, modelFiles, n, specific, effectiveLLM, verbose, timeout)
	if err != nil {
		if strings.Contains(err.Error(), "no relevant models") {
			fmt.Fprintf(cmd.ErrOrStderr(), "Hint: try 'lattice suggest \"%s\"' for broader recommendations\n", problem)
		}
		return err
	}

	// Save to history (unless --no-history)
	if !noHistory {
		mgr, histErr := history.NewManager()
		if histErr == nil {
			var modelSlugs []string
			for _, m := range result.Models {
				modelSlugs = append(modelSlugs, m.ModelSlug)
			}
			entry := &history.Entry{
				Type:    "think",
				Problem: problem,
				Models:  modelSlugs,
				Summary: result.Summary,
			}
			_ = mgr.Save(entry) // Ignore errors silently
		}
	}

	if jsonOutput {
		out, err := think.FormatJSON(result)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		fmt.Print(formatThinkResult(result))
	}

	return nil
}

func formatThinkResult(r *think.Result) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s %s\n\n", color.Bold("# Thinking:"), r.Problem))
	b.WriteString(fmt.Sprintf("Models applied: %s\n\n", color.Bold(fmt.Sprintf("%d", len(r.Models)))))

	for _, m := range r.Models {
		b.WriteString(formatApplyResultColored(&m))
		b.WriteString("\n---\n\n")
	}

	if r.Summary != "" {
		b.WriteString(fmt.Sprintf("%s\n\n", color.Bold("## Synthesis")))
		b.WriteString(r.Summary)
		b.WriteString("\n")
	}

	return b.String()
}

func formatApplyResultColored(r *apply.Result) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s\n", color.BoldCyan(fmt.Sprintf("## %s", r.ModelName))))
	b.WriteString(fmt.Sprintf("Category: %s\n\n", r.Category))

	if r.Synthesis != "" {
		b.WriteString(r.Synthesis)
		b.WriteString("\n")
	} else {
		b.WriteString(fmt.Sprintf("%s\n", color.Bold("### Thinking Steps")))
		for i, step := range r.Steps {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
		b.WriteString(fmt.Sprintf("\n%s\n", color.Bold("### Coaching Questions")))
		for _, q := range r.Questions {
			b.WriteString(fmt.Sprintf("- %s\n", q))
		}
	}

	return b.String()
}
