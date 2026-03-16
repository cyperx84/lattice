package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cyperx84/lattice/internal/config"
	"github.com/spf13/cobra"
)

var addFrom string

var addCmd = &cobra.Command{
	Use:   "add <model-name>",
	Short: "Add a new mental model to your local collection",
	Long: `Generate and save a new mental model to ~/.config/lattice/models/.
Uses an LLM to generate the model in the standard format.

Examples:
  lattice add "Network Effects"
  lattice add "Lindy Effect" --from "https://fs.blog/lindy-effect/"
  lattice add "OODA Loop" --from "Military decision-making framework by John Boyd"`,
	Args: cobra.ExactArgs(1),
	Run:  runAdd,
}

func init() {
	addCmd.Flags().StringVar(&addFrom, "from", "", "Source URL or description to inform model generation")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) {
	name := args[0]
	slug := toSlug(name)

	// Check if already exists (embedded or local)
	idx, _, err := loadAllData()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading data: %v\n", err)
		os.Exit(1)
	}

	if existing := idx.FindBySlug(slug); existing != nil {
		fmt.Fprintf(os.Stderr, "Model '%s' already exists (slug: %s, id: %s)\n", existing.Name, existing.Slug, existing.ID)
		os.Exit(1)
	}

	// Get LLM command
	cfg := config.Load()
	llmCmdStr := cfg.LLMCmd
	if llmCmd != "" {
		llmCmdStr = llmCmd
	}

	// Build prompt
	extra := ""
	if addFrom != "" {
		extra = fmt.Sprintf("Additional context/source: %s", addFrom)
	}
	prompt := fmt.Sprintf(addModelTemplate, name, extra)

	// Call LLM
	fmt.Printf("Generating model: %s...\n", name)
	content, err := runLLMForAdd(prompt, llmCmdStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	content = strings.TrimSpace(content)

	// Determine next ID
	nextID := nextLocalID(idx)

	// Save to local models dir
	dir := localModelsDir()
	if dir == "" {
		fmt.Fprintln(os.Stderr, "Error: could not determine home directory")
		os.Exit(1)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating models dir: %v\n", err)
		os.Exit(1)
	}

	filename := fmt.Sprintf("m%d_%s.md", nextID, slug)
	path := filepath.Join(dir, filename)

	if err := os.WriteFile(path, []byte(content+"\n"), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing model: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		fmt.Printf("{\"id\":\"m%d\",\"name\":%q,\"slug\":%q,\"path\":%q}\n", nextID, name, slug, path)
	} else {
		fmt.Printf("✓ Added: %s\n", name)
		fmt.Printf("  File: %s\n", path)
		fmt.Printf("  Slug: %s\n", slug)
		fmt.Printf("  ID:   m%d\n", nextID)
	}
}

func toSlug(name string) string {
	s := strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	return s
}

func runLLMForAdd(prompt, cmdStr string) (string, error) {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "", fmt.Errorf("no LLM command configured — set llm_cmd in ~/.config/lattice/config.yml")
	}

	timeoutSec := timeout
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
		return "", fmt.Errorf("LLM command failed: %w\n%s", err, stderr.String())
	}

	return stdout.String(), nil
}
