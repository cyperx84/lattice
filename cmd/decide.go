package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cyperx84/lattice/internal/color"
	"github.com/cyperx84/lattice/internal/config"
	"github.com/cyperx84/lattice/internal/index"
	"github.com/cyperx84/lattice/internal/journal"
	"github.com/cyperx84/lattice/internal/modelfile"
	"github.com/cyperx84/lattice/internal/think"
	"github.com/spf13/cobra"
)

var (
	decideQuick      bool
	decideProject    bool
	decideModels     string
	decidePrediction string
)

var decideCmd = &cobra.Command{
	Use:   "decide <decision>",
	Short: "Record a decision with mental model analysis and prediction",
	Long: `Capture a decision, apply relevant mental models, record your prediction,
and schedule future reviews. Creates a markdown journal entry.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDecideCmd,
}

func init() {
	rootCmd.AddCommand(decideCmd)
	decideCmd.Flags().BoolVar(&decideQuick, "quick", false, "Skip thinking steps, just capture prediction")
	decideCmd.Flags().BoolVar(&decideProject, "project", false, "Save to ./decisions/ instead of global journal")
	decideCmd.Flags().StringVar(&decideModels, "models", "", "Force specific models (comma-separated slugs)")
	decideCmd.Flags().StringVar(&decidePrediction, "prediction", "", "Pass prediction inline (skip prompt)")
}

func runDecideCmd(cmd *cobra.Command, args []string) error {
	setupColor()

	decision := strings.Join(args, " ")
	cfg := config.Load()

	idx, modelFiles, err := loadAllData()
	if err != nil {
		return err
	}

	// Resolve models
	var modelSlugs []string
	var modelNames []string

	if decideModels != "" {
		// User-specified models
		for _, slug := range strings.Split(decideModels, ",") {
			slug = strings.TrimSpace(slug)
			entry := idx.FindBySlug(slug)
			if entry == nil {
				return fmt.Errorf("model not found: %s", slug)
			}
			modelSlugs = append(modelSlugs, entry.Slug)
			modelNames = append(modelNames, fmt.Sprintf("**%s** (%s)", entry.Name, entry.Category))
		}
	} else {
		// Auto-select top 3
		top := idx.TopNForQuery(decision, 3)
		if len(top) == 0 {
			return fmt.Errorf("no relevant models found for: %s", decision)
		}
		for _, entry := range top {
			modelSlugs = append(modelSlugs, entry.Slug)
			modelNames = append(modelNames, fmt.Sprintf("**%s** (%s)", entry.Name, entry.Category))
		}
	}

	// Print decision and models
	fmt.Printf("\n%s %s\n\n", color.Bold("Decision:"), decision)
	fmt.Printf("%s\n", color.Bold("Models applied:"))
	for _, name := range modelNames {
		fmt.Printf("  %s %s\n", color.Arrow(), name)
	}
	fmt.Println()

	// Show thinking steps for top model (unless --quick)
	var reasoning string
	if !decideQuick {
		effectiveLLM := llmCmd
		if effectiveLLM == "" {
			effectiveLLM = cfg.LLMCmd
		}
		if noLLM {
			effectiveLLM = ""
		}

		result, err := think.Think(decision, idx, modelFiles, 3, modelSlugs, effectiveLLM, verbose, timeout)
		if err == nil && len(result.Models) > 0 {
			// Show first model's thinking steps
			m := result.Models[0]
			fmt.Printf("%s\n", color.BoldCyan(fmt.Sprintf("## %s — Thinking Steps", m.ModelName)))
			if m.Synthesis != "" {
				fmt.Println(m.Synthesis)
			} else {
				for i, step := range m.Steps {
					fmt.Printf("%d. %s\n", i+1, step)
				}
			}
			fmt.Println()

			// Capture reasoning from all models
			var rb strings.Builder
			for _, mr := range result.Models {
				rb.WriteString(fmt.Sprintf("### %s\n", mr.ModelName))
				if mr.Synthesis != "" {
					rb.WriteString(mr.Synthesis)
				} else {
					for i, step := range mr.Steps {
						rb.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
					}
				}
				rb.WriteString("\n\n")
			}
			if result.Summary != "" {
				rb.WriteString("### Synthesis\n")
				rb.WriteString(result.Summary)
			}
			reasoning = rb.String()
		} else if err == nil {
			// Fallback: show static steps from model files
			for _, slug := range modelSlugs[:1] {
				if content, ok := modelFiles[findModelPath(idx, slug)]; ok {
					model := modelfile.Parse(content)
					if model != nil && len(model.ThinkingSteps) > 0 {
						fmt.Printf("%s\n", color.BoldCyan(fmt.Sprintf("## %s — Thinking Steps", model.Name)))
						for i, step := range model.ThinkingSteps {
							fmt.Printf("%d. %s\n", i+1, step)
						}
						fmt.Println()
					}
				}
			}
		}
	}

	// Get prediction
	prediction := decidePrediction
	if prediction == "" {
		fmt.Printf("%s ", color.Bold("What do you expect to happen?"))
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read prediction: %w", err)
		}
		prediction = strings.TrimSpace(line)
	}

	if prediction == "" {
		prediction = "_No prediction made_"
	}

	// Generate entry
	now := time.Now()
	dateStr := now.Format("2006-01-02")

	dir := journal.JournalDir()
	if decideProject {
		dir = journal.ProjectJournalDir()
	}
	// Ensure dir exists for ID generation
	os.MkdirAll(dir, 0755)

	id := journal.GenerateIDForDir(dir, now.Format("20060102"))

	review30 := now.AddDate(0, 1, 0).Format("2006-01-02")
	review90 := now.AddDate(0, 3, 0).Format("2006-01-02")

	entry := &journal.Entry{
		ID:          id,
		Date:        dateStr,
		Decision:    decision,
		Models:      modelSlugs,
		Reasoning:   reasoning,
		Prediction:  prediction,
		Status:      "open",
		ReviewDates: []string{review30, review90},
	}

	if decideProject {
		entry.Project = "local"
	}

	path, err := journal.Save(entry, decideProject)
	if err != nil {
		return err
	}

	// Vault sync
	syncToVault(path, cfg)

	fmt.Printf("\n%s Decision recorded: %s\n", color.Checkmark(), color.Bold(id))
	fmt.Printf("  %s %s\n", color.Arrow(), path)
	fmt.Printf("  %s Review dates: %s, %s\n", color.Arrow(), review30, review90)

	return nil
}

func findModelPath(idx *index.ModelIndex, slug string) string {
	entry := idx.FindBySlug(slug)
	if entry == nil {
		return ""
	}
	return entry.Path
}

func syncToVault(srcPath string, cfg *config.Config) {
	if cfg.VaultPath == "" {
		return
	}

	folder := cfg.VaultFolder
	if folder == "" {
		folder = "decisions"
	}

	destDir := filepath.Join(cfg.VaultPath, folder)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return
	}

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return
	}

	destPath := filepath.Join(destDir, filepath.Base(srcPath))
	_ = os.WriteFile(destPath, data, 0644)
}
