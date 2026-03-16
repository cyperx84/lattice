package modelfile

import (
	"testing"
)

const sampleModel = `## Mental Model = Inversion

**Category = General Thinking Tools**

**Description:**
Instead of thinking about how to achieve something, think about what would guarantee failure and avoid those things. This approach, championed by Charlie Munger, helps identify hidden risks and assumptions.

**When to Avoid This Model:**
When the problem is straightforward and doesn't benefit from reverse thinking. Over-applying inversion to simple decisions can lead to analysis paralysis.

**Keywords for Situations:**
problem solving, risk assessment, strategy, avoid failure, reverse thinking

**Thinking Steps:**
1. Define the desired outcome clearly
2. Ask: what would guarantee the opposite?
3. List all the ways you could fail
4. Invert those failure modes into action items
5. Prioritize the most impactful inversions

**Coaching Questions:**
- What would make this project fail spectacularly?
- Which assumptions are you not questioning?
- If you wanted to destroy value here, how would you do it?
- What are your competitors hoping you'll do wrong?
- What's the worst decision you could make right now?
`

func TestParse(t *testing.T) {
	model := Parse(sampleModel)

	if model.Name != "Inversion" {
		t.Errorf("expected Name 'Inversion', got %q", model.Name)
	}
	if model.Category != "General Thinking Tools" {
		t.Errorf("expected Category 'General Thinking Tools', got %q", model.Category)
	}
	if model.Description == "" {
		t.Error("expected non-empty Description")
	}
	if model.WhenToAvoid == "" {
		t.Error("expected non-empty WhenToAvoid")
	}
	if model.Keywords == "" {
		t.Error("expected non-empty Keywords")
	}
	if len(model.ThinkingSteps) != 5 {
		t.Errorf("expected 5 thinking steps, got %d", len(model.ThinkingSteps))
	}
	if len(model.CoachingQuestions) != 5 {
		t.Errorf("expected 5 coaching questions, got %d", len(model.CoachingQuestions))
	}

	// Verify content of first step
	if model.ThinkingSteps[0] != "Define the desired outcome clearly" {
		t.Errorf("unexpected first step: %q", model.ThinkingSteps[0])
	}

	// Verify content of first question
	if model.CoachingQuestions[0] != "What would make this project fail spectacularly?" {
		t.Errorf("unexpected first question: %q", model.CoachingQuestions[0])
	}
}

func TestParseEmpty(t *testing.T) {
	model := Parse("")

	if model.Name != "" {
		t.Errorf("expected empty Name, got %q", model.Name)
	}
	if model.Category != "" {
		t.Errorf("expected empty Category, got %q", model.Category)
	}
	if len(model.ThinkingSteps) != 0 {
		t.Errorf("expected 0 thinking steps, got %d", len(model.ThinkingSteps))
	}
	if len(model.CoachingQuestions) != 0 {
		t.Errorf("expected 0 coaching questions, got %d", len(model.CoachingQuestions))
	}
}

func TestParsePartial(t *testing.T) {
	// Model with only name, category, and steps — no description, avoid, keywords, or questions
	partial := `## Mental Model = Partial Model

**Category = Test**

**Thinking Steps:**
1. First step
2. Second step
`
	model := Parse(partial)

	if model.Name != "Partial Model" {
		t.Errorf("expected Name 'Partial Model', got %q", model.Name)
	}
	if model.Category != "Test" {
		t.Errorf("expected Category 'Test', got %q", model.Category)
	}
	if model.Description != "" {
		t.Errorf("expected empty Description, got %q", model.Description)
	}
	if model.WhenToAvoid != "" {
		t.Errorf("expected empty WhenToAvoid, got %q", model.WhenToAvoid)
	}
	if model.Keywords != "" {
		t.Errorf("expected empty Keywords, got %q", model.Keywords)
	}
	if len(model.ThinkingSteps) != 2 {
		t.Errorf("expected 2 thinking steps, got %d", len(model.ThinkingSteps))
	}
	if len(model.CoachingQuestions) != 0 {
		t.Errorf("expected 0 coaching questions, got %d", len(model.CoachingQuestions))
	}
}
