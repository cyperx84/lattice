package apply

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cyperx84/lattice/internal/modelfile"
)

// Result represents the output of applying a model.
type Result struct {
	ModelName  string   `json:"model_name"`
	ModelSlug  string   `json:"model_slug"`
	Category   string   `json:"category"`
	Steps      []string `json:"thinking_steps"`
	Questions  []string `json:"coaching_questions"`
	Context    string   `json:"context"`
	Synthesis  string   `json:"synthesis,omitempty"`
}

// Apply applies a mental model to a given context.
// If llmCmd is provided, it uses the LLM for synthesis; otherwise returns the model's steps.
func Apply(model *modelfile.Model, slug, context, llmCmd string, verbose bool, timeoutSec int) (*Result, error) {
	result := &Result{
		ModelName:  model.Name,
		ModelSlug:  slug,
		Category:   model.Category,
		Steps:      model.ThinkingSteps,
		Questions:  model.CoachingQuestions,
		Context:    context,
	}

	if llmCmd != "" {
		prompt := buildApplyPrompt(model, context)
		synthesis, err := callLLM(prompt, llmCmd, verbose, timeoutSec)
		if err != nil {
			if verbose {
				fmt.Printf("[apply] LLM call failed: %v, falling back to static output\n", err)
			}
		} else {
			result.Synthesis = strings.TrimSpace(synthesis)
		}
	}

	return result, nil
}

func buildApplyPrompt(model *modelfile.Model, context string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Apply the mental model \"%s\" to the following context.\n\n", model.Name))
	b.WriteString(fmt.Sprintf("## Model: %s\n", model.Name))
	b.WriteString(fmt.Sprintf("Category: %s\n\n", model.Category))
	b.WriteString(fmt.Sprintf("Description: %s\n\n", model.Description))

	b.WriteString("## Thinking Steps\n")
	for i, step := range model.ThinkingSteps {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
	}

	b.WriteString(fmt.Sprintf("\n## Context to Analyze\n%s\n\n", context))

	b.WriteString("## Instructions\n")
	b.WriteString("Walk through each thinking step applied to the context above.\n")
	b.WriteString("For each step, provide specific, actionable insights.\n")
	b.WriteString("End with a synthesis of the key takeaways.\n")
	b.WriteString("Be concise but thorough.\n")

	return b.String()
}

// FormatResult formats a Result as human-readable text.
func FormatResult(r *Result) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s\n", r.ModelName))
	b.WriteString(fmt.Sprintf("Category: %s\n\n", r.Category))

	if r.Synthesis != "" {
		b.WriteString(r.Synthesis)
		b.WriteString("\n")
	} else {
		b.WriteString("### Thinking Steps\n")
		for i, step := range r.Steps {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
		b.WriteString("\n### Coaching Questions\n")
		for _, q := range r.Questions {
			b.WriteString(fmt.Sprintf("- %s\n", q))
		}
	}

	return b.String()
}

// FormatJSON returns the result as indented JSON.
func FormatJSON(r *Result) (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func callLLM(prompt, cmdStr string, verbose bool, timeoutSec int) (string, error) {
	if verbose {
		fmt.Printf("[apply] Running LLM command: %s\n", cmdStr)
	}

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty LLM command")
	}

	if timeoutSec <= 0 {
		timeoutSec = 60
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Stdin = strings.NewReader(prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("LLM timed out after %ds", timeoutSec)
		}
		return "", fmt.Errorf("command failed: %w\nstderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
