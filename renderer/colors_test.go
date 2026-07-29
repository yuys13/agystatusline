package renderer

import (
	"strings"
	"testing"
)

func TestGetColorAnsiCode(t *testing.T) {
	tests := []struct {
		name       string
		colorName  string
		colorLevel string
		isBg       bool
		expected   string
	}{
		{"red foreground ansi16", "red", "ansi16", false, "\x1b[31m"},
		{"red background ansi16", "bgRed", "ansi16", true, "\x1b[41m"},
		{"brightRed foreground ansi16", "brightRed", "ansi16", false, "\x1b[91m"},
		{"brightRed background ansi16", "bgBrightRed", "ansi16", true, "\x1b[101m"},
		{"ansi256 foreground", "ansi256:160", "ansi256", false, "\x1b[38;5;160m"},
		{"hex truecolor foreground", "hex:ff0000", "truecolor", false, "\x1b[38;2;255;0;0m"},
		{"hex truecolor background", "hex:00ff00", "truecolor", true, "\x1b[48;2;0;255;0m"},
		{"invalid hex string", "hex:invalid", "truecolor", false, ""},
		{"short hex string", "hex:12", "truecolor", false, ""},
		{"non-hex characters", "hex:zzzzzz", "truecolor", false, ""},
		{"unknown color name", "unknown_color", "truecolor", false, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := GetColorAnsiCode(tc.colorName, tc.colorLevel, tc.isBg)
			if actual != tc.expected {
				t.Errorf("For (%q, %q, %t) expected ANSI %q, got %q", tc.colorName, tc.colorLevel, tc.isBg, tc.expected, actual)
			}
		})
	}
}

func TestApplyColors(t *testing.T) {
	t.Run("bold and red foreground", func(t *testing.T) {
		text := "Test"
		bold := true
		actual := ApplyColors(text, "red", "", &bold, "ansi16", nil)

		if !strings.Contains(actual, "Test") {
			t.Errorf("Expected result to contain 'Test', got %q", actual)
		}
		if !strings.HasPrefix(actual, "\x1b[1m\x1b[31m") && !strings.HasPrefix(actual, "\x1b[31m\x1b[1m") {
			t.Errorf("Expected bold and red prefix, got %q", actual)
		}
		if !strings.HasSuffix(actual, "\x1b[39m\x1b[22m") && !strings.HasSuffix(actual, "\x1b[22m\x1b[39m") {
			t.Errorf("Expected bold and red reset suffixes, got %q", actual)
		}
	})

	t.Run("invalid gradient fallback", func(t *testing.T) {
		res := ApplyColors("Text", "gradient:hex:invalid,hex:0000FF", "ansi16", nil, "truecolor", nil)
		if res == "" {
			t.Errorf("Expected ApplyColors to handle invalid gradient gracefully, got empty string")
		}
	})
}

func TestBgToFg(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"bgRed to red", "bgRed", "red"},
		{"bgBrightGreen to brightGreen", "bgBrightGreen", "brightGreen"},
		{"ansi256 unchanged", "ansi256:123", "ansi256:123"},
		{"hex unchanged", "hex:ffffff", "hex:ffffff"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := BgToFg(tc.input)
			if actual != tc.expected {
				t.Errorf("Expected BgToFg(%q) -> %q, got %q", tc.input, tc.expected, actual)
			}
		})
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
				t.Fatalf("Expected %q theme to exist, got nil", tc.id)
			}
			if theme.Name != tc.name {
				t.Errorf("Expected %q theme name, got %q", tc.name, theme.Name)
			}
			if theme.Colors16 == nil {
				t.Errorf("Expected Colors16 defined for %q", tc.id)
			} else if len(theme.Colors16.Fg) != 5 || len(theme.Colors16.Bg) != 5 {
				t.Errorf("Expected 5 Colors16 levels for %q, got Fg:%d, Bg:%d", tc.id, len(theme.Colors16.Fg), len(theme.Colors16.Bg))
			}

			if theme.Colors256 == nil {
				t.Errorf("Expected Colors256 defined for %q", tc.id)
			} else if len(theme.Colors256.Fg) != 5 || len(theme.Colors256.Bg) != 5 {
				t.Errorf("Expected 5 Colors256 levels for %q, got Fg:%d, Bg:%d", tc.id, len(theme.Colors256.Fg), len(theme.Colors256.Bg))
			}

			if theme.Truecolor == nil {
				t.Errorf("Expected Truecolor defined for %q", tc.id)
			} else if len(theme.Truecolor.Fg) != 5 || len(theme.Truecolor.Bg) != 5 {
				t.Errorf("Expected 5 Truecolor levels for %q, got Fg:%d, Bg:%d", tc.id, len(theme.Truecolor.Fg), len(theme.Truecolor.Bg))
			}
		})
	}

	t.Run("nonexistent fallback", func(t *testing.T) {
		theme := GetPowerlineTheme("nonexistent-theme-name")
		if theme != nil {
			t.Errorf("Expected nil for nonexistent powerline theme, got %v", theme)
		}
	})
}

func TestApplyGradientToText_EdgeCases(t *testing.T) {
	stops := []RGB{{R: 255, G: 0, B: 0}, {R: 0, G: 255, B: 0}}

	tests := []struct {
		name   string
		input  string
		verify func(t *testing.T, result string)
	}{
		{
			name:  "empty text",
			input: "",
			verify: func(t *testing.T, res string) {
				if res != "" {
					t.Errorf("Expected empty string for empty text gradient, got %q", res)
				}
			},
		},
		{
			name:  "single character text",
			input: "X",
			verify: func(t *testing.T, res string) {
				if !strings.Contains(res, "X") {
					t.Errorf("Expected result to contain 'X', got %q", res)
				}
			},
		},
		{
			name:  "text with ANSI escape sequences",
			input: "BoldText",
			verify: func(t *testing.T, res string) {
				if !strings.Contains(res, "B") || !strings.Contains(res, "t") {
					t.Errorf("Expected result to contain characters from 'BoldText', got %q", res)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := applyGradientToText(tc.input, stops, "truecolor")
			tc.verify(t, res)
		})
	}
}

func TestWrapSolidColor(t *testing.T) {
	c := RGB{R: 255, G: 128, B: 0}

	tests := []struct {
		name       string
		colorLevel string
		verify     func(t *testing.T, result string)
	}{
		{
			name:       "truecolor mode",
			colorLevel: "truecolor",
			verify: func(t *testing.T, res string) {
				if !strings.Contains(res, "255;128;0") || !strings.Contains(res, "test") {
					t.Errorf("Expected truecolor wrap, got %q", res)
				}
			},
		},
		{
			name:       "ansi256 mode",
			colorLevel: "ansi256",
			verify: func(t *testing.T, res string) {
				if !strings.Contains(res, "\x1b[38;5;") || !strings.Contains(res, "test") {
					t.Errorf("Expected ansi256 wrap, got %q", res)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := wrapSolidColor("test", c, tc.colorLevel)
			tc.verify(t, res)
		})
	}
}
