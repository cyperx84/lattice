package index

import (
	"encoding/json"
	"testing"
)

// testIndex builds a small ModelIndex for testing.
func testIndex() *ModelIndex {
	return &ModelIndex{
		Version:     "1.0.0",
		Generated:   "2025-01-01",
		TotalModels: 6,
		Categories: map[string][]string{
			"General Thinking Tools":    {"m01", "m02"},
			"Economics":                 {"m03", "m04"},
			"Human Nature & Judgment":   {"m05", "m06"},
		},
		Models: []ModelEntry{
			{ID: "m01", Name: "Inversion", Slug: "inversion", Category: "General Thinking Tools", Path: "models/m01_inversion.md", Keywords: []string{"flip", "reverse", "avoid failure"}, Summary: "Think backwards from the goal to identify what to avoid"},
			{ID: "m02", Name: "Second-Order Thinking", Slug: "second-order-thinking", Category: "General Thinking Tools", Path: "models/m02_second-order-thinking.md", Keywords: []string{"consequences", "long-term", "ripple effects"}, Summary: "Consider the consequences of the consequences"},
			{ID: "m03", Name: "Trade-offs & Opportunity Cost", Slug: "trade-offs", Category: "Economics", Path: "models/m03_trade-offs.md", Keywords: []string{"cost", "decision", "resource allocation", "hiring decisions"}, Summary: "Every decision closes other doors. Map the opportunity cost of each path."},
			{ID: "m04", Name: "Specialization", Slug: "specialization", Category: "Economics", Path: "models/m04_specialization.md", Keywords: []string{"focus", "mastery", "outsourcing", "team structure"}, Summary: "Focus drives mastery. Consider where specialization serves you best."},
			{ID: "m05", Name: "Social Proof", Slug: "social-proof", Category: "Human Nature & Judgment", Path: "models/m05_social-proof.md", Keywords: []string{"crowd", "herd", "conformity"}, Summary: "People follow what others do, especially under uncertainty"},
			{ID: "m06", Name: "Incentives", Slug: "incentives", Category: "Human Nature & Judgment", Path: "models/m06_incentives.md", Keywords: []string{"motivation", "reward", "behavior", "team"}, Summary: "Never think about what people say, think about what they are incentivized to do"},
		},
	}
}

func TestLoad(t *testing.T) {
	idx := testIndex()
	data, err := json.Marshal(idx)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.TotalModels != 6 {
		t.Errorf("expected 6 models, got %d", loaded.TotalModels)
	}
	if len(loaded.Models) != 6 {
		t.Errorf("expected 6 model entries, got %d", len(loaded.Models))
	}
	if loaded.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", loaded.Version)
	}
}

func TestSearch(t *testing.T) {
	idx := testIndex()

	results := idx.Search("inversion")
	if len(results) == 0 {
		t.Fatal("expected search results for 'inversion'")
	}
	if results[0].Slug != "inversion" {
		t.Errorf("expected first result to be inversion, got %s", results[0].Slug)
	}

	results = idx.Search("crowd")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'crowd', got %d", len(results))
	}
	if results[0].Slug != "social-proof" {
		t.Errorf("expected social-proof, got %s", results[0].Slug)
	}

	results = idx.Search("zzzznonexistent")
	if len(results) != 0 {
		t.Errorf("expected 0 results for nonexistent query, got %d", len(results))
	}
}

func TestTopNForQuery(t *testing.T) {
	idx := testIndex()

	results := idx.TopNForQuery("decision making trade-offs", 3)
	if len(results) == 0 {
		t.Fatal("expected results for 'decision making trade-offs'")
	}
	// trade-offs model should rank high
	found := false
	for _, r := range results {
		if r.Slug == "trade-offs" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected trade-offs model in top results")
	}
}

func TestConceptMapExpansion(t *testing.T) {
	idx := testIndex()

	// "hire" should expand to team/delegation/resource concepts and find relevant models
	results := idx.TopNForQuery("hire", 5)
	if len(results) == 0 {
		t.Fatal("expected results for 'hire' via concept expansion")
	}

	// Should find models with team/delegation/resource keywords
	slugs := make(map[string]bool)
	for _, r := range results {
		slugs[r.Slug] = true
	}
	// Trade-offs has "hiring decisions" keyword and "resource allocation"
	if !slugs["trade-offs"] {
		t.Error("expected 'trade-offs' to appear for 'hire' query (has hiring decisions keyword)")
	}
}

func TestFilterByCategory(t *testing.T) {
	idx := testIndex()

	results := idx.FilterByCategory("Economics")
	if len(results) != 2 {
		t.Errorf("expected 2 Economics models, got %d", len(results))
	}

	results = idx.FilterByCategory("Nonexistent Category")
	if len(results) != 0 {
		t.Errorf("expected 0 results for nonexistent category, got %d", len(results))
	}

	// Case insensitive
	results = idx.FilterByCategory("economics")
	if len(results) != 2 {
		t.Errorf("expected case-insensitive match, got %d", len(results))
	}
}

func TestFindBySlug(t *testing.T) {
	idx := testIndex()

	entry := idx.FindBySlug("inversion")
	if entry == nil {
		t.Fatal("expected to find inversion by slug")
	}
	if entry.ID != "m01" {
		t.Errorf("expected ID m01, got %s", entry.ID)
	}

	entry = idx.FindBySlug("Inversion") // case insensitive
	if entry == nil {
		t.Fatal("expected case-insensitive slug lookup")
	}

	entry = idx.FindBySlug("nonexistent")
	if entry != nil {
		t.Error("expected nil for nonexistent slug")
	}
}

func TestFindByID(t *testing.T) {
	idx := testIndex()

	entry := idx.FindByID("m03")
	if entry == nil {
		t.Fatal("expected to find m03 by ID")
	}
	if entry.Slug != "trade-offs" {
		t.Errorf("expected trade-offs, got %s", entry.Slug)
	}

	entry = idx.FindByID("m99")
	if entry != nil {
		t.Error("expected nil for nonexistent ID")
	}
}

func TestStopWords(t *testing.T) {
	// Verify common stop words are in the map
	words := []string{"a", "the", "is", "and", "or", "to", "of", "in", "for"}
	for _, w := range words {
		if !stopWords[w] {
			t.Errorf("expected '%s' to be a stop word", w)
		}
	}

	// Verify non-stop words are not in the map
	nonStop := []string{"decision", "risk", "strategy", "team"}
	for _, w := range nonStop {
		if stopWords[w] {
			t.Errorf("'%s' should not be a stop word", w)
		}
	}
}
