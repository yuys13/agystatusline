package renderer

import "testing"

func TestStripAnsi(t *testing.T) {
	input := "\x1b[31mHello\x1b[0m \x1b[1mWorld\x1b[22m"
	expected := "Hello World"
	actual := StripAnsi(input)
	if actual != expected {
		t.Errorf("Expected '%s', got '%s'", expected, actual)
	}
}

func TestGetVisibleWidth(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"Hello", 5},
		{"\x1b[31mHello\x1b[0m", 5},
		{"こんにちは", 10}, // East Asian width (5 characters * 2)
		{"\x1b[1mこんにちは\x1b[22m", 10},
		{"Hello こんにちは", 16},
	}

	for _, tc := range tests {
		actual := GetVisibleWidth(tc.input)
		if actual != tc.expected {
			t.Errorf("For input '%s', expected width %d, got %d", tc.input, tc.expected, actual)
		}
	}
}

func TestRenderOsc8Link(t *testing.T) {
	url := "https://github.com"
	text := "GitHub"
	expected := "\x1b]8;;" + url + "\x1b\\" + text + "\x1b]8;;\x1b\\"
	actual := RenderOsc8Link(url, text)
	if actual != expected {
		t.Errorf("Expected '%q', got '%q'", expected, actual)
	}
}

func TestTruncateStyledText(t *testing.T) {
	tests := []struct {
		input    string
		width    int
		expected string
	}{
		{"Hello World", 8, "Hello..."},
		{"\x1b[31mHello\x1b[0m World", 8, "\x1b[31mHello\x1b[0m..."},
		{"\x1b]8;;http://x.com\x1b\\Hello\x1b]8;;\x1b\\", 8, "\x1b]8;;http://x.com\x1b\\Hello\x1b]8;;\x1b\\"}, // fits
	}

	for _, tc := range tests {
		actual := TruncateStyledText(tc.input, tc.width)
		if actual != tc.expected {
			t.Errorf("For input '%q' with width %d, expected '%q', got '%q'", tc.input, tc.width, tc.expected, actual)
		}
	}
}
