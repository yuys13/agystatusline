package renderer

import (
	math_rand "math/rand"
	"strings"
	"testing"
)

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
				t.Errorf("Expected %q, got %q", tc.expected, actual)
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
				t.Errorf("For input %q, expected width %d, got %d", tc.input, tc.expected, actual)
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
		{"OSC8 link fitting within width", "\x1b]8;;https://example.com\x1b\\Hello\x1b]8;;\x1b\\", 8, "\x1b]8;;https://example.com\x1b\\Hello\x1b]8;;\x1b\\"},
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

func TestParseEscape_OSC8_TerminationAndEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOk      bool
		wantIsOsc8  bool
		wantIsClose bool
	}{
		{
			name:        "OSC8 opening link with ST termination",
			input:       "\x1b]8;;https://example.com\x1b\\",
			wantOk:      true,
			wantIsOsc8:  true,
			wantIsClose: false,
		},
		{
			name:        "OSC8 opening link with BEL termination",
			input:       "\x1b]8;;https://example.com\x07",
			wantOk:      true,
			wantIsOsc8:  true,
			wantIsClose: false,
		},
		{
			name:        "OSC8 closing link with ST termination",
			input:       "\x1b]8;;\x1b\\",
			wantOk:      true,
			wantIsOsc8:  true,
			wantIsClose: true,
		},
		{
			name:        "OSC8 closing link with BEL termination",
			input:       "\x1b]8;;\x07",
			wantOk:      true,
			wantIsOsc8:  true,
			wantIsClose: true,
		},
		{
			name:        "OSC8 with parameters (e.g. key-value id)",
			input:       "\x1b]8;id=link123:foo=bar;https://example.net\x1b\\",
			wantOk:      true,
			wantIsOsc8:  true,
			wantIsClose: false,
		},
		{
			name:        "OSC8 closing link with parameters",
			input:       "\x1b]8;id=link123;\x1b\\",
			wantOk:      true,
			wantIsOsc8:  true,
			wantIsClose: true,
		},
		{
			name:        "OSC8 URL ending with semicolon",
			input:       "\x1b]8;;https://example.com/api;\x1b\\",
			wantOk:      true,
			wantIsOsc8:  true,
			wantIsClose: false,
		},
		{
			name:        "Unterminated OSC sequence scanning",
			input:       "\x1b]8;;https://example.org",
			wantOk:      true,
			wantIsOsc8:  false,
			wantIsClose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, ok := parseEscape(tt.input, 0)
			if ok != tt.wantOk {
				t.Fatalf("parseEscape(%q, 0) ok = %v; want %v", tt.input, ok, tt.wantOk)
			}
			if ok {
				if res.isOsc8 != tt.wantIsOsc8 {
					t.Errorf("isOsc8 = %v; want %v", res.isOsc8, tt.wantIsOsc8)
				}
				if res.isClose != tt.wantIsClose {
					t.Errorf("isClose = %v; want %v", res.isClose, tt.wantIsClose)
				}
				if res.nextIndex != len(tt.input) {
					t.Errorf("nextIndex = %d; want %d", res.nextIndex, len(tt.input))
				}
			}
		})
	}
}

func TestGetVisibleWidth_Osc8AndCsiEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedWidth int
	}{
		{
			name:          "Standard OSC 8 link",
			input:         RenderOsc8Link("https://example.com", "Example"),
			expectedWidth: 7,
		},
		{
			name:          "OSC 8 link with multi-byte East Asian label",
			input:         RenderOsc8Link("https://example.org", "表示テスト"),
			expectedWidth: 10,
		},
		{
			name:          "OSC 8 link with CSI styled label",
			input:         RenderOsc8Link("https://example.net", "\x1b[31mRed\x1b[0m Link"),
			expectedWidth: 8,
		},
		{
			name:          "OSC 8 link with query string containing semicolon",
			input:         "\x1b]8;;https://example.com/api?q=1;v=2\x1b\\QueryLink\x1b]8;;\x1b\\",
			expectedWidth: 9,
		},
		{
			name:          "Colon-delimited truecolor SGR sequence",
			input:         "\x1b[38:2:255:100:50mColored\x1b[0m",
			expectedWidth: 7,
		},
		{
			name:          "Non-alphabet terminated CSI sequence",
			input:         "\x1b[1~HomeKey\x1b[200~",
			expectedWidth: 7,
		},
		{
			name:          "Two-byte escape sequence",
			input:         "\x1bMText\x1b7",
			expectedWidth: 4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := GetVisibleWidth(tc.input)
			if actual != tc.expectedWidth {
				t.Errorf("Expected visible width %d, got %d for %q", tc.expectedWidth, actual, tc.input)
			}
		})
	}
}

