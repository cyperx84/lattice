package color

import (
	"os"
	"strings"
	"testing"
)

func TestColorEnabled(t *testing.T) {
	// Save original state
	origEnabled := enabled
	origNoColor := os.Getenv("NO_COLOR")
	defer func() {
		enabled = origEnabled
		if origNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", origNoColor)
		}
	}()

	tests := []struct {
		name     string
		noColor  string
		expected bool
	}{
		{"color enabled by default", "", true},
		{"NO_COLOR disables color", "1", false},
		{"NO_COLOR any value disables", "true", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset for each test
			enabled = true
			if tt.noColor != "" {
				os.Setenv("NO_COLOR", tt.noColor)
				// Re-run init logic
				enabled = false
			} else {
				os.Unsetenv("NO_COLOR")
				enabled = true
			}

			if Enabled() != tt.expected {
				t.Errorf("Enabled() = %v, want %v", Enabled(), tt.expected)
			}
		})
	}
}

func TestColorFunctions(t *testing.T) {
	// Test with color enabled
	enabled = true

	if !strings.Contains(Bold("test"), "\033[1m") {
		t.Error("Bold should contain ANSI escape when enabled")
	}
	if !strings.Contains(Green("test"), "\033[32m") {
		t.Error("Green should contain ANSI escape when enabled")
	}
	if !strings.Contains(Cyan("test"), "\033[36m") {
		t.Error("Cyan should contain ANSI escape when enabled")
	}
	if !strings.Contains(Dim("test"), "\033[2m") {
		t.Error("Dim should contain ANSI escape when enabled")
	}
	if !strings.Contains(Blue("test"), "\033[34m") {
		t.Error("Blue should contain ANSI escape when enabled")
	}
	if !strings.Contains(Yellow("test"), "\033[33m") {
		t.Error("Yellow should contain ANSI escape when enabled")
	}
	if !strings.Contains(BoldCyan("test"), "\033[1;36m") {
		t.Error("BoldCyan should contain ANSI escape when enabled")
	}
	if !strings.Contains(BoldBlue("test"), "\033[1;34m") {
		t.Error("BoldBlue should contain ANSI escape when enabled")
	}

	// Test with color disabled
	enabled = false

	if Bold("test") != "test" {
		t.Error("Bold should return plain string when disabled")
	}
	if Green("test") != "test" {
		t.Error("Green should return plain string when disabled")
	}
	if Cyan("test") != "test" {
		t.Error("Cyan should return plain string when disabled")
	}
	if Dim("test") != "test" {
		t.Error("Dim should return plain string when disabled")
	}
}

func TestCheckmarkAndArrow(t *testing.T) {
	enabled = true
	if !strings.Contains(Checkmark(), "✓") {
		t.Error("Checkmark should contain ✓")
	}
	if !strings.Contains(Arrow(), "→") {
		t.Error("Arrow should contain →")
	}

	enabled = false
	if Checkmark() != "✓" {
		t.Error("Checkmark should return plain ✓ when disabled")
	}
	if Arrow() != "→" {
		t.Error("Arrow should return plain → when disabled")
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"plain text", "plain text"},
		{"\033[1mbold\033[0m", "bold"},
		{"\033[32mgreen\033[0m text", "green text"},
		{"\033[1;36mcyan\033[0m and \033[2mdim\033[0m", "cyan and dim"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := StripANSI(tt.input)
			if result != tt.expected {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDisableEnable(t *testing.T) {
	enabled = true
	Disable()
	if enabled {
		t.Error("Disable() should set enabled to false")
	}

	Enable()
	if !enabled {
		t.Error("Enable() should set enabled to true")
	}
}
