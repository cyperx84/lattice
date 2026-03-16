package main

import (
	"embed"

	"github.com/cyperx84/lattice/cmd"
)

//go:embed all:data
var dataFS embed.FS

func main() {
	cmd.SetDataFS(dataFS)
	cmd.Execute()
}
