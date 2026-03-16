package journal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Entry represents a single decision journal entry.
type Entry struct {
	ID          string   `yaml:"id" json:"id"`
	Date        string   `yaml:"date" json:"date"`
	Decision    string   `yaml:"decision" json:"decision"`
	Models      []string `yaml:"models" json:"models"`
	Reasoning   string   `yaml:"reasoning,omitempty" json:"reasoning,omitempty"`
	Prediction  string   `yaml:"prediction" json:"prediction"`
	Status      string   `yaml:"status" json:"status"`
	Outcome     string   `yaml:"outcome,omitempty" json:"outcome,omitempty"`
	ReviewDates []string `yaml:"review_dates" json:"review_dates"`
	Project     string   `yaml:"project,omitempty" json:"project,omitempty"`
}

// JournalDir returns the global journal directory.
func JournalDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".lattice", "journal")
	}
	return filepath.Join(home, ".config", "lattice", "journal")
}

// ProjectJournalDir returns the project-local decisions directory.
func ProjectJournalDir() string {
	return filepath.Join(".", "decisions")
}

// Save writes an entry as a markdown file with YAML frontmatter.
func Save(entry *Entry, projectMode bool) (string, error) {
	dir := JournalDir()
	if projectMode {
		dir = ProjectJournalDir()
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create journal directory: %w", err)
	}

	filename := entry.ID + ".md"
	path := filepath.Join(dir, filename)

	content := renderMarkdown(entry)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write journal entry: %w", err)
	}

	return path, nil
}

// List reads all entries from a directory, sorted by date descending.
func List(dir string, limit int) ([]Entry, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read journal directory: %w", err)
	}

	var entries []Entry
	for _, de := range dirEntries {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, de.Name())
		entry, err := LoadEntry(path)
		if err != nil {
			continue
		}
		entries = append(entries, *entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date > entries[j].Date
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}

// LoadEntry loads a single entry from a markdown file with YAML frontmatter.
func LoadEntry(path string) (*Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read entry: %w", err)
	}

	content := string(data)

	// Extract YAML frontmatter between --- markers
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("no frontmatter found in %s", path)
	}

	rest := content[4:] // skip opening ---\n
	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return nil, fmt.Errorf("no closing frontmatter marker in %s", path)
	}

	frontmatter := rest[:endIdx]

	var entry Entry
	if err := yaml.Unmarshal([]byte(frontmatter), &entry); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Extract reasoning from markdown body if not in frontmatter
	body := rest[endIdx+4:] // skip \n---
	if entry.Reasoning == "" {
		entry.Reasoning = extractSection(body, "## Reasoning")
	}
	if entry.Outcome == "" || entry.Outcome == "_Pending review_" {
		outcome := extractSection(body, "## Outcome")
		if outcome != "" && outcome != "_Pending review_" {
			entry.Outcome = outcome
		}
	}

	return &entry, nil
}

// DueForReview returns entries where today >= any review_date and status is "open".
func DueForReview(entries []Entry) []Entry {
	today := time.Now().Format("2006-01-02")
	var due []Entry
	for _, e := range entries {
		if e.Status != "open" {
			continue
		}
		for _, rd := range e.ReviewDates {
			if rd <= today {
				due = append(due, e)
				break
			}
		}
	}
	return due
}

// GenerateID returns a decision ID in "d-YYYYMMDD-NNN" format.
func GenerateID() string {
	now := time.Now()
	dateStr := now.Format("20060102")
	return GenerateIDForDir(JournalDir(), dateStr)
}

// GenerateIDForDir generates an ID checking a specific directory for existing entries.
func GenerateIDForDir(dir, dateStr string) string {
	prefix := fmt.Sprintf("d-%s-", dateStr)

	// Count existing entries for today
	count := 1
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), prefix) {
				count++
			}
		}
	}

	return fmt.Sprintf("d-%s-%03d", dateStr, count)
}

func renderMarkdown(e *Entry) string {
	var b strings.Builder

	// YAML frontmatter
	b.WriteString("---\n")
	fm, _ := yaml.Marshal(e)
	b.Write(fm)
	b.WriteString("---\n\n")

	// Markdown body
	b.WriteString(fmt.Sprintf("# Decision: %s\n\n", e.Decision))

	b.WriteString("## Models Applied\n")
	for _, m := range e.Models {
		b.WriteString(fmt.Sprintf("- %s\n", m))
	}
	b.WriteString("\n")

	b.WriteString("## Reasoning\n")
	if e.Reasoning != "" {
		b.WriteString(e.Reasoning)
	} else {
		b.WriteString("_No reasoning captured_")
	}
	b.WriteString("\n\n")

	b.WriteString("## Prediction\n")
	if e.Prediction != "" {
		b.WriteString(e.Prediction)
	} else {
		b.WriteString("_No prediction made_")
	}
	b.WriteString("\n\n")

	b.WriteString("## Outcome\n")
	if e.Outcome != "" {
		b.WriteString(e.Outcome)
	} else {
		b.WriteString("_Pending review_")
	}
	b.WriteString("\n")

	return b.String()
}

func extractSection(body, header string) string {
	idx := strings.Index(body, header)
	if idx < 0 {
		return ""
	}

	rest := body[idx+len(header):]
	// Skip the newline after header
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	}

	// Find next ## header or end of string
	nextHeader := strings.Index(rest, "\n## ")
	if nextHeader >= 0 {
		rest = rest[:nextHeader]
	}

	return strings.TrimSpace(rest)
}
