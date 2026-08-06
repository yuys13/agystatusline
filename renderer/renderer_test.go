package renderer

import (
	"fmt"
	"strings"
	"sync"
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

func TestGetPowerlineTheme_LookupsAndFallback(t *testing.T) {
	settings := types.Settings{
		Powerline: types.PowerlineConfig{
			Enabled: true,
			Theme:   "nord",
		},
		Lines: [][]types.WidgetItem{
			{
				{Type: "model", Color: "red"},
			},
		},
	}

	nilTheme := GetPowerlineTheme("non_existent_theme_name_12345")
	if nilTheme != nil {
		t.Errorf("Expected nil for non-existent theme, got %v", nilTheme)
	}

	ctx := types.RenderContext{}
	lines := RenderStatusLines(settings, ctx)
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}
}

func TestRenderStatusLines_ExtremeTerminalWidths(t *testing.T) {
	settings := types.Settings{
		DefaultPadding:   " ",
		DefaultSeparator: " | ",
		Lines: [][]types.WidgetItem{
			{
				{Type: "custom", CustomText: "Long Custom Statusline Text For Testing Truncation Bounds"},
			},
		},
	}

	termWidths := []int{-100, -1, 0, 1, 2, 3, 4, 5, 6, 7, 10, 80, 10000}
	for _, tw := range termWidths {
		for _, isPreview := range []bool{false, true} {
			widthVal := tw
			ctx := types.RenderContext{
				TerminalWidth: &widthVal,
				IsPreview:     isPreview,
			}
			lines := RenderStatusLines(settings, ctx)
			if len(lines) != 1 {
				t.Errorf("Expected 1 line for termWidth %d preview=%v, got %d", tw, isPreview, len(lines))
			}
		}
	}
}

func TestRenderStatusLines_EmptyLinesAndNilPointers(t *testing.T) {
	settings := types.Settings{
		Lines: [][]types.WidgetItem{
			{}, // empty line
			{
				{Type: "invalid_widget_type_123"},
				{Type: "separator"},
				{Type: "flex-separator"},
			},
		},
	}

	ctx := types.RenderContext{}
	lines := RenderStatusLines(settings, ctx)
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "" {
		t.Errorf("Expected empty line for empty item slice, got %q", lines[0])
	}
}

func TestRenderer_ConcurrentRenders(t *testing.T) {
	settings := types.Settings{
		DefaultPadding:   " ",
		DefaultSeparator: " | ",
		ColorLevel:       3,
		GlobalBold:       true,
		Powerline: types.PowerlineConfig{
			Enabled:    true,
			Theme:      "nord",
			Separators: []string{"\uE0B0"},
			StartCaps:  []string{"\uE0B2"},
			EndCaps:    []string{"\uE0B0"},
		},
		Lines: [][]types.WidgetItem{
			{
				{Type: "model", Color: "gradient:red,blue", BackgroundColor: "bgBlack"},
				{Type: "context-tokens", Dim: "parens"},
				{Type: "custom", CustomText: RenderOsc8Link("https://example.com", "TestLink")},
			},
		},
	}

	tw := 80
	ctx := types.RenderContext{
		TerminalWidth: &tw,
		IsPreview:     true,
	}

	var wg sync.WaitGroup
	numGoroutines := 20
	iterationsPerGoroutine := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterationsPerGoroutine; j++ {
				lines := RenderStatusLines(settings, ctx)
				if len(lines) == 0 {
					t.Errorf("Concurrent RenderStatusLines returned empty output")
				}
				trunc := TruncateStyledText("Some \x1b[31mstyled\x1b[0m text with \x1b]8;;https://example.com\x1b\\link\x1b]8;;\x1b\\", 20)
				if trunc == "" {
					t.Errorf("Concurrent TruncateStyledText returned empty output")
				}
				code := GetColorAnsiCode("gradient:red,green,blue", "truecolor", false)
				if code == "" {
					t.Errorf("Concurrent GetColorAnsiCode returned empty code")
				}
				app := ApplyColors("Sample", "hex:123456", "bgRed", nil, "ansi256", "parens")
				if app == "" {
					t.Errorf("Concurrent ApplyColors returned empty output")
				}
			}
		}()
	}

	wg.Wait()
}

func TestRenderer_HugeInputs(t *testing.T) {
	hugeString := strings.Repeat("A\x1b[31mB\x1b[0mC ", 2000) // ~28KB
	stripped := StripAnsi(hugeString)
	if len(stripped) == 0 {
		t.Errorf("Expected non-empty stripped output for huge input")
	}

	w := GetVisibleWidth(hugeString)
	if w <= 0 {
		t.Errorf("Expected positive visible width for huge input, got %d", w)
	}

	truncated := TruncateStyledText(hugeString, 500)
	if GetVisibleWidth(truncated) > 500 {
		t.Errorf("Truncated width %d exceeds max 500", GetVisibleWidth(truncated))
	}
}
