package renderer

import (
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
	"github.com/yuys13/agystatusline/widgets"
)

func TestRenderStatusLines_NormalMode(t *testing.T) {
	// Register widgets
	widgets.RegisterAll()

	settings := types.DefaultSettings()
	settings.Powerline.Enabled = false
	settings.DefaultSeparator = "|"
	settings.DefaultPadding = " "

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID:          "claude-3-5-sonnet",
				DisplayName: "Claude 3.5 Sonnet",
			},
		},
	}

	lines := RenderStatusLines(settings, ctx)

	if len(lines) == 0 {
		t.Fatalf("Expected rendered lines, got none")
	}

	// Model widget output should be present on first line
	firstLine := StripAnsi(lines[0])
	if !strings.Contains(firstLine, "Claude 3.5 Sonnet") {
		t.Errorf("Expected 'Claude 3.5 Sonnet' in first line, got '%s'", firstLine)
	}

	// Separator should be applied (in between active widgets, but here we only have model active)
	// Let's add a test for auto-separators instead
}

func TestRenderStatusLines_AutoSeparators(t *testing.T) {
	widgets.RegisterAll()

	settings := types.DefaultSettings()
	settings.Powerline.Enabled = false
	settings.DefaultSeparator = "/"
	settings.DefaultPadding = ""
	settings.Lines = [][]types.WidgetItem{
		{
			{ID: "1", Type: "custom-text", CustomText: "A"},
			{ID: "2", Type: "custom-text", CustomText: "B"},
			{ID: "3", Type: "custom-text", CustomText: "C"},
		},
	}

	ctx := types.RenderContext{
		Data: types.StatusJSON{},
	}

	lines := RenderStatusLines(settings, ctx)
	if len(lines) == 0 {
		t.Fatalf("Expected lines")
	}

	firstLine := StripAnsi(lines[0])
	expected := "A/B/C"
	if firstLine != expected {
		t.Errorf("Expected auto-separator formatting '%s', got '%s'", expected, firstLine)
	}
}

func TestRenderStatusLines_PowerlineMode(t *testing.T) {
	widgets.RegisterAll()

	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	settings.Powerline.Theme = "nord"
	settings.Lines = [][]types.WidgetItem{
		{
			{ID: "1", Type: "model"},
			{ID: "2", Type: "context-length"},
		},
	}

	inputTokens := float64(1000)
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID: "Claude",
			},
			ContextWindow: &types.ContextWindowInfo{
				TotalInputTokens: &inputTokens,
			},
		},
	}

	lines := RenderStatusLines(settings, ctx)
	firstLine := lines[0]

	if !strings.Contains(firstLine, "\uE0B0") {
		t.Errorf("Expected powerline separator '\\uE0B0' in output, got '%q'", firstLine)
	}

	if !strings.Contains(firstLine, "\x1b[") {
		t.Errorf("Expected ANSI color escapes in powerline output, got '%q'", firstLine)
	}
}

func TestRenderStatusLines_PowerlineCaps(t *testing.T) {
	widgets.RegisterAll()

	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	settings.Powerline.Theme = "nord"
	settings.Powerline.StartCaps = []string{"\uE0B2"}
	settings.Powerline.EndCaps = []string{"\uE0B0"}
	settings.Lines = [][]types.WidgetItem{
		{
			{ID: "1", Type: "model"},
			{ID: "2", Type: "context-length"},
		},
	}

	inputTokens := float64(1000)
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID: "Claude",
			},
			ContextWindow: &types.ContextWindowInfo{
				TotalInputTokens: &inputTokens,
			},
		},
	}

	lines := RenderStatusLines(settings, ctx)
	firstLine := lines[0]

	// Start cap (\uE0B2) should be present at the beginning of the rendered string
	if !strings.Contains(firstLine, "\uE0B2") {
		t.Errorf("Expected start cap '\\uE0B2' in output, got '%q'", firstLine)
	}

	// End cap (\uE0B0) should be present at the end of the rendered string
	if !strings.Contains(firstLine, "\uE0B0") {
		t.Errorf("Expected end cap '\\uE0B0' in output, got '%q'", firstLine)
	}

	// Start cap should use first widget's background color as foreground
	// Since colorLevel is 2 (256-color) by default, the first widget background is ansi256:73
	expectedStartCapFg := "\x1b[38;5;73m"
	if !strings.Contains(firstLine, expectedStartCapFg+"\uE0B2") {
		t.Errorf("Expected start cap to be colored with %q, but got '%q'", expectedStartCapFg, firstLine)
	}

	// End cap should use last widget's background color as foreground
	// In Nord, the second widget (context-length) is index 1 -> ansi256:239
	expectedEndCapFg := "\x1b[38;5;239m"
	if !strings.Contains(firstLine, expectedEndCapFg+"\uE0B0") {
		t.Errorf("Expected end cap to be colored with %q, but got '%q'", expectedEndCapFg, firstLine)
	}
}
