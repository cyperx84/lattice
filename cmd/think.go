package cmd

import (
	"fmt"
	"strings"

	"github.com/cyperx84/lattice/internal/config"
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

	n := cfg.DefaultModels
	var specific []string
	if thinkModels != "" {
		specific = strings.Split(thinkModels, ",")
		for i := range specific {
			specific[i] = strings.TrimSpace(specific[i])
		}
	}

	result, err := think.Think(problem, idx, modelFiles, n, specific, effectiveLLM, verbose)
	if err != nil {
		return err
	}

	if jsonOutput {
		out, err := think.FormatJSON(result)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		fmt.Print(think.FormatResult(result))
	}

	return nil
}
