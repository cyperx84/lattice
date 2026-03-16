package think

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cyperx84/lattice/internal/apply"
	"github.com/cyperx84/lattice/internal/index"
	"github.com/cyperx84/lattice/internal/modelfile"
)

// Result represents the output of the think command.
type Result struct {
	Problem  string          `json:"problem"`
	Models   []apply.Result  `json:"models"`
	Summary  string          `json:"summary,omitempty"`
}

// Think searches for the top N models relevant to the problem, applies each, and synthesizes.
func Think(problem string, idx *index.ModelIndex, modelFiles map[string]string, n int, specificModels []string, llmCmd string, verbose bool, timeoutSec int) (*Result, error) {
	var entries []index.ModelEntry

	if len(specificModels) > 0 {
		for _, slug := range specificModels {
			entry := idx.FindBySlug(slug)
			if entry == nil {
				entry = idx.FindByID(slug)
			}
			if entry != nil {
				entries = append(entries, *entry)
			} else if verbose {
				fmt.Printf("[think] Model not found: %s\n", slug)
			}
		}
	} else {
		entries = idx.TopNForQuery(problem, n)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no relevant models found for: %s", problem)
	}

	result := &Result{
		Problem: problem,
	}

	for _, entry := range entries {
		content, ok := modelFiles[entry.Path]
		if !ok {
			if verbose {
				fmt.Printf("[think] Model file not found: %s\n", entry.Path)
			}
			continue
		}

		model := modelfile.Parse(content)
		if model.Name == "" {
			model.Name = entry.Name
		}
		if model.Category == "" {
			model.Category = entry.Category
		}

		applyResult, err := apply.Apply(model, entry.Slug, problem, llmCmd, verbose, timeoutSec)
		if err != nil {
			if verbose {
				fmt.Printf("[think] Failed to apply %s: %v\n", entry.Name, err)
			}
			continue
		}

		result.Models = append(result.Models, *applyResult)
	}

	// Synthesize if LLM is available
	if llmCmd != "" && len(result.Models) > 1 {
		summary, err := synthesize(problem, result.Models, llmCmd, verbose, timeoutSec)
		if err != nil {
			if verbose {
				fmt.Printf("[think] Synthesis failed: %v\n", err)
			}
		} else {
			result.Summary = summary
		}
	}

	return result, nil
}

func synthesize(problem string, models []apply.Result, llmCmd string, verbose bool, timeoutSec int) (string, error) {
	var b strings.Builder
	b.WriteString("Synthesize the insights from these mental models applied to the following problem.\n\n")
	b.WriteString(fmt.Sprintf("## Problem\n%s\n\n", problem))
	b.WriteString("## Models Applied\n\n")

	for _, m := range models {
		b.WriteString(fmt.Sprintf("### %s (%s)\n", m.ModelName, m.Category))
		if m.Synthesis != "" {
			b.WriteString(m.Synthesis)
		} else {
			b.WriteString("Thinking steps:\n")
			for i, step := range m.Steps {
				b.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
			}
		}
		b.WriteString("\n\n")
	}

	b.WriteString("## Instructions\n")
	b.WriteString("Provide a concise synthesis that:\n")
	b.WriteString("1. Identifies the key insight from combining these models\n")
	b.WriteString("2. Highlights where the models agree or reinforce each other\n")
	b.WriteString("3. Notes any tensions between the models\n")
	b.WriteString("4. Gives 2-3 concrete next actions based on the combined analysis\n")
	b.WriteString("Keep it under 300 words.\n")

	return callLLM(b.String(), llmCmd, verbose, timeoutSec)
}

// FormatResult formats a ThinkResult as human-readable text.
func FormatResult(r *Result) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Thinking: %s\n\n", r.Problem))
	b.WriteString(fmt.Sprintf("Models applied: %d\n\n", len(r.Models)))

	for _, m := range r.Models {
		b.WriteString(apply.FormatResult(&m))
		b.WriteString("\n---\n\n")
	}

	if r.Summary != "" {
		b.WriteString("## Synthesis\n\n")
		b.WriteString(r.Summary)
		b.WriteString("\n")
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
		fmt.Printf("[think] Running LLM command: %s\n", cmdStr)
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
