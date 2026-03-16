package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManager(t *testing.T) {
	// Use a temp directory for testing
	tmpDir := t.TempDir()

	m := &Manager{dir: tmpDir}

	// Test Save
	entry := &Entry{
		Type:    "think",
		Problem: "Should we build or buy our auth system?",
		Models:  []string{"first-principles", "inversion"},
		Summary: "Test summary",
	}

	if err := m.Save(entry); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file was created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}
	if !strings.HasSuffix(files[0].Name(), ".json") {
		t.Errorf("Expected .json file, got %s", files[0].Name())
	}

	// Test List
	entries, err := m.List(10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Problem != entry.Problem {
		t.Errorf("Problem mismatch: got %q, want %q", entries[0].Problem, entry.Problem)
	}
	if entries[0].Type != "think" {
		t.Errorf("Type mismatch: got %q, want %q", entries[0].Type, "think")
	}
}

func TestManagerClear(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{dir: tmpDir}

	// Create some entries
	for i := 0; i < 3; i++ {
		entry := &Entry{
			Type:    "think",
			Problem: "Test problem",
		}
		if err := m.Save(entry); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	// Verify entries exist
	entries, err := m.List(10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Clear
	if err := m.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify cleared
	entries, err = m.List(10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(entries))
	}
}

func TestManagerListLimit(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{dir: tmpDir}

	// Create 5 entries
	for i := 0; i < 5; i++ {
		entry := &Entry{
			Type:    "think",
			Problem: "Test problem",
		}
		if err := m.Save(entry); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	// List with limit
	entries, err := m.List(3)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries with limit, got %d", len(entries))
	}
}

func TestFormatList(t *testing.T) {
	entries := []*Entry{
		{
			Timestamp: "2024-01-15T10:30:00Z",
			Type:      "think",
			Problem:   "Should we build or buy?",
			Models:    []string{"first-principles", "inversion"},
		},
		{
			Timestamp: "2024-01-14T15:00:00Z",
			Type:      "apply",
			Problem:   "Design microservices",
			Models:    []string{"systems-thinking"},
		},
	}

	output := FormatList(entries)

	if !strings.Contains(output, "think") {
		t.Error("Output should contain 'think'")
	}
	if !strings.Contains(output, "apply") {
		t.Error("Output should contain 'apply'")
	}
	if !strings.Contains(output, "Should we build or buy?") {
		t.Error("Output should contain problem text")
	}
}

func TestFormatListEmpty(t *testing.T) {
	output := FormatList(nil)

	if !strings.Contains(output, "No history") {
		t.Error("Empty list should contain 'No history'")
	}
}

func TestNewManager(t *testing.T) {
	// This test uses the real home directory
	m, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	expectedDir := filepath.Join(os.Getenv("HOME"), ".config", "lattice", "history")
	if m.dir != expectedDir {
		t.Errorf("dir = %q, want %q", m.dir, expectedDir)
	}
}
