package renderer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
	"github.com/yuys13/agystatusline/widgets"
)

func TestRenderStatusLines(t *testing.T) {
	widgets.RegisterAll()

	t.Run("normal mode", func(t *testing.T) {
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

		firstLine := StripAnsi(lines[0])
		if !strings.Contains(firstLine, "Claude 3.5 Sonnet") {
			t.Errorf("Expected 'Claude 3.5 Sonnet' in first line, got %q", firstLine)
		}
	})

	t.Run("auto separators", func(t *testing.T) {
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
			t.Errorf("Expected auto-separator formatting %q, got %q", expected, firstLine)
		}
	})

	t.Run("powerline mode", func(t *testing.T) {
		settings := types.DefaultSettings()
		settings.Powerline.Enabled = true
		settings.Powerline.Theme = "nord"
		settings.Lines = [][]types.WidgetItem{
			{
				{ID: "1", Type: "model"},
				{ID: "2", Type: "custom-text", CustomText: "12.5k"},
			},
		}

		ctx := types.RenderContext{
			Data: types.StatusJSON{
				Model: types.ModelInfo{
					ID: "Claude",
				},
			},
		}

		lines := RenderStatusLines(settings, ctx)
		firstLine := lines[0]

		if !strings.Contains(firstLine, "\uE0B0") {
			t.Errorf("Expected powerline separator '\\uE0B0' in output, got %q", firstLine)
		}

		if !strings.Contains(firstLine, "\x1b[") {
			t.Errorf("Expected ANSI color escapes in powerline output, got %q", firstLine)
		}
	})

	t.Run("powerline caps", func(t *testing.T) {
		settings := types.DefaultSettings()
		settings.Powerline.Enabled = true
		settings.Powerline.Theme = "nord"
		settings.Powerline.StartCaps = []string{"\uE0B2"}
		settings.Powerline.EndCaps = []string{"\uE0B0"}
		settings.Lines = [][]types.WidgetItem{
			{
				{ID: "1", Type: "model"},
				{ID: "2", Type: "custom-text", CustomText: "12.5k"},
			},
		}

		ctx := types.RenderContext{
			Data: types.StatusJSON{
				Model: types.ModelInfo{
					ID: "Claude",
				},
			},
		}

		lines := RenderStatusLines(settings, ctx)
		firstLine := lines[0]

		if !strings.Contains(firstLine, "\uE0B2") {
			t.Errorf("Expected start cap '\\uE0B2' in output, got %q", firstLine)
		}

		if !strings.Contains(firstLine, "\uE0B0") {
			t.Errorf("Expected end cap '\\uE0B0' in output, got %q", firstLine)
		}

		expectedStartCapFg := "\x1b[38;5;73m"
		if !strings.Contains(firstLine, expectedStartCapFg+"\uE0B2") {
			t.Errorf("Expected start cap to be colored with %q, but got %q", expectedStartCapFg, firstLine)
		}

		expectedEndCapFg := "\x1b[38;5;239m"
		if !strings.Contains(firstLine, expectedEndCapFg+"\uE0B0") {
			t.Errorf("Expected end cap to be colored with %q, but got %q", expectedEndCapFg, firstLine)
		}
	})

	t.Run("caps and non-ASCII separator", func(t *testing.T) {
		settings := types.DefaultSettings()
		settings.Powerline.Enabled = false
		settings.DefaultSeparator = " - "
		settings.Lines = [][]types.WidgetItem{
			{
				{ID: "1", Type: "custom-text", CustomText: "Left"},
				{ID: "2", Type: "custom-text", CustomText: "Right"},
			},
		}

		ctx := types.RenderContext{Data: types.StatusJSON{}}
		lines := RenderStatusLines(settings, ctx)

		if len(lines) == 0 {
			t.Fatalf("Expected lines")
		}

		stripped := StripAnsi(lines[0])
		if !strings.Contains(stripped, "Left - Right") {
			t.Errorf("Expected 'Left - Right' in stripped line, got %q", stripped)
		}
	})

	t.Run("powerline color levels", func(t *testing.T) {
		ctx := types.RenderContext{Data: types.StatusJSON{}}

		for _, colorLevel := range []int{0, 2, 3} {
			t.Run(fmt.Sprintf("colorLevel_%d", colorLevel), func(t *testing.T) {
				settings := types.DefaultSettings()
				settings.Powerline.Enabled = true
				settings.Powerline.Theme = "dracula"
				settings.ColorLevel = colorLevel
				settings.Lines = [][]types.WidgetItem{
					{
						{ID: "1", Type: "custom-text", CustomText: "P1"},
						{ID: "2", Type: "custom-text", CustomText: "P2"},
					},
				}

				lines := RenderStatusLines(settings, ctx)
				if len(lines) == 0 {
					t.Fatalf("Expected rendered powerline lines for color level %d", colorLevel)
				}
				if !strings.Contains(lines[0], "P1") || !strings.Contains(lines[0], "P2") {
					t.Errorf("Expected P1 and P2 in powerline output for color level %d", colorLevel)
				}
			})
		}
	})
}
