package renderer

import "testing"

func TestStripAnsi(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "red and bold styled text",
			input:    "\x1b[31mHello\x1b[0m \x1b[1mWorld\x1b[22m",
			expected: "Hello World",
		},
		{
			name:     "plain text without ANSI escape sequences",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := StripAnsi(tc.input)
			if actual != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, actual)
			}
		})
	}
}

func TestGetVisibleWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"ASCII text", "Hello", 5},
		{"ANSI styled text", "\x1b[31mHello\x1b[0m", 5},
		{"East Asian full-width text", "こんにちは", 10},
		{"ANSI styled East Asian text", "\x1b[1mこんにちは\x1b[22m", 10},
		{"Mixed ASCII and East Asian text", "Hello こんにちは", 16},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := GetVisibleWidth(tc.input)
			if actual != tc.expected {
				t.Errorf("For input '%s', expected width %d, got %d", tc.input, tc.expected, actual)
			}
		})
	}
}

func TestRenderOsc8Link(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		text     string
		expected string
	}{
		{
			name:     "standard link",
			url:      "https://example.com",
			text:     "Example",
			expected: "\x1b]8;;https://example.com\x1b\\Example\x1b]8;;\x1b\\",
		},
		{
			name:     "empty text",
			url:      "https://example.org",
			text:     "",
			expected: "\x1b]8;;https://example.org\x1b\\\x1b]8;;\x1b\\",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := RenderOsc8Link(tc.url, tc.text)
			if actual != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, actual)
			}
		})
	}
}

func TestTruncateStyledText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{"plain text truncation", "Hello World", 8, "Hello..."},
		{"ANSI styled text truncation", "\x1b[31mHello\x1b[0m World", 8, "\x1b[31mHello\x1b[0m..."},
		{"OSC8 link fitting within width", "\x1b]8;;http://example.com\x1b\\Hello\x1b]8;;\x1b\\", 8, "\x1b]8;;http://example.com\x1b\\Hello\x1b]8;;\x1b\\"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := TruncateStyledText(tc.input, tc.width)
			if actual != tc.expected {
				t.Errorf("For input %q with width %d, expected %q, got %q", tc.input, tc.width, tc.expected, actual)
			}
		})
	}
}
