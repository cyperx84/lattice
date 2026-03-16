package index

import (
	"encoding/json"
	"strings"
)

// ModelIndex is the root structure of model-index.json.
type ModelIndex struct {
	Version     string              `json:"version"`
	Generated   string              `json:"generated"`
	TotalModels int                 `json:"total_models"`
	Categories  map[string][]string `json:"categories"`
	Models      []ModelEntry        `json:"models"`
}

// ModelEntry represents one model in the index.
type ModelEntry struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Slug     string   `json:"slug"`
	Category string   `json:"category"`
	Path     string   `json:"path"`
	Keywords []string `json:"keywords"`
	Summary  string   `json:"summary"`
}

// Load parses a model-index.json from bytes.
func Load(data []byte) (*ModelIndex, error) {
	var idx ModelIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

// Search finds models matching a keyword query (substring match on name, keywords, summary).
func (idx *ModelIndex) Search(query string) []ModelEntry {
	q := strings.ToLower(query)
	var results []ModelEntry
	for _, m := range idx.Models {
		if matchesQuery(m, q) {
			results = append(results, m)
		}
	}
	return results
}

// FilterByCategory returns models in the given category.
func (idx *ModelIndex) FilterByCategory(category string) []ModelEntry {
	cat := strings.ToLower(category)
	var results []ModelEntry
	for _, m := range idx.Models {
		if strings.ToLower(m.Category) == cat {
			results = append(results, m)
		}
	}
	return results
}

// FindBySlug returns the model with the given slug.
func (idx *ModelIndex) FindBySlug(slug string) *ModelEntry {
	s := strings.ToLower(slug)
	for _, m := range idx.Models {
		if strings.ToLower(m.Slug) == s {
			return &m
		}
	}
	return nil
}

// FindByID returns the model with the given ID.
func (idx *ModelIndex) FindByID(id string) *ModelEntry {
	for _, m := range idx.Models {
		if m.ID == id {
			return &m
		}
	}
	return nil
}

// TopNForQuery returns the top N models matching the query, scored by relevance.
func (idx *ModelIndex) TopNForQuery(query string, n int) []ModelEntry {
	q := strings.ToLower(query)
	type scored struct {
		entry ModelEntry
		score int
	}
	var matches []scored
	for _, m := range idx.Models {
		s := scoreMatch(m, q)
		if s > 0 {
			matches = append(matches, scored{m, s})
		}
	}

	// Sort by score descending (simple insertion sort for small N)
	for i := 1; i < len(matches); i++ {
		for j := i; j > 0 && matches[j].score > matches[j-1].score; j-- {
			matches[j], matches[j-1] = matches[j-1], matches[j]
		}
	}

	if n > len(matches) {
		n = len(matches)
	}
	results := make([]ModelEntry, n)
	for i := 0; i < n; i++ {
		results[i] = matches[i].entry
	}
	return results
}

func matchesQuery(m ModelEntry, q string) bool {
	return scoreMatch(m, q) > 0
}

func scoreMatch(m ModelEntry, q string) int {
	score := 0
	name := strings.ToLower(m.Name)
	slug := strings.ToLower(m.Slug)
	summary := strings.ToLower(m.Summary)

	// Name match is highest priority
	if strings.Contains(name, q) {
		score += 10
	}
	if strings.Contains(slug, q) {
		score += 8
	}

	// Keywords match
	for _, kw := range m.Keywords {
		if strings.Contains(strings.ToLower(kw), q) {
			score += 5
			break
		}
	}

	// Summary match
	if strings.Contains(summary, q) {
		score += 2
	}

	// Multi-word: check if all words match somewhere
	words := strings.Fields(q)
	if len(words) > 1 {
		allFound := true
		for _, w := range words {
			found := strings.Contains(name, w) || strings.Contains(summary, w)
			if !found {
				for _, kw := range m.Keywords {
					if strings.Contains(strings.ToLower(kw), w) {
						found = true
						break
					}
				}
			}
			if !found {
				allFound = false
				break
			}
		}
		if allFound {
			score += 3
		}
	}

	return score
}