func TestTruncateStyledText_InvariantProperty(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "OSC 8 link with ST terminator",
			input: RenderOsc8Link("https://example.com/very/long/path", "LongHyperlinkLabelText"),
		},
		{
			name:  "OSC 8 link with BEL terminator",
			input: "\x1b]8;;https://example.org/path\x07LabelText\x1b]8;;\x07",
		},
		{
			name:  "OSC 8 link ending with semicolon",
			input: "\x1b]8;;https://example.net/api;\x1b\\SemicolonURLLinkText\x1b]8;;\x1b\\",
		},
		{
			name:  "East Asian multi-byte text",
			input: "あいうえお かきくけこ さしすせそ",
		},
		{
			name:  "Emoji and fullwidth characters",
			input: "🚀ステータスライン表示テスト😀😁😂",
		},
		{
			name:  "Complex CSI styling",
			input: "\x1b[38:2:255:128:64mTruecolor\x1b[0m \x1b[1;34mBoldBlue\x1b[0m \x1b[4mUnderline\x1b[0m",
		},
		{
			name:  "Mixed CSI and OSC 8 with multi-byte runes",
			input: "\x1b[1;34m\x1b]8;;https://example.com\x1b\\リンク: \x1b[32mOK\x1b[0m\x1b]8;;\x1b\\ \x1b[42m完了\x1b[0m",
		},
		{
			name:  "Unterminated OSC sequence",
			input: "\x1b]8;;https://example.org/unterminated",
		},
		{
			name:  "Unterminated CSI sequence",
			input: "\x1b[31mUnterminatedCSI",
		},
	}

	widthsToTest := []int{-100, -5, -1, 0, 1, 2, 3, 4, 5, 8, 10, 15, 20, 25, 30, 50, 100}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, w := range widthsToTest {
				truncated := TruncateStyledText(tc.input, w)
				visWidth := GetVisibleWidth(truncated)

				expectedMax := w
				if expectedMax < 0 {
					expectedMax = 0
				}

				if visWidth > expectedMax {
					t.Errorf("Invariant violation for max %d: got visible width %d (output: %q)", w, visWidth, truncated)
				}

				if w <= 0 && truncated != "" {
					t.Errorf("Expected empty string for w=%d, got %q", w, truncated)
				}

				if w > 0 && GetVisibleWidth(tc.input) <= w && truncated != tc.input {
					t.Errorf("Expected no-op for w=%d, input width %d <= max, but got %q", w, GetVisibleWidth(tc.input), truncated)
				}

				// Verify StripAnsi idempotency on output
				s1 := StripAnsi(truncated)
				s2 := StripAnsi(s1)
				if s1 != s2 {
					t.Errorf("StripAnsi idempotency violated for output %q: s1=%q, s2=%q", truncated, s1, s2)
				}
			}
		})
	}
}

func TestTruncateStyledText_Osc8HyperlinkClosure(t *testing.T) {
	link := RenderOsc8Link("https://example.com/very/long/path", "LongHyperlinkLabelText")
	// Visible width of "LongHyperlinkLabelText" is 22

	for w := 4; w < 22; w++ {
		truncated := TruncateStyledText(link, w)
		visWidth := GetVisibleWidth(truncated)

		if visWidth > w {
			t.Errorf("Width invariant failed for max %d: got visible width %d (%q)", w, visWidth, truncated)
		}

		if !strings.Contains(truncated, "\x1b]8;;\x1b\\...") && !strings.Contains(truncated, "\x1b]8;;\x07...") {
			t.Errorf("Truncated OSC8 link missing closing sequence before ellipsis for width %d: %q", w, truncated)
		}
	}
}

func TestProperty_TruncateStyledText_RandomGenerator(t *testing.T) {
	urls := []string{
		"https://example.com",
		"https://example.org/path?a=1;b=2",
		"https://example.net#hash",
		"https://example.com/api;",
	}
	csiStyles := []string{
		"\x1b[31m", "\x1b[1;32m", "\x1b[4m", "\x1b[38;5;200m", "\x1b[38:2:50:100:150m", "\x1b[0m", "\x1b[1~",
	}
	textSnippets := []string{
		"abc", "123", " ", "こんにちは", "世界", "🚀", "ステータス", "ERROR",
	}

	r := math_rand.New(math_rand.NewSource(12345))

	for i := 0; i < 5000; i++ {
		var sb strings.Builder
		numParts := r.Intn(8) + 1
		for j := 0; j < numParts; j++ {
			switch r.Intn(5) {
			case 0:
				sb.WriteString(textSnippets[r.Intn(len(textSnippets))])
			case 1:
				sb.WriteString(csiStyles[r.Intn(len(csiStyles))])
			case 2:
				url := urls[r.Intn(len(urls))]
				label := textSnippets[r.Intn(len(textSnippets))]
				sb.WriteString(RenderOsc8Link(url, label))
			case 3:
				sb.WriteString("\x1b]8;;https://example.com\x1b\\")
			case 4:
				sb.WriteString("\x1b]8;;\x07")
			}
		}
		inputText := sb.String()
		maxWidth := r.Intn(60) - 10 // range -10 to 49

		truncated := TruncateStyledText(inputText, maxWidth)
		actualWidth := GetVisibleWidth(truncated)

		expectedMax := maxWidth
		if expectedMax < 0 {
			expectedMax = 0
		}

		if actualWidth > expectedMax {
			t.Fatalf("[Iteration %d] Invariant Violation! Input: %q, maxWidth: %d -> got width %d > max %d (result: %q)",
				i, inputText, maxWidth, actualWidth, expectedMax, truncated)
		}

		if maxWidth > 0 && GetVisibleWidth(inputText) <= maxWidth && truncated != inputText {
			t.Fatalf("[Iteration %d] No-op Violation! Input width <= max, but text was modified. Input: %q, Truncated: %q",
				i, inputText, truncated)
		}

		s1 := StripAnsi(truncated)
		s2 := StripAnsi(s1)
		if s1 != s2 {
			t.Fatalf("[Iteration %d] StripAnsi Idempotency Violation! s1=%q, s2=%q", i, s1, s2)
		}
	}
}
