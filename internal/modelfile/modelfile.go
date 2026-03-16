package modelfile

import (
	"strings"
)

// Model represents a parsed mental model markdown file.
type Model struct {
	Name              string   `json:"name"`
	Category          string   `json:"category"`
	Description       string   `json:"description"`
	WhenToAvoid       string   `json:"when_to_avoid"`
	Keywords          string   `json:"keywords"`
	ThinkingSteps     []string `json:"thinking_steps"`
	CoachingQuestions  []string `json:"coaching_questions"`
}

// Parse parses a mental model markdown file into a Model struct.
func Parse(content string) *Model {
	m := &Model{}
	lines := strings.Split(content, "\n")

	var currentSection string
	var sectionBuf strings.Builder

	flushSection := func() {
		text := strings.TrimSpace(sectionBuf.String())
		switch currentSection {
		case "description":
			m.Description = text
		case "avoid":
			m.WhenToAvoid = text
		case "keywords":
			m.Keywords = text
		case "steps":
			m.ThinkingSteps = parseNumberedList(text)
		case "questions":
			m.CoachingQuestions = parseBulletList(text)
		}
		sectionBuf.Reset()
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Parse title
		if strings.HasPrefix(trimmed, "## Mental Model =") {
			m.Name = strings.TrimSpace(strings.TrimPrefix(trimmed, "## Mental Model ="))
			continue
		}

		// Parse category
		if strings.HasPrefix(trimmed, "**Category =") {
			cat := strings.TrimPrefix(trimmed, "**Category =")
			cat = strings.TrimSuffix(cat, "**")
			m.Category = strings.TrimSpace(cat)
			continue
		}

		// Detect section headers
		if strings.HasPrefix(trimmed, "**Description:") {
			flushSection()
			currentSection = "description"
			continue
		}
		if strings.HasPrefix(trimmed, "**When to Avoid") {
			flushSection()
			currentSection = "avoid"
			continue
		}
		if strings.HasPrefix(trimmed, "**Keywords for Situations:") {
			flushSection()
			currentSection = "keywords"
			continue
		}
		if strings.HasPrefix(trimmed, "**Thinking Steps:") {
			flushSection()
			currentSection = "steps"
			continue
		}
		if strings.HasPrefix(trimmed, "**Coaching Questions:") {
			flushSection()
			currentSection = "questions"
			continue
		}

		if currentSection != "" {
			sectionBuf.WriteString(line)
			sectionBuf.WriteString("\n")
		}
	}
	flushSection()

	return m
}

func parseNumberedList(text string) []string {
	var items []string
	lines := strings.Split(text, "\n")
	var current strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check if starts with a number like "1." "2." etc.
		if len(trimmed) >= 2 && trimmed[0] >= '1' && trimmed[0] <= '9' && trimmed[1] == '.' {
			if current.Len() > 0 {
				items = append(items, strings.TrimSpace(current.String()))
				current.Reset()
			}
			// Strip the number prefix
			rest := strings.TrimSpace(trimmed[2:])
			current.WriteString(rest)
		} else {
			if current.Len() > 0 {
				current.WriteString(" ")
			}
			current.WriteString(trimmed)
		}
	}
	if current.Len() > 0 {
		items = append(items, strings.TrimSpace(current.String()))
	}
	return items
}

func parseBulletList(text string) []string {
	var items []string
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Remove leading bullet markers
		trimmed = strings.TrimPrefix(trimmed, "- ")
		trimmed = strings.TrimPrefix(trimmed, "* ")
		items = append(items, strings.TrimSpace(trimmed))
	}
	return items
}
