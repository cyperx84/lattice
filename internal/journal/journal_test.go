package journal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()

	entry := &Entry{
		ID:          "d-20260317-001",
		Date:        "2026-03-17",
		Decision:    "Raise prices 20%",
		Models:      []string{"trade-offs", "second-order-thinking", "scarcity"},
		Reasoning:   "Analysis of pricing impact",
		Prediction:  "Churn stays under 5%, revenue up 15% in 3 months",
		Status:      "open",
		ReviewDates: []string{"2026-04-17", "2026-06-17"},
	}

	// Override JournalDir by saving directly
	dir := tmpDir
	os.MkdirAll(dir, 0755)

	filename := entry.ID + ".md"
	path := filepath.Join(dir, filename)
	content := renderMarkdown(entry)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Expected file to exist")
	}

	// Verify content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, "id: d-20260317-001") {
		t.Error("Expected ID in frontmatter")
	}
	if !strings.Contains(s, "Raise prices 20%") {
		t.Error("Expected decision in content")
	}
	if !strings.Contains(s, "trade-offs") {
		t.Error("Expected models in content")
	}
	if !strings.Contains(s, "Churn stays under 5%") {
		t.Error("Expected prediction in content")
	}
	if !strings.Contains(s, "_Pending review_") {
		t.Error("Expected pending outcome")
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()

	original := &Entry{
		ID:          "d-20260317-001",
		Date:        "2026-03-17",
		Decision:    "Raise prices 20%",
		Models:      []string{"trade-offs", "second-order-thinking"},
		Reasoning:   "Detailed analysis here",
		Prediction:  "Revenue up 15%",
		Status:      "open",
		ReviewDates: []string{"2026-04-17", "2026-06-17"},
	}

	path := filepath.Join(tmpDir, original.ID+".md")
	content := renderMarkdown(original)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	loaded, err := LoadEntry(path)
	if err != nil {
		t.Fatalf("LoadEntry failed: %v", err)
	}

	if loaded.ID != original.ID {
		t.Errorf("ID mismatch: got %q, want %q", loaded.ID, original.ID)
	}
	if loaded.Date != original.Date {
		t.Errorf("Date mismatch: got %q, want %q", loaded.Date, original.Date)
	}
	if loaded.Decision != original.Decision {
		t.Errorf("Decision mismatch: got %q, want %q", loaded.Decision, original.Decision)
	}
	if len(loaded.Models) != len(original.Models) {
		t.Errorf("Models count mismatch: got %d, want %d", len(loaded.Models), len(original.Models))
	}
	if loaded.Prediction != original.Prediction {
		t.Errorf("Prediction mismatch: got %q, want %q", loaded.Prediction, original.Prediction)
	}
	if loaded.Status != original.Status {
		t.Errorf("Status mismatch: got %q, want %q", loaded.Status, original.Status)
	}
	if len(loaded.ReviewDates) != 2 {
		t.Errorf("ReviewDates count mismatch: got %d, want 2", len(loaded.ReviewDates))
	}
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()

	dates := []string{"2026-03-15", "2026-03-17", "2026-03-16"}
	for i, date := range dates {
		entry := &Entry{
			ID:          fmt.Sprintf("d-%s-001", strings.ReplaceAll(date, "-", "")),
			Date:        date,
			Decision:    fmt.Sprintf("Decision %d", i+1),
			Models:      []string{"inversion"},
			Prediction:  "Test prediction",
			Status:      "open",
			ReviewDates: []string{"2026-06-17"},
		}
		path := filepath.Join(tmpDir, entry.ID+".md")
		content := renderMarkdown(entry)
		os.WriteFile(path, []byte(content), 0644)
	}

	entries, err := List(tmpDir, 0)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// Should be sorted by date descending
	if entries[0].Date != "2026-03-17" {
		t.Errorf("First entry should be newest, got %s", entries[0].Date)
	}
	if entries[1].Date != "2026-03-16" {
		t.Errorf("Second entry should be middle, got %s", entries[1].Date)
	}
	if entries[2].Date != "2026-03-15" {
		t.Errorf("Third entry should be oldest, got %s", entries[2].Date)
	}
}

func TestListLimit(t *testing.T) {
	tmpDir := t.TempDir()

	for i := 0; i < 5; i++ {
		entry := &Entry{
			ID:          fmt.Sprintf("d-20260317-%03d", i+1),
			Date:        "2026-03-17",
			Decision:    fmt.Sprintf("Decision %d", i+1),
			Models:      []string{"inversion"},
			Prediction:  "Test",
			Status:      "open",
			ReviewDates: []string{"2026-06-17"},
		}
		path := filepath.Join(tmpDir, entry.ID+".md")
		os.WriteFile(path, []byte(renderMarkdown(entry)), 0644)
	}

	entries, err := List(tmpDir, 3)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries with limit, got %d", len(entries))
	}
}

func TestDueForReview(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	past := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	future := time.Now().AddDate(0, 0, 30).Format("2006-01-02")

	entries := []Entry{
		{ID: "d-1", Status: "open", ReviewDates: []string{past}},
		{ID: "d-2", Status: "open", ReviewDates: []string{future}},
		{ID: "d-3", Status: "open", ReviewDates: []string{today}},
		{ID: "d-4", Status: "resolved", ReviewDates: []string{past}},
		{ID: "d-5", Status: "open", ReviewDates: []string{future, past}},
	}

	due := DueForReview(entries)

	// d-1 (past), d-3 (today), d-5 (has past date) should be due
	if len(due) != 3 {
		t.Fatalf("Expected 3 due entries, got %d", len(due))
	}

	ids := make(map[string]bool)
	for _, e := range due {
		ids[e.ID] = true
	}

	if !ids["d-1"] {
		t.Error("Expected d-1 (past review date) to be due")
	}
	if !ids["d-3"] {
		t.Error("Expected d-3 (today review date) to be due")
	}
	if !ids["d-5"] {
		t.Error("Expected d-5 (has past review date) to be due")
	}
	if ids["d-2"] {
		t.Error("d-2 (future only) should not be due")
	}
	if ids["d-4"] {
		t.Error("d-4 (resolved) should not be due")
	}
}

func TestGenerateID(t *testing.T) {
	tmpDir := t.TempDir()
	dateStr := "20260317"

	id := GenerateIDForDir(tmpDir, dateStr)

	if !strings.HasPrefix(id, "d-20260317-") {
		t.Errorf("ID should start with d-20260317-, got %s", id)
	}
	if id != "d-20260317-001" {
		t.Errorf("First ID should be d-20260317-001, got %s", id)
	}

	// Create a file to simulate existing entry
	os.WriteFile(filepath.Join(tmpDir, "d-20260317-001.md"), []byte("test"), 0644)

	id2 := GenerateIDForDir(tmpDir, dateStr)
	if id2 != "d-20260317-002" {
		t.Errorf("Second ID should be d-20260317-002, got %s", id2)
	}
}

func TestListEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	entries, err := List(tmpDir, 10)
	if err != nil {
		t.Fatalf("List on empty dir failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(entries))
	}
}

func TestListNonexistentDir(t *testing.T) {
	entries, err := List("/nonexistent/path/xyz", 10)
	if err != nil {
		t.Fatalf("List on nonexistent dir should not error, got: %v", err)
	}
	if entries != nil {
		t.Errorf("Expected nil entries, got %v", entries)
	}
}

