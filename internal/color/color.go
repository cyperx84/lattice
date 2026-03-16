package color

import (
	"os"
	"strings"
)

var enabled = true

func init() {
	// Respect NO_COLOR standard (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		enabled = false
	}
}

// Disable programmatically (e.g., when --json flag is set or stdout is not a terminal)
func Disable() {
	enabled = false
}

// Enable programmatically
func Enable() {
	enabled = true
}

// Enabled returns whether color output is enabled
func Enabled() bool {
	return enabled
}

// Bold returns the string in bold
func Bold(s string) string {
	if !enabled {
		return s
	}
	return "\033[1m" + s + "\033[0m"
}

// Dim returns the string dimmed
func Dim(s string) string {
	if !enabled {
		return s
	}
	return "\033[2m" + s + "\033[0m"
}

// Green returns the string in green
func Green(s string) string {
	if !enabled {
		return s
	}
	return "\033[32m" + s + "\033[0m"
}

// Yellow returns the string in yellow
func Yellow(s string) string {
	if !enabled {
		return s
	}
	return "\033[33m" + s + "\033[0m"
}

// Blue returns the string in blue
func Blue(s string) string {
	if !enabled {
		return s
	}
	return "\033[34m" + s + "\033[0m"
}

// Cyan returns the string in cyan
func Cyan(s string) string {
	if !enabled {
		return s
	}
	return "\033[36m" + s + "\033[0m"
}

// BoldCyan returns the string in bold cyan
func BoldCyan(s string) string {
	if !enabled {
		return s
	}
	return "\033[1;36m" + s + "\033[0m"
}

// BoldBlue returns the string in bold blue
func BoldBlue(s string) string {
	if !enabled {
		return s
	}
	return "\033[1;34m" + s + "\033[0m"
}

// Checkmark returns a green checkmark (or plain if color disabled)
func Checkmark() string {
	if enabled {
		return "\033[32m✓\033[0m"
	}
	return "✓"
}

// Arrow returns a green arrow (or plain if color disabled)
func Arrow() string {
	if enabled {
		return "\033[32m→\033[0m"
	}
	return "→"
}

// StripANSI removes ANSI escape codes from a string
func StripANSI(s string) string {
	// Simple approach: remove common ANSI sequences
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}
