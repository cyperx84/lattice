package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	llmCmd     string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "lattice",
	Short: "Mental models engine — apply Munger's latticework to any problem",
	Long: `lattice surfaces and applies mental models from Charlie Munger's
latticework of 98 cognitive frameworks. Think through problems,
search for relevant models, and apply structured thinking steps.`,
	Version: "0.1.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().StringVar(&llmCmd, "llm-cmd", "", "LLM command for synthesis (default: from config or 'claude -p')")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")
}
