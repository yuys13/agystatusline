package renderer

import (
	"strings"
	"testing"
)

func TestGetColorAnsiCode(t *testing.T) {
	tests := []struct {
		colorName  string
		colorLevel string
		isBg       bool
		expected   string
	}{
		{"red", "ansi16", false, "\x1b[31m"},
		{"bgRed", "ansi16", true, "\x1b[41m"},
		{"brightRed", "ansi16", false, "\x1b[91m"},
		{"bgBrightRed", "ansi16", true, "\x1b[101m"},
		{"ansi256:160", "ansi256", false, "\x1b[38;5;160m"},
		{"hex:ff0000", "truecolor", false, "\x1b[38;2;255;0;0m"},
		{"hex:00ff00", "truecolor", true, "\x1b[48;2;0;255;0m"},
	}

	for _, tc := range tests {
		actual := GetColorAnsiCode(tc.colorName, tc.colorLevel, tc.isBg)
		if actual != tc.expected {
			t.Errorf("For (%s, %s, %t) expected ANSI '%q', got '%q'", tc.colorName, tc.colorLevel, tc.isBg, tc.expected, actual)
		}
	}
}

func TestApplyColors(t *testing.T) {
	text := "Test"
	// Bold and red foreground
	bold := true
	actual := ApplyColors(text, "red", "", &bold, "ansi16", nil)
	
	// Expect \x1b[1m (bold) + \x1b[31m (red) + Test + \x1b[39m (fg reset) + \x1b[22m (bold reset)
	// Order could depend on implementation, but let's verify containment and resets.
	if !strings.Contains(actual, "Test") {
		t.Errorf("Expected result to contain 'Test', got '%q'", actual)
	}
	if !strings.HasPrefix(actual, "\x1b[1m\x1b[31m") && !strings.HasPrefix(actual, "\x1b[31m\x1b[1m") {
		t.Errorf("Expected bold and red prefix, got '%q'", actual)
	}
	if !strings.HasSuffix(actual, "\x1b[39m\x1b[22m") && !strings.HasSuffix(actual, "\x1b[22m\x1b[39m") {
		t.Errorf("Expected bold and red reset suffixes, got '%q'", actual)
	}
}

func TestBgToFg(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"bgRed", "red"},
		{"bgBrightGreen", "brightGreen"},
		{"ansi256:123", "ansi256:123"},
		{"hex:ffffff", "hex:ffffff"},
	}

	for _, tc := range tests {
		actual := BgToFg(tc.input)
		if actual != tc.expected {
			t.Errorf("Expected BgToFg('%s') -> '%s', got '%s'", tc.input, tc.expected, actual)
		}
	}
}

func TestGetPowerlineTheme(t *testing.T) {
	theme := GetPowerlineTheme("nord")
	if theme == nil {
		t.Fatalf("Expected 'nord' theme to exist, got nil")
	}

	if theme.Name != "Nord" {
		t.Errorf("Expected Nord theme name, got '%s'", theme.Name)
	}

	colors256 := theme.Colors256
	if colors256 == nil {
		t.Fatalf("Expected Colors256 Nord definitions, got nil")
	}

	if len(colors256.Fg) == 0 || len(colors256.Bg) == 0 {
		t.Errorf("Expected Nord colors256 to have fg/bg entries")
	}
}
