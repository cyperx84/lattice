package cmd

import (
	"fmt"
	"strings"

	"github.com/cyperx84/lattice/internal/apply"
	"github.com/cyperx84/lattice/internal/config"
	"github.com/cyperx84/lattice/internal/modelfile"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply <model-slug> <context>",
	Short: "Apply one model's thinking steps to a context",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runApply,
}

func init() {
	rootCmd.AddCommand(applyCmd)
}

func runApply(cmd *cobra.Command, args []string) error {
	slug := args[0]
	context := strings.Join(args[1:], " ")
	cfg := config.Load()

	idx, modelFiles, err := loadAllData()
	if err != nil {
		return err
	}

	entry := idx.FindBySlug(slug)
	if entry == nil {
		entry = idx.FindByID(slug)
	}
	if entry == nil {
		// Suggest similar models
		similar := idx.Search(slug)
		if len(similar) > 5 {
			similar = similar[:5]
		}
		if len(similar) > 0 {
			var slugs []string
			for _, s := range similar {
				slugs = append(slugs, s.Slug)
			}
			return fmt.Errorf("model not found: %s\n\nDid you mean one of these?\n  %s", slug, strings.Join(slugs, "\n  "))
		}
		return fmt.Errorf("model not found: %s\n\nRun 'lattice list' to see all available models", slug)
	}

	content, ok := modelFiles[entry.Path]
	if !ok {
		return fmt.Errorf("model file not found: %s", entry.Path)
	}

	model := modelfile.Parse(content)
	if model.Name == "" {
		model.Name = entry.Name
	}
	if model.Category == "" {
		model.Category = entry.Category
	}

	effectiveLLM := llmCmd
	if effectiveLLM == "" {
		effectiveLLM = cfg.LLMCmd
	}
	if noLLM {
		effectiveLLM = ""
	}

	result, err := apply.Apply(model, slug, context, effectiveLLM, verbose, timeout)
	if err != nil {
		return err
	}

	if jsonOutput {
		out, err := apply.FormatJSON(result)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		fmt.Print(apply.FormatResult(result))
	}

	return nil
}
