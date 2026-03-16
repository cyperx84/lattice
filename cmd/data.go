package cmd

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cyperx84/lattice/internal/index"
	"github.com/cyperx84/lattice/internal/modelfile"
)

var dataFS embed.FS

// SetDataFS sets the embedded filesystem containing model data.
func SetDataFS(fs embed.FS) {
	dataFS = fs
}

// localModelsDir returns ~/.config/lattice/models/
func localModelsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "lattice", "models")
}

// loadAllData loads embedded data then merges local models on top.
func loadAllData() (*index.ModelIndex, map[string]string, error) {
	idx, modelFiles, err := loadEmbeddedData()
	if err != nil {
		return nil, nil, err
	}

	// Merge local models
	dir := localModelsDir()
	if dir == "" {
		return idx, modelFiles, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		// No local dir — that's fine
		return idx, modelFiles, nil
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}

		// Parse to get metadata
		model := modelfile.Parse(string(content))
		slug := strings.TrimSuffix(e.Name(), ".md")

		// Remove any ID prefix like "m99_"
		cleanSlug := slug
		if len(slug) > 3 && slug[0] == 'm' && slug[1] >= '0' && slug[1] <= '9' {
			if idx := strings.Index(slug, "_"); idx > 0 {
				cleanSlug = slug[idx+1:]
			}
		}

		// Build an index path for the local file
		localPath := "local/" + e.Name()
		modelFiles[localPath] = string(content)

		// Check if this slug already exists — if so, override
		found := false
		for i, m := range idx.Models {
			if strings.ToLower(m.Slug) == strings.ToLower(cleanSlug) {
				// Override existing entry with local version
				idx.Models[i].Path = localPath
				found = true
				break
			}
		}

		if !found {
			// Add new entry
			newID := fmt.Sprintf("m%d", idx.TotalModels+1)
			category := model.Category
			if category == "" {
				category = "User Added"
			}
			entry := index.ModelEntry{
				ID:       newID,
				Name:     model.Name,
				Slug:     cleanSlug,
				Category: category,
				Path:     localPath,
				Keywords: parseKeywordString(model.Keywords),
				Summary:  truncate(model.Description, 200),
			}
			if entry.Name == "" {
				entry.Name = cleanSlug
			}
			idx.Models = append(idx.Models, entry)
			idx.TotalModels++

			// Add to categories map
			if _, ok := idx.Categories[category]; !ok {
				idx.Categories[category] = []string{}
			}
			idx.Categories[category] = append(idx.Categories[category], newID)
		}
	}

	return idx, modelFiles, nil
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
	err = fs.WalkDir(dataFS, "data/models", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		content, readErr := dataFS.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		relPath := strings.TrimPrefix(path, "data/")
		modelFiles[relPath] = string(content)
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("walk model files: %w", err)
	}

	return idx, modelFiles, nil
}

// nextLocalID returns the next available model ID number for local models.
func nextLocalID(idx *index.ModelIndex) int {
	max := idx.TotalModels
	for _, m := range idx.Models {
		if len(m.ID) > 1 && m.ID[0] == 'm' {
			var n int
			if _, err := fmt.Sscanf(m.ID[1:], "%d", &n); err == nil && n > max {
				max = n
			}
		}
	}
	return max + 1
}

// addModelTemplate is the template for generating a model via LLM.
const addModelTemplate = `Generate a mental model file in the exact format below.

The model name is: %s
%s

Output ONLY the markdown content, no extra commentary.

Format:
## Mental Model = <Name>

**Category = <appropriate category>**

**Description:**
<2-3 paragraph description of the mental model, its origins, and how it works>

**When to Avoid This Model:**
<When this model doesn't apply or could mislead>

**Keywords for Situations:**
<Comma-separated keywords for when this model is useful>

**Thinking Steps:**
1. <Step 1>
2. <Step 2>
3. <Step 3>
4. <Step 4>
5. <Step 5>

**Coaching Questions:**
- <Question 1>
- <Question 2>
- <Question 3>
- <Question 4>
- <Question 5>
`

func parseKeywordString(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// writeLocalIndexCache writes a merged index to ~/.config/lattice/index-cache.json.
// This is optional — used by other tools that want to read the full model list.
func writeLocalIndexCache(idx *index.ModelIndex) {
	dir := localModelsDir()
	if dir == "" {
		return
	}
	parent := filepath.Dir(dir)
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(parent, "index-cache.json"), data, 0644)
}
