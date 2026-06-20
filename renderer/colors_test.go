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
	tests := []struct {
		id   string
		name string
	}{
		{"nord", "Nord"},
		{"nord-aurora", "Nord Aurora"},
		{"monokai", "Monokai"},
		{"solarized", "Solarized"},
		{"minimal", "Minimal"},
		{"dracula", "Dracula"},
		{"catppuccin", "Catppuccin"},
		{"gruvbox", "Gruvbox"},
		{"onedark", "One Dark"},
		{"tokyonight", "Tokyo Night"},
	}

	for _, tc := range tests {
		t.Run(tc.id, func(t *testing.T) {
			theme := GetPowerlineTheme(tc.id)
			if theme == nil {
				t.Fatalf("Expected '%s' theme to exist, got nil", tc.id)
			}
			if theme.Name != tc.name {
				t.Errorf("Expected %s theme name, got '%s'", tc.name, theme.Name)
			}
			if theme.Colors16 == nil {
				t.Errorf("Expected Colors16 defined for '%s'", tc.id)
			} else if len(theme.Colors16.Fg) != 5 || len(theme.Colors16.Bg) != 5 {
				t.Errorf("Expected 5 Colors16 levels for '%s', got Fg:%d, Bg:%d", tc.id, len(theme.Colors16.Fg), len(theme.Colors16.Bg))
			}

			if theme.Colors256 == nil {
				t.Errorf("Expected Colors256 defined for '%s'", tc.id)
			} else if len(theme.Colors256.Fg) != 5 || len(theme.Colors256.Bg) != 5 {
				t.Errorf("Expected 5 Colors256 levels for '%s', got Fg:%d, Bg:%d", tc.id, len(theme.Colors256.Fg), len(theme.Colors256.Bg))
			}

			if theme.Truecolor == nil {
				t.Errorf("Expected Truecolor defined for '%s'", tc.id)
			} else if len(theme.Truecolor.Fg) != 5 || len(theme.Truecolor.Bg) != 5 {
				t.Errorf("Expected 5 Truecolor levels for '%s', got Fg:%d, Bg:%d", tc.id, len(theme.Truecolor.Fg), len(theme.Truecolor.Bg))
			}
		})
	}
}
