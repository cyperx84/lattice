package cmd

import (
	"os"

	"github.com/cyperx84/lattice/internal/mcp"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server on stdio (Model Context Protocol)",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	idx, modelFiles, err := loadAllData()
	if err != nil {
		return err
	}

	server := mcp.NewServer(idx, modelFiles, verbose, os.Stderr)
	return server.Run(os.Stdin, os.Stdout)
}
