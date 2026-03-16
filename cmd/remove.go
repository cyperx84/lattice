package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <slug>",
	Short: "Remove a user-added mental model",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	slug := args[0]

	idx, _, err := loadAllData()
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

	// Check if this is a local (user-added) model
	if !strings.HasPrefix(entry.Path, "local/") {
		return fmt.Errorf("cannot remove built-in model '%s'. Only user-added models in ~/.config/lattice/models/ can be removed", entry.Slug)
	}

	// Resolve local file path
	dir := localModelsDir()
	if dir == "" {
		return fmt.Errorf("could not determine home directory")
	}

	filename := strings.TrimPrefix(entry.Path, "local/")
	path := filepath.Join(dir, filename)

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove model file: %w", err)
	}

	if jsonOutput {
		fmt.Printf("{\"removed\":%q,\"slug\":%q,\"path\":%q}\n", entry.Name, entry.Slug, path)
	} else {
		fmt.Printf("Removed: %s (%s)\n", entry.Name, entry.Slug)
		fmt.Printf("  Deleted: %s\n", path)
	}

	return nil
}
