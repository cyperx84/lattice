package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cyperx84/lattice/internal/modelfile"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <slug>",
	Short: "Show full details for a mental model",
	Args:  cobra.ExactArgs(1),
	RunE:  runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	slug := args[0]

	idx, modelFiles, err := loadAllData()
	if err != nil {
		return err
	}

	entry := idx.FindBySlug(slug)
	if entry == nil {
		entry = idx.FindByID(slug)
	}
	if entry == nil {
		return fmt.Errorf("model not found: %s", slug)
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

	if jsonOutput {
		info := struct {
			Name     string   `json:"name"`
			ID       string   `json:"id"`
			Slug     string   `json:"slug"`
			Category string   `json:"category"`
			Desc     string   `json:"description"`
			Avoid    string   `json:"when_to_avoid"`
			Keywords string   `json:"keywords"`
			Steps    []string `json:"thinking_steps"`
			Qs       []string `json:"coaching_questions"`
		}{
			Name:     model.Name,
			ID:       entry.ID,
			Slug:     entry.Slug,
			Category: model.Category,
			Desc:     model.Description,
			Avoid:    model.WhenToAvoid,
			Keywords: model.Keywords,
			Steps:    model.ThinkingSteps,
			Qs:       model.CoachingQuestions,
		}
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", model.Name))
	b.WriteString(fmt.Sprintf("ID:       %s\n", entry.ID))
	b.WriteString(fmt.Sprintf("Slug:     %s\n", entry.Slug))
	b.WriteString(fmt.Sprintf("Category: %s\n\n", model.Category))

	if model.Description != "" {
		b.WriteString(fmt.Sprintf("## Description\n%s\n\n", model.Description))
	}

	if len(model.ThinkingSteps) > 0 {
		b.WriteString("## Thinking Steps\n")
		for i, step := range model.ThinkingSteps {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
		b.WriteString("\n")
	}

	if len(model.CoachingQuestions) > 0 {
		b.WriteString("## Coaching Questions\n")
		for _, q := range model.CoachingQuestions {
			b.WriteString(fmt.Sprintf("- %s\n", q))
		}
		b.WriteString("\n")
	}

	if model.WhenToAvoid != "" {
		b.WriteString(fmt.Sprintf("## When to Avoid\n%s\n\n", model.WhenToAvoid))
	}

	if model.Keywords != "" {
		b.WriteString(fmt.Sprintf("## Keywords\n%s\n", model.Keywords))
	}

	fmt.Print(b.String())
	return nil
}
