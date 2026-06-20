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
	if !strings.Contains(firstLine, "Model: Claude 3.5 Sonnet") {
		t.Errorf("Expected 'Model: Claude 3.5 Sonnet' in first line, got '%s'", firstLine)
	}

	// Separator should be applied
	if !strings.Contains(firstLine, " | ") {
		t.Errorf("Expected separator ' | ', got '%s'", firstLine)
	}
}

func TestRenderStatusLines_FlexSeparator(t *testing.T) {
	widgets.RegisterAll()

	settings := types.DefaultSettings()
	settings.Powerline.Enabled = false
	settings.DefaultSeparator = " "
	settings.DefaultPadding = ""
	
	// Set custom lines with a flex-separator: [model, flex-separator, context-length]
	settings.Lines = [][]types.WidgetItem{
		{
			{ID: "1", Type: "model"},
			{ID: "2", Type: "flex-separator"},
			{ID: "3", Type: "context-length"},
		},
	}

	width := 40
	inputTokens := float64(1000)
	ctx := types.RenderContext{
		TerminalWidth: &width,
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
	firstLine := StripAnsi(lines[0])

	// Expected output: "Model: Claude                     1.0k"
	// Output length should be exactly 40 (since terminal width is 40)
	if len(firstLine) != 40 {
		t.Errorf("Expected line length 40, got %d ('%s')", len(firstLine), firstLine)
	}

	if !strings.HasPrefix(firstLine, "Model: Claude") || !strings.HasSuffix(firstLine, "1.0k") {
		t.Errorf("Expected flex-separator placement, got '%s'", firstLine)
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
