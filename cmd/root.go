package cmd

import (
	"fmt"
	"os"

	"github.com/cyperx84/lattice/internal/color"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	llmCmd     string
	verbose    bool
	noLLM      bool
	noHistory  bool
	timeout    int
)

var version = "0.4.0"

var rootCmd = &cobra.Command{
	Use:   "lattice",
	Short: "Mental models engine — apply Munger's latticework to any problem",
	Long: `lattice surfaces and applies mental models from Charlie Munger's
latticework of 98 cognitive frameworks. Think through problems,
search for relevant models, and apply structured thinking steps.`,
	Version: version,
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
	rootCmd.PersistentFlags().BoolVar(&noLLM, "no-llm", false, "Skip LLM calls entirely")
	rootCmd.PersistentFlags().BoolVar(&noHistory, "no-history", false, "Skip saving to history")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 60, "LLM timeout in seconds")
}

// setupColor configures color output based on flags and environment
func setupColor() {
	if jsonOutput {
		color.Disable()
	}
}
