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

// conceptMap maps common query words to model-relevant concepts.
var conceptMap = map[string][]string{
	"hire":       {"team", "delegation", "resource", "organizational", "decision"},
	"hiring":     {"team", "delegation", "resource", "organizational", "decision"},
	"outsource":  {"specialization", "trade-offs", "efficiency", "interdependence", "delegation"},
	"build":      {"innovation", "process", "strategy", "resource", "decision"},
	"buy":        {"trade-offs", "optimization", "efficiency", "resource"},
	"sell":       {"valuation", "negotiation", "market", "pricing"},
	"invest":     {"risk", "investing", "capital", "allocation", "financial"},
	"launch":     {"strategy", "risk", "innovation", "timing"},
	"pivot":      {"adaptation", "strategy", "innovation", "change"},
	"scale":      {"growth", "system", "bottleneck", "efficiency", "leverage"},
	"grow":       {"growth", "strategy", "scale", "leverage"},
	"compete":    {"competition", "strategy", "competitive", "market"},
	"price":      {"pricing", "valuation", "market", "supply", "demand"},
	"lead":       {"leadership", "team", "management", "delegation", "trust"},
	"manage":     {"management", "leadership", "team", "organizational"},
	"decide":     {"decision", "trade-offs", "risk", "strategy"},
	"choose":     {"decision", "trade-offs", "opportunity"},
	"risk":       {"risk", "probability", "uncertainty", "margin"},
	"fail":       {"failure", "risk", "inversion", "resilience"},
	"start":      {"activation", "innovation", "momentum", "initiative"},
	"quit":       {"sunk cost", "trade-offs", "opportunity", "inversion"},
	"negotiate":  {"negotiation", "leverage", "incentives", "fairness"},
	"partner":    {"cooperation", "interdependence", "trust", "alliance"},
	"focus":      {"specialization", "priority", "resource", "competence"},
	"automate":   {"efficiency", "leverage", "optimization", "process"},
	"engineers":  {"team", "specialization", "resource", "competence"},
	"team":       {"team", "cooperation", "trust", "organizational"},
	"money":      {"financial", "capital", "resource", "investment"},
	"time":       {"resource", "efficiency", "priority", "trade-offs"},
	"product":    {"product", "market", "innovation", "design"},
	"customer":   {"audience", "market", "demand", "feedback"},
	"market":     {"market", "competition", "supply", "demand"},
}

// TopNForQuery returns the top N models matching the query, scored by relevance.
func (idx *ModelIndex) TopNForQuery(query string, n int) []ModelEntry {
	q := strings.ToLower(query)
	type scored struct {
		entry ModelEntry
		score int
	}
	var matches []scored

	// Also expand query words via concept map
	expanded := q
	for _, w := range strings.Fields(q) {
		if concepts, ok := conceptMap[w]; ok {
			expanded += " " + strings.Join(concepts, " ")
		}
	}

	for _, m := range idx.Models {
		s := scoreMatch(m, expanded)
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

// stopWords are common words to skip during scoring.
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "is": true, "are": true, "was": true,
	"were": true, "be": true, "been": true, "being": true, "have": true, "has": true,
	"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
	"could": true, "should": true, "may": true, "might": true, "shall": true,
	"can": true, "to": true, "of": true, "in": true, "for": true, "on": true,
	"with": true, "at": true, "by": true, "from": true, "or": true, "and": true,
	"not": true, "no": true, "but": true, "if": true, "then": true, "than": true,
	"so": true, "as": true, "it": true, "its": true, "i": true, "my": true,
	"me": true, "we": true, "our": true, "you": true, "your": true, "they": true,
	"their": true, "this": true, "that": true, "these": true, "those": true,
	"what": true, "which": true, "who": true, "whom": true, "how": true, "when": true,
	"where": true, "why": true, "more": true, "whether": true, "about": true,
	"im": true, "i'm": true,
}

func scoreMatch(m ModelEntry, q string) int {
	score := 0
	name := strings.ToLower(m.Name)
	slug := strings.ToLower(m.Slug)
	summary := strings.ToLower(m.Summary)

	// Exact/substring match on full query (highest priority)
	if strings.Contains(name, q) {
		score += 10
	}
	if strings.Contains(slug, q) {
		score += 8
	}

	// Keywords match on full query
	for _, kw := range m.Keywords {
		if strings.Contains(strings.ToLower(kw), q) {
			score += 5
			break
		}
	}

	// Summary match on full query
	if strings.Contains(summary, q) {
		score += 2
	}

	// Per-word scoring (for natural language queries)
	words := strings.Fields(q)
	meaningfulWords := 0
	matchedWords := 0
	for _, w := range words {
		if stopWords[w] || len(w) < 3 {
			continue
		}
		meaningfulWords++
		wordMatched := false

		if strings.Contains(name, w) {
			score += 4
			wordMatched = true
		}
		if strings.Contains(slug, w) {
			score += 3
			wordMatched = true
		}
		for _, kw := range m.Keywords {
			if strings.Contains(strings.ToLower(kw), w) {
				score += 3
				wordMatched = true
				break
			}
		}
		if strings.Contains(summary, w) {
			score += 1
			wordMatched = true
		}
		if wordMatched {
			matchedWords++
		}
	}

	// Bonus for matching multiple meaningful words
	if meaningfulWords > 0 && matchedWords >= 2 {
		score += matchedWords * 2
	}

	return score
}
