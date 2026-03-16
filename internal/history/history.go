package history

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Entry represents a history entry for a think or apply session
type Entry struct {
	Timestamp string   `json:"timestamp"`
	Slug      string   `json:"slug"`
	Type      string   `json:"type"` // "think" or "apply"
	Problem   string   `json:"problem"`
	Models    []string `json:"models,omitempty"`
	Summary   string   `json:"summary,omitempty"`
}

// Manager handles history storage and retrieval
type Manager struct {
	dir string
}

// NewManager creates a new history manager
func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	dir := filepath.Join(home, ".config", "lattice", "history")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}
	return &Manager{dir: dir}, nil
}

// Save saves a history entry
func (m *Manager) Save(entry *Entry) error {
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Create filename from timestamp and slug
	ts := strings.ReplaceAll(entry.Timestamp, ":", "-")
	ts = strings.ReplaceAll(ts, "T", "-")
	ts = strings.ReplaceAll(ts, "Z", "")
	ts = strings.Split(ts, ".")[0] // Remove subseconds if present

	slug := entry.Slug
	if slug == "" {
		slug = "session"
	}
	slug = strings.ToLower(strings.ReplaceAll(slug, " ", "-"))

	filename := fmt.Sprintf("%s-%s-%04x.json", ts, slug, rand.Intn(65536))
	path := filepath.Join(m.dir, filename)

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history entry: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// List returns the most recent history entries
func (m *Manager) List(limit int) ([]*Entry, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read history directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			files = append(files, entry.Name())
		}
	}

	// Sort by filename (which starts with timestamp)
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	if limit > 0 && len(files) > limit {
		files = files[:limit]
	}

	var results []*Entry
	for _, f := range files {
		path := filepath.Join(m.dir, f)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var entry Entry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		results = append(results, &entry)
	}

	return results, nil
}

// Clear deletes all history entries
func (m *Manager) Clear() error {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read history directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			path := filepath.Join(m.dir, entry.Name())
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// FormatList formats history entries for display
func FormatList(entries []*Entry) string {
	var b strings.Builder

	if len(entries) == 0 {
		b.WriteString("No history entries found.\n")
		b.WriteString("Run 'lattice think' or 'lattice apply' to create entries.\n")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("Recent sessions (%d):\n\n", len(entries)))
	for i, e := range entries {
		// Parse timestamp for nicer display
		ts := e.Timestamp
		if t, err := time.Parse(time.RFC3339, e.Timestamp); err == nil {
			ts = t.Format("2006-01-02 15:04")
		}

		b.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, ts, e.Type))
		b.WriteString(fmt.Sprintf("   %s\n", truncate(e.Problem, 60)))
		if len(e.Models) > 0 {
			b.WriteString(fmt.Sprintf("   Models: %s\n", strings.Join(e.Models, ", ")))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// FormatJSON returns entries as JSON
func FormatJSON(entries []*Entry) (string, error) {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
