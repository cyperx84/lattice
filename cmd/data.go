package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/cyperx84/lattice/internal/index"
)

var dataFS embed.FS

// SetDataFS sets the embedded filesystem containing model data.
func SetDataFS(fs embed.FS) {
	dataFS = fs
}

// loadEmbeddedData loads the model index and all model files from the embedded FS.
func loadEmbeddedData() (*index.ModelIndex, map[string]string, error) {
	// Load index
	indexData, err := dataFS.ReadFile("data/model-index.json")
	if err != nil {
		return nil, nil, fmt.Errorf("read model-index.json: %w", err)
	}

	idx, err := index.Load(indexData)
	if err != nil {
		return nil, nil, fmt.Errorf("parse model-index.json: %w", err)
	}

	// Load all model files
	modelFiles := make(map[string]string)
	err = fs.WalkDir(dataFS, "data/models", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		content, err := dataFS.ReadFile(path)
		if err != nil {
			return err
		}
		// Store with the relative path that matches the index (e.g., "models/Mental_Model_General/m07_inversion.md")
		relPath := strings.TrimPrefix(path, "data/")
		modelFiles[relPath] = string(content)
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("walk model files: %w", err)
	}

	return idx, modelFiles, nil
}
