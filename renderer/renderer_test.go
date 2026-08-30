package renderer

import (
	"strings"
	"sync"
	"testing"

	"github.com/yuys13/agystatusline/types"
	"github.com/yuys13/agystatusline/widgets"
)

func init() {
	widgets.RegisterAll()
}

func TestRenderStatusLines_StandardNonPowerline(t *testing.T) {
	tests := []struct {
		name       string
		settings   types.Settings
		ctx        types.RenderContext
		assertFunc func(t *testing.T, lines []string)
	}{
		{
			name: "default model rendering",
			settings: types.Settings{
				General: types.GeneralConfig{
					Separator: " · ",
				},
				Lines: [][]types.WidgetItem{
					{
						{Type: "model"},
					},
				},
			},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Model: types.ModelInfo{
						DisplayName: "Claude 3.5 Sonnet",
					},
				},
			},
			assertFunc: func(t *testing.T, lines []string) {
				if len(lines) != 1 {
					t.Fatalf("expected 1 line, got %d", len(lines))
				}
				stripped := StripAnsi(lines[0])
				if !strings.Contains(stripped, "Claude 3.5 Sonnet") {
					t.Errorf("expected 'Claude 3.5 Sonnet' in output, got %q", stripped)
				}
			},
		},
		{
			name: "title and body styling",
			settings: types.Settings{
				General: types.GeneralConfig{
					ColorLevel: 1,
				},
				Lines: [][]types.WidgetItem{
					{
						{Type: "sandbox"},
					},
				},
			},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Sandbox: &types.SandboxInfo{
						Enabled: new(true),
					},
				},
			},
			assertFunc: func(t *testing.T, lines []string) {
				if len(lines) != 1 {
					t.Fatalf("expected 1 line, got %d", len(lines))
				}
				// Sandbox widget returns ("sandbox", "on")
				// Title "sandbox" should be colored with brightBlack (\x1b[90m)
				// Body "on" should be colored with brightGreen (\x1b[92m)
				line := lines[0]
				if !strings.Contains(line, "\x1b[90msandbox\x1b[39m") {
					t.Errorf("expected title 'sandbox' in brightBlack, got %q", line)
				}
				if !strings.Contains(line, "\x1b[92mon\x1b[39m") {
					t.Errorf("expected body 'on' in brightGreen, got %q", line)
				}
			},
		},
		{
			name: "custom color override",
			settings: types.Settings{
				General: types.GeneralConfig{
					ColorLevel: 1,
				},
				Lines: [][]types.WidgetItem{
					{
						{Type: "custom-text", Text: "Hello", Color: "red"},
					},
				},
			},
			ctx: types.RenderContext{},
			assertFunc: func(t *testing.T, lines []string) {
				if len(lines) != 1 {
					t.Fatalf("expected 1 line, got %d", len(lines))
				}
				line := lines[0]
				if !strings.Contains(line, "\x1b[31mHello\x1b[39m") {
					t.Errorf("expected custom color red (\x1b[31m) for Hello, got %q", line)
				}
			},
		},
		{
			name: "item raw mode omits title",
			settings: types.Settings{
				Lines: [][]types.WidgetItem{
					{
						{Type: "sandbox", Raw: true},
					},
				},
			},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Sandbox: &types.SandboxInfo{
						Enabled: new(true),
					},
				},
			},
			assertFunc: func(t *testing.T, lines []string) {
				if len(lines) != 1 {
					t.Fatalf("expected 1 line, got %d", len(lines))
				}
				stripped := StripAnsi(lines[0])
				if stripped != "on" {
					t.Errorf("expected raw output 'on', got %q", stripped)
				}
			},
		},
		{
			name: "settings minimalist mode omits title",
			settings: types.Settings{
				General: types.GeneralConfig{
					Minimalist: true,
				},
				Lines: [][]types.WidgetItem{
					{
						{Type: "sandbox"},
						{Type: "artifacts"},
					},
				},
			},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Sandbox: &types.SandboxInfo{
						Enabled: new(true),
					},
					ArtifactCount: new(5),
				},
			},
			assertFunc: func(t *testing.T, lines []string) {
				if len(lines) != 1 {
					t.Fatalf("expected 1 line, got %d", len(lines))
				}
				stripped := StripAnsi(lines[0])
				if strings.Contains(stripped, "sandbox") || strings.Contains(stripped, "artifacts") {
					t.Errorf("expected titles to be omitted in minimalist mode, got %q", stripped)
				}
				if !strings.Contains(stripped, "on") || !strings.Contains(stripped, "5") {
					t.Errorf("expected bodies 'on' and '5' in output, got %q", stripped)
				}
			},
		},
		{
			name: "context minimalist mode omits title",
			settings: types.Settings{
				Lines: [][]types.WidgetItem{
					{
						{Type: "sandbox"},
					},
				},
			},
			ctx: types.RenderContext{
				Minimalist: true,
				Data: types.StatusJSON{
					Sandbox: &types.SandboxInfo{
						Enabled: new(true),
					},
				},
			},
			assertFunc: func(t *testing.T, lines []string) {
				if len(lines) != 1 {
					t.Fatalf("expected 1 line, got %d", len(lines))
				}
				stripped := StripAnsi(lines[0])
				if stripped != "on" {
					t.Errorf("expected minimalist output 'on', got %q", stripped)
				}
			},
		},
		{
			name: "custom separator and padding",
			settings: types.Settings{
				General: types.GeneralConfig{
					Separator: "|",
					Padding:   " ",
				},
				Lines: [][]types.WidgetItem{
					{
						{Type: "custom-text", Text: "A"},
						{Type: "custom-text", Text: "B"},
						{Type: "custom-text", Text: "C"},
					},
				},
			},
			ctx: types.RenderContext{},
			assertFunc: func(t *testing.T, lines []string) {
				if len(lines) != 1 {
					t.Fatalf("expected 1 line, got %d", len(lines))
				}
				stripped := StripAnsi(lines[0])
				expected := " A  |  B  |  C "
				if stripped != expected {
					t.Errorf("expected %q, got %q", expected, stripped)
				}
			},
		},
		{
			name: "comma separator formatting",
			settings: types.Settings{
				General: types.GeneralConfig{
					Separator: ",",
				},
				Lines: [][]types.WidgetItem{
					{
						{Type: "custom-text", Text: "X"},
						{Type: "custom-text", Text: "Y"},
					},
				},
			},
			ctx: types.RenderContext{},
			assertFunc: func(t *testing.T, lines []string) {
				if len(lines) != 1 {
					t.Fatalf("expected 1 line, got %d", len(lines))
				}
				stripped := StripAnsi(lines[0])
				expected := "X, Y"
				if stripped != expected {
					t.Errorf("expected %q, got %q", expected, stripped)
				}
			},
		},
		{
			name: "dash separator formatting",
			settings: types.Settings{
				General: types.GeneralConfig{
					Separator: "-",
				},
				Lines: [][]types.WidgetItem{
					{
						{Type: "custom-text", Text: "X"},
						{Type: "custom-text", Text: "Y"},
					},
				},
			},
			ctx: types.RenderContext{},
			assertFunc: func(t *testing.T, lines []string) {
				if len(lines) != 1 {
					t.Fatalf("expected 1 line, got %d", len(lines))
				}
				stripped := StripAnsi(lines[0])
				expected := "X - Y"
				if stripped != expected {
					t.Errorf("expected %q, got %q", expected, stripped)
				}
			},
		},
		{
			name: "space separator formatting",
			settings: types.Settings{
				General: types.GeneralConfig{
					Separator: " ",
				},
				Lines: [][]types.WidgetItem{
					{
						{Type: "custom-text", Text: "X"},
						{Type: "custom-text", Text: "Y"},
					},
				},
			},
			ctx: types.RenderContext{},
			assertFunc: func(t *testing.T, lines []string) {
				if len(lines) != 1 {
					t.Fatalf("expected 1 line, got %d", len(lines))
				}
				stripped := StripAnsi(lines[0])
				expected := "X Y"
				if stripped != expected {
					t.Errorf("expected %q, got %q", expected, stripped)
				}
			},
		},
		{
			name: "default middle dot separator when unset",
			settings: types.Settings{
				General: types.GeneralConfig{
					Separator: "",
				},
				Lines: [][]types.WidgetItem{
					{
						{Type: "custom-text", Text: "X"},
						{Type: "custom-text", Text: "Y"},
					},
				},
			},
			ctx: types.RenderContext{},
			assertFunc: func(t *testing.T, lines []string) {
				if len(lines) != 1 {
					t.Fatalf("expected 1 line, got %d", len(lines))
				}
				stripped := StripAnsi(lines[0])
				expected := "X · Y"
				if stripped != expected {
					t.Errorf("expected %q, got %q", expected, stripped)
				}
			},
		},
		{
			name: "multiple lines with empty line handling",
			settings: types.Settings{
				Lines: [][]types.WidgetItem{
					{
						{Type: "custom-text", Text: "Line 1"},
					},
					{},
					{
						{Type: "custom-text", Text: "Line 3"},
					},
				},
			},
			ctx: types.RenderContext{},
			assertFunc: func(t *testing.T, lines []string) {
				if len(lines) != 3 {
					t.Fatalf("expected 3 lines, got %d", len(lines))
				}
				if StripAnsi(lines[0]) != "Line 1" {
					t.Errorf("expected line 1 to be 'Line 1', got %q", lines[0])
				}
				if lines[1] != "" {
					t.Errorf("expected line 2 to be empty, got %q", lines[1])
				}
				if StripAnsi(lines[2]) != "Line 3" {
					t.Errorf("expected line 3 to be 'Line 3', got %q", lines[2])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := RenderStatusLines(tt.settings, tt.ctx)
			tt.assertFunc(t, lines)
		})
	}
}

func TestRenderStatusLines_Powerline(t *testing.T) {
	allThemes := []string{
		"nord", "nord-aurora", "monokai", "solarized", "minimal",
		"dracula", "catppuccin", "gruvbox", "onedark", "tokyonight",
	}

	for _, themeName := range allThemes {
		t.Run("theme_"+themeName, func(t *testing.T) {
			settings := types.Settings{
				Powerline: types.PowerlineConfig{
					Enabled:   true,
					Theme:     themeName,
					Separator: "\uE0B0",
				},
				General: types.GeneralConfig{
					ColorLevel: 2,
				},
				Lines: [][]types.WidgetItem{
					{
						{Type: "custom-text", Text: "Item1"},
						{Type: "custom-text", Text: "Item2"},
					},
				},
			}

			lines := RenderStatusLines(settings, types.RenderContext{})
			if len(lines) != 1 {
				t.Fatalf("expected 1 line for theme %s, got %d", themeName, len(lines))
			}

			line := lines[0]
			if !strings.Contains(line, "Item1") || !strings.Contains(line, "Item2") {
				t.Errorf("expected items in powerline output for theme %s, got %q", themeName, line)
			}
			if !strings.Contains(line, "\uE0B0") {
				t.Errorf("expected separator in powerline output for theme %s, got %q", themeName, line)
			}
			if !strings.Contains(line, "\x1b[") {
				t.Errorf("expected ANSI sequences in powerline output for theme %s, got %q", themeName, line)
			}
		})
	}

	t.Run("start and end caps", func(t *testing.T) {
		settings := types.Settings{
			Powerline: types.PowerlineConfig{
				Enabled:   true,
				Theme:     "nord",
				Separator: "\uE0B0",
				StartCaps: "\uE0B2",
				EndCaps:   "\uE0B0",
			},
			General: types.GeneralConfig{
				ColorLevel: 2,
			},
			Lines: [][]types.WidgetItem{
				{
					{Type: "custom-text", Text: "P1"},
					{Type: "custom-text", Text: "P2"},
				},
			},
		}

		lines := RenderStatusLines(settings, types.RenderContext{})
		if len(lines) != 1 {
			t.Fatalf("expected 1 line, got %d", len(lines))
		}

		line := lines[0]
		if !strings.Contains(line, "\uE0B2") {
			t.Errorf("expected start cap in powerline output, got %q", line)
		}
		if !strings.Contains(line, "\uE0B0") {
			t.Errorf("expected end cap in powerline output, got %q", line)
		}
		// In nord 256: first element bg is ansi256:73 -> capFg is ansi256:73 -> \x1b[38;5;73m
		if !strings.Contains(line, "\x1b[38;5;73m\uE0B2") {
			t.Errorf("expected colored start cap, got %q", line)
		}
	})

	t.Run("empty items in powerline", func(t *testing.T) {
		settings := types.Settings{
			Powerline: types.PowerlineConfig{
				Enabled: true,
				Theme:   "nord",
			},
			Lines: [][]types.WidgetItem{
				{
					{Type: "separator"},
					{Type: "invalid_widget_xyz"},
				},
			},
		}

		lines := RenderStatusLines(settings, types.RenderContext{})
		if len(lines) != 1 {
			t.Fatalf("expected 1 line, got %d", len(lines))
		}
		if lines[0] != "" {
			t.Errorf("expected empty string when no active widgets exist, got %q", lines[0])
		}
	})
}

func TestRenderStatusLines_ColorLevels(t *testing.T) {
	tests := []struct {
		name       string
		colorLevel int
		powerline  bool
		verify     func(t *testing.T, output string)
	}{
		{
			name:       "non-powerline ansi16",
			colorLevel: 1,
			powerline:  false,
			verify: func(t *testing.T, output string) {
				if !strings.Contains(output, "\x1b[31m") {
					t.Errorf("expected ansi16 code \x1b[31m, got %q", output)
				}
			},
		},
		{
			name:       "non-powerline ansi256",
			colorLevel: 2,
			powerline:  false,
			verify: func(t *testing.T, output string) {
				if !strings.Contains(output, "\x1b[38;5;160m") {
					t.Errorf("expected ansi256 code \x1b[38;5;160m, got %q", output)
				}
			},
		},
		{
			name:       "non-powerline truecolor",
			colorLevel: 3,
			powerline:  false,
			verify: func(t *testing.T, output string) {
				if !strings.Contains(output, "\x1b[38;2;204;0;0m") {
					t.Errorf("expected truecolor code \x1b[38;2;204;0;0m, got %q", output)
				}
			},
		},
		{
			name:       "powerline truecolor",
			colorLevel: 3,
			powerline:  true,
			verify: func(t *testing.T, output string) {
				if !strings.Contains(output, "\x1b[48;2;") {
					t.Errorf("expected truecolor bg code \x1b[48;2;..., got %q", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := types.Settings{
				Powerline: types.PowerlineConfig{
					Enabled: tt.powerline,
					Theme:   "nord-aurora",
				},
				General: types.GeneralConfig{
					ColorLevel: tt.colorLevel,
				},
				Lines: [][]types.WidgetItem{
					{
						{Type: "custom-text", Text: "ColorTest", Color: "red"},
					},
				},
			}

			lines := RenderStatusLines(settings, types.RenderContext{})
			if len(lines) != 1 {
				t.Fatalf("expected 1 line, got %d", len(lines))
			}
			tt.verify(t, lines[0])
		})
	}
}

func TestRenderStatusLines_TerminalWidthAndTruncation(t *testing.T) {
	settings := types.Settings{
		General: types.GeneralConfig{
			Separator: " · ",
		},
		Lines: [][]types.WidgetItem{
			{
				{Type: "custom-text", Text: "A very long statusline message that exceeds terminal width"},
			},
		},
	}

	tests := []struct {
		name          string
		terminalWidth int
		isPreview     bool
		maxExpected   int
	}{
		{"width 30", 30, false, 30},
		{"width 15", 15, false, 15},
		{"width 5", 5, false, 5},
		{"preview with width 40", 40, true, 34},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tw := tt.terminalWidth
			ctx := types.RenderContext{
				TerminalWidth: &tw,
				IsPreview:     tt.isPreview,
			}

			lines := RenderStatusLines(settings, ctx)
			if len(lines) != 1 {
				t.Fatalf("expected 1 line, got %d", len(lines))
			}

			visWidth := GetVisibleWidth(lines[0])
			if visWidth > tt.maxExpected {
				t.Errorf("visible width %d exceeds max expected %d", visWidth, tt.maxExpected)
			}
			if !strings.HasSuffix(lines[0], "...") {
				t.Errorf("expected truncated line to end with '...', got %q", lines[0])
			}
		})
	}
}

func TestRenderStatusLines_ExtremeTerminalWidths(t *testing.T) {
	settings := types.Settings{
		General: types.GeneralConfig{
			Padding:   " ",
			Separator: " | ",
		},
		Lines: [][]types.WidgetItem{
			{
				{Type: "custom-text", Text: "Long Custom Statusline Text For Testing Truncation Bounds"},
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
				t.Errorf("expected 1 line for termWidth %d preview=%v, got %d", tw, isPreview, len(lines))
			}
		}
	}
}

func TestGetPowerlineTheme_LookupsAndFallback(t *testing.T) {
	allThemes := []string{
		"nord", "nord-aurora", "monokai", "solarized", "minimal",
		"dracula", "catppuccin", "gruvbox", "onedark", "tokyonight",
	}

	for _, name := range allThemes {
		theme := GetPowerlineTheme(name)
		if theme == nil {
			t.Errorf("expected theme %q to exist", name)
		}
	}

	nilTheme := GetPowerlineTheme("non_existent_theme_name_12345")
	if nilTheme != nil {
		t.Errorf("expected nil for non-existent theme, got %v", nilTheme)
	}
}

func TestRenderer_ConcurrentRenders(t *testing.T) {
	settings := types.Settings{
		General: types.GeneralConfig{
			Padding:    " ",
			Separator:  " | ",
			ColorLevel: 3,
		},
		Powerline: types.PowerlineConfig{
			Enabled:   true,
			Theme:     "nord",
			Separator: "\uE0B0",
			StartCaps: "\uE0B2",
			EndCaps:   "\uE0B0",
		},
		Lines: [][]types.WidgetItem{
			{
				{Type: "model", Color: "gradient:red,blue"},
				{Type: "custom-text", Text: RenderOsc8Link("https://example.com", "TestLink")},
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

	for range numGoroutines {
		wg.Go(func() {
			for range iterationsPerGoroutine {
				lines := RenderStatusLines(settings, ctx)
				if len(lines) == 0 {
					t.Errorf("concurrent RenderStatusLines returned empty output")
				}
				trunc := TruncateStyledText("Some \x1b[31mstyled\x1b[0m text with \x1b]8;;https://example.com\x1b\\link\x1b]8;;\x1b\\", 20)
				if trunc == "" {
					t.Errorf("concurrent TruncateStyledText returned empty output")
				}
				code := GetColorAnsiCode("gradient:red,green,blue", "truecolor", false)
				if code == "" {
					t.Errorf("concurrent GetColorAnsiCode returned empty code")
				}
				app := ApplyColors("Sample", "hex:123456", "bgRed", nil, "ansi256", nil)
				if app == "" {
					t.Errorf("concurrent ApplyColors returned empty output")
				}
			}
		})
	}

	wg.Wait()
}

func TestRenderer_HugeInputs(t *testing.T) {
	hugeString := strings.Repeat("A\x1b[31mB\x1b[0mC ", 2000) // ~28KB
	stripped := StripAnsi(hugeString)
	if len(stripped) == 0 {
		t.Errorf("expected non-empty stripped output for huge input")
	}

	w := GetVisibleWidth(hugeString)
	if w <= 0 {
		t.Errorf("expected positive visible width for huge input, got %d", w)
	}

	truncated := TruncateStyledText(hugeString, 500)
	if GetVisibleWidth(truncated) > 500 {
		t.Errorf("truncated width %d exceeds max 500", GetVisibleWidth(truncated))
	}
}

func TestAdversarial_AllThemesAllColorLevelsMatrix(t *testing.T) {
	allThemes := []string{
		"nord", "nord-aurora", "monokai", "solarized", "minimal",
		"dracula", "catppuccin", "gruvbox", "onedark", "tokyonight",
	}
	colorLevels := []int{1, 2, 3}

	statusData := types.StatusJSON{
		Model:         types.ModelInfo{DisplayName: "Claude 3.5 Sonnet"},
		Sandbox:       &types.SandboxInfo{Enabled: new(true)},
		ArtifactCount: new(7),
		VCS:           &types.VCSInfo{Branch: "main"},
		Quota: map[string]types.QuotaInfo{
			"rpc": {RemainingFraction: new(0.5)},
		},
	}

	for _, themeName := range allThemes {
		for _, colorLevel := range colorLevels {
			for _, withCaps := range []bool{false, true} {
				t.Run(themeName+"_level"+string(rune('0'+colorLevel))+"_caps_"+string(rune('0'+boolToInt(withCaps))), func(t *testing.T) {
					startCap := ""
					endCap := ""
					if withCaps {
						startCap = "\uE0B2"
						endCap = "\uE0B0"
					}

					settings := types.Settings{
						Powerline: types.PowerlineConfig{
							Enabled:   true,
							Theme:     themeName,
							Separator: "\uE0B0",
							StartCaps: startCap,
							EndCaps:   endCap,
						},
						General: types.GeneralConfig{
							Padding:    " ",
							ColorLevel: colorLevel,
						},
						Lines: [][]types.WidgetItem{
							{
								{Type: "model"},
								{Type: "sandbox"},
								{Type: "artifacts"},
								{Type: "git-branch"},
								{Type: "quota", Key: "rpc"},
								{Type: "custom-text", Text: "Custom"},
							},
						},
					}

					ctx := types.RenderContext{
						Data: statusData,
					}

					lines := RenderStatusLines(settings, ctx)
					if len(lines) != 1 {
						t.Fatalf("expected 1 line, got %d", len(lines))
					}

					line := lines[0]
					if line == "" {
						t.Fatalf("expected non-empty rendered line for theme %s level %d", themeName, colorLevel)
					}

					// Verify clean termination resets
					if !strings.Contains(line, "\x1b[49m\x1b[39m") {
						t.Errorf("expected clean reset sequences \\x1b[49m\\x1b[39m in line: %q", line)
					}

					// Verify caps if configured
					if withCaps {
						if !strings.Contains(line, startCap) {
							t.Errorf("expected start cap %q in line", startCap)
						}
						if !strings.Contains(line, endCap) {
							t.Errorf("expected end cap %q in line", endCap)
						}
					}

					// Verify color level specific escape codes
					switch colorLevel {
					case 2: // ANSI 256
						if !strings.Contains(line, "\x1b[38;5;") && !strings.Contains(line, "\x1b[48;5;") {
							t.Errorf("expected ANSI 256 codes in theme %s, got %q", themeName, line)
						}
					case 3: // Truecolor
						if !strings.Contains(line, "\x1b[38;2;") && !strings.Contains(line, "\x1b[48;2;") {
							t.Errorf("expected Truecolor codes in theme %s, got %q", themeName, line)
						}
					}

					// Verify plain text contains widgets content
					stripped := StripAnsi(line)
					for _, expectedContent := range []string{"Claude 3.5 Sonnet", "on", "7", "main", "50.00%", "Custom"} {
						if !strings.Contains(stripped, expectedContent) {
							t.Errorf("expected %q in stripped output, got %q", expectedContent, stripped)
						}
					}
				})
			}
		}
	}
}

func TestAdversarial_ExtremeTerminalWidthsAndTruncationMatrix(t *testing.T) {
	testStrings := []struct {
		name string
		text string
	}{
		{"pure ASCII", "Simple Status Line With Standard Characters"},
		{"CJK Japanese", "Claude 3.5 Sonnet モデル · サンドボックス 有効 · 日本語ステータスライン"},
		{"Emoji and Symbols", "🚀 Model: Claude 🟢 Sandbox: ON 📦 Artifacts: 12 🌿 main ⚡"},
		{"OSC 8 Hyperlinks", "Link: " + RenderOsc8Link("https://example.com/docs", "Documentation") + " and more"},
		{"Malformed ANSI", "Broken: \x1b[31mRed \x1b[32mGreen \x1b[99999m \x1b[ Unterminated"},
		{"Mixed Heavy", "🚀 " + RenderOsc8Link("https://example.org", "リンク") + " \x1b[1;34m太字青\x1b[0m 🍣 寿司 (+100,-50)"},
	}

	widths := []int{0, 1, 2, 5, 10, 80, 1000}

	for _, ts := range testStrings {
		for _, width := range widths {
			for _, isPreview := range []bool{false, true} {
				for _, powerline := range []bool{false, true} {
					t.Run(ts.name+"_w"+string(rune('0'+width/100))+string(rune('0'+(width%100)/10))+string(rune('0'+width%10))+"_prev_"+string(rune('0'+boolToInt(isPreview)))+"_pl_"+string(rune('0'+boolToInt(powerline))), func(t *testing.T) {
						settings := types.Settings{
							Powerline: types.PowerlineConfig{
								Enabled:   powerline,
								Theme:     "nord",
								Separator: "\uE0B0",
							},
							General: types.GeneralConfig{
								ColorLevel: 3,
								Separator:  " · ",
							},
							Lines: [][]types.WidgetItem{
								{
									{Type: "custom-text", Text: ts.text},
								},
							},
						}

						wVal := width
						ctx := types.RenderContext{
							TerminalWidth: &wVal,
							IsPreview:     isPreview,
						}

						lines := RenderStatusLines(settings, ctx)
						if len(lines) != 1 {
							t.Fatalf("expected 1 line, got %d", len(lines))
						}

						line := lines[0]
						visWidth := GetVisibleWidth(line)

						if width > 0 {
							effectiveMax := width
							if isPreview {
								effectiveMax = max(width-6, 0)
							}

							if visWidth > effectiveMax {
								t.Errorf("visible width %d exceeds effective maximum %d (input width %d, isPreview %v, line: %q)",
									visWidth, effectiveMax, width, isPreview, line)
							}
						}

						// Verify StripAnsi idempotency
						s1 := StripAnsi(line)
						s2 := StripAnsi(s1)
						if s1 != s2 {
							t.Errorf("StripAnsi is not idempotent: %q vs %q", s1, s2)
						}
					})
				}
			}
		}
	}
}

func TestAdversarial_MinimalistAndRawModes_AllWidgets(t *testing.T) {
	statusData := types.StatusJSON{
		Model:         types.ModelInfo{DisplayName: "Claude 3.5 Sonnet"},
		ContextWindow: &types.ContextWindowInfo{UsedPercentage: new(25.0)},
		AgentState:    "working",
		ArtifactCount: new(3),
		Subagents:     2,
		TaskCount:     new(4),
		Sandbox:       &types.SandboxInfo{Enabled: new(true)},
		Quota: map[string]types.QuotaInfo{
			"rpc": {RemainingFraction: new(0.2)},
		},
		VCS: &types.VCSInfo{Branch: "feature-x"},
	}

	widgetTypes := []struct {
		item         types.WidgetItem
		expectedBody string
		title        string
	}{
		{types.WidgetItem{Type: "model"}, "Claude 3.5 Sonnet", ""},
		{types.WidgetItem{Type: "context-bar"}, "25.0%", "ctx"},
		{types.WidgetItem{Type: "agent-state"}, "WORKING", ""},
		{types.WidgetItem{Type: "artifacts"}, "3", "artifacts"},
		{types.WidgetItem{Type: "subagents"}, "2", "subagents"},
		{types.WidgetItem{Type: "tasks"}, "4", "tasks"},
		{types.WidgetItem{Type: "sandbox"}, "on", "sandbox"},
		{types.WidgetItem{Type: "quota", Key: "rpc"}, "20.00%", "rpc"},
		{types.WidgetItem{Type: "quota-bar", Key: "rpc"}, "20.0%", "rpc"},
		{types.WidgetItem{Type: "git-branch"}, "feature-x", ""},
		{types.WidgetItem{Type: "custom-text", Text: "CustomStatic"}, "CustomStatic", ""},
	}

	for _, wt := range widgetTypes {
		t.Run("widget_"+wt.item.Type+"_modes", func(t *testing.T) {
			// Mode 1: Standard (title should be present if non-empty)
			itemStd := wt.item
			settingsStd := types.Settings{
				General: types.GeneralConfig{Separator: " · "},
				Lines: [][]types.WidgetItem{
					{itemStd},
				},
			}
			linesStd := RenderStatusLines(settingsStd, types.RenderContext{Data: statusData})
			if len(linesStd) != 1 {
				t.Fatalf("expected 1 line, got %d", len(linesStd))
			}
			strippedStd := StripAnsi(linesStd[0])
			if !strings.Contains(strippedStd, wt.expectedBody) {
				t.Errorf("standard mode expected body %q, got %q", wt.expectedBody, strippedStd)
			}
			if wt.title != "" && !strings.Contains(strippedStd, wt.title) {
				t.Errorf("standard mode expected title %q, got %q", wt.title, strippedStd)
			}

			// Mode 2: Item Raw = true (title must be omitted)
			itemRaw := wt.item
			itemRaw.Raw = true
			settingsRaw := types.Settings{
				General: types.GeneralConfig{Separator: " · "},
				Lines: [][]types.WidgetItem{
					{itemRaw},
				},
			}
			linesRaw := RenderStatusLines(settingsRaw, types.RenderContext{Data: statusData})
			strippedRaw := StripAnsi(linesRaw[0])
			if !strings.Contains(strippedRaw, wt.expectedBody) {
				t.Errorf("raw mode expected body %q, got %q", wt.expectedBody, strippedRaw)
			}
			if wt.title != "" && strings.Contains(strippedRaw, wt.title+" ") {
				t.Errorf("raw mode should omit title %q, got %q", wt.title, strippedRaw)
			}

			// Mode 3: Settings Minimalist = true (title must be omitted)
			settingsMin := types.Settings{
				General: types.GeneralConfig{Minimalist: true, Separator: " · "},
				Lines: [][]types.WidgetItem{
					{wt.item},
				},
			}
			linesMin := RenderStatusLines(settingsMin, types.RenderContext{Data: statusData})
			strippedMin := StripAnsi(linesMin[0])
			if !strings.Contains(strippedMin, wt.expectedBody) {
				t.Errorf("minimalist settings mode expected body %q, got %q", wt.expectedBody, strippedMin)
			}
			if wt.title != "" && strings.Contains(strippedMin, wt.title+" ") {
				t.Errorf("minimalist settings mode should omit title %q, got %q", wt.title, strippedMin)
			}

			// Mode 4: Context Minimalist = true (title must be omitted)
			linesCtxMin := RenderStatusLines(settingsStd, types.RenderContext{Data: statusData, Minimalist: true})
			strippedCtxMin := StripAnsi(linesCtxMin[0])
			if !strings.Contains(strippedCtxMin, wt.expectedBody) {
				t.Errorf("minimalist context mode expected body %q, got %q", wt.expectedBody, strippedCtxMin)
			}
			if wt.title != "" && strings.Contains(strippedCtxMin, wt.title+" ") {
				t.Errorf("minimalist context mode should omit title %q, got %q", wt.title, strippedCtxMin)
			}
		})
	}
}

func TestAdversarial_AdjacentPowerlineElementsSameColor(t *testing.T) {
	settings := types.Settings{
		Powerline: types.PowerlineConfig{
			Enabled:   true,
			Theme:     "nord",
			Separator: "\uE0B0",
		},
		General: types.GeneralConfig{
			ColorLevel: 2,
		},
		Lines: [][]types.WidgetItem{
			{
				{Type: "custom-text", Text: "El1", Color: "red"},
				{Type: "custom-text", Text: "El2", Color: "red"},
			},
		},
	}

	lines := RenderStatusLines(settings, types.RenderContext{})
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	line := lines[0]
	if !strings.Contains(line, "El1") || !strings.Contains(line, "El2") {
		t.Errorf("expected both elements in output, got %q", line)
	}
	if !strings.Contains(line, "\uE0B0") {
		t.Errorf("expected separator in output, got %q", line)
	}
}

func TestAdversarial_MalformedANSIAndOSC8(t *testing.T) {
	malformedCases := []struct {
		name     string
		input    string
		maxWidth int
	}{
		{"unterminated CSI", "\x1b[31mRedTextWithoutTermination\x1b[32", 10},
		{"unterminated OSC", "\x1b]8;;https://example.com/unterminated", 15},
		{"stray ESC byte", "Hello\x1bWorld\x1b", 8},
		{"multiple ESC bytes", "\x1b\x1b\x1b[31mTripleEsc\x1b[0m", 6},
		{"CSI with empty params", "\x1b[;mNoParams\x1b[m", 5},
		{"colon syntax in SGR", "\x1b[38:2:255:0:0mColonTruecolor\x1b[0m", 10},
		{"OSC with parameters", "\x1b]8;id=abc:type=link;https://example.org\x1b\\ParamLink\x1b]8;;\x1b\\", 8},
		{"OSC with BEL terminator", "\x1b]8;;https://example.net\x07BelLink\x1b]8;;\x07", 6},
		{"CJK with malformed CSI", "こんにちは\x1b[31m世界\x1b[99999", 8},
		{"Emoji with malformed OSC", "🚀\x1b]8;;https://example.com\x1b\\EmojiLink", 4},
	}

	for _, mc := range malformedCases {
		t.Run(mc.name, func(t *testing.T) {
			// GetVisibleWidth should never panic
			w := GetVisibleWidth(mc.input)
			if w < 0 {
				t.Errorf("GetVisibleWidth(%q) returned negative width %d", mc.input, w)
			}

			// StripAnsi should never panic
			_ = StripAnsi(mc.input)

			// TruncateStyledText should never panic and obey bounds
			truncated := TruncateStyledText(mc.input, mc.maxWidth)
			visWidth := GetVisibleWidth(truncated)
			if visWidth > mc.maxWidth {
				t.Errorf("TruncateStyledText(%q, %d) width %d exceeds max", mc.input, mc.maxWidth, visWidth)
			}
		})
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func TestRenderStatusLines_AllQuotaWidgets_ColorLevel1_Integrity(t *testing.T) {
	highFraction := 0.85
	midFraction := 0.35
	lowFraction := 0.10
	secs := 3600.0

	statusData := types.StatusJSON{
		Quota: map[string]types.QuotaInfo{
			"gemini-5h": {
				RemainingFraction: &highFraction,
				ResetInSeconds:    &secs,
			},
			"gemini-weekly": {
				RemainingFraction: &midFraction,
				ResetInSeconds:    &secs,
			},
			"3p-5h": {
				RemainingFraction: &lowFraction,
				ResetInSeconds:    &secs,
			},
			"3p-weekly": {
				RemainingFraction: &highFraction,
				ResetInSeconds:    &secs,
			},
		},
	}

	allQuotaWidgets := []types.WidgetItem{
		{Type: "quota-5h"},
		{Type: "quota-7d"},
		{Type: "quota-3p-5h"},
		{Type: "quota-3p-7d"},
		{Type: "quota-bar-5h"},
		{Type: "quota-bar-7d"},
		{Type: "quota-bar-3p-5h"},
		{Type: "quota-bar-3p-7d"},
	}

	t.Run("Standard Non-Powerline ColorLevel 1", func(t *testing.T) {
		settings := types.Settings{
			General: types.GeneralConfig{
				ColorLevel: 1, // Default ANSI 16
				Separator:  " · ",
			},
			Lines: [][]types.WidgetItem{
				allQuotaWidgets,
			},
		}

		ctx := types.RenderContext{
			Data: statusData,
		}

		lines := RenderStatusLines(settings, ctx)
		if len(lines) != 1 {
			t.Fatalf("expected 1 line, got %d", len(lines))
		}

		out := lines[0]

		// Must not contain 256-color or truecolor sequences
		if strings.Contains(out, "\x1b[38;5;") || strings.Contains(out, "\x1b[48;5;") {
			t.Errorf("Unexpected ANSI 256 escape code in ColorLevel 1 output: %q", out)
		}
		if strings.Contains(out, "\x1b[38;2;") || strings.Contains(out, "\x1b[48;2;") {
			t.Errorf("Unexpected Truecolor escape code in ColorLevel 1 output: %q", out)
		}

		// Must contain ANSI 16 bright colors:
		// brightBlack (\x1b[90m) for titles & separator
		// brightWhite (\x1b[97m) for quota text
		// brightGreen (\x1b[92m) for high quota bar
		// brightYellow (\x1b[93m) for mid quota bar
		// brightRed (\x1b[91m) for low quota bar
		if !strings.Contains(out, "\x1b[90m") {
			t.Errorf("Expected \\x1b[90m (brightBlack) in output: %q", out)
		}
		if !strings.Contains(out, "\x1b[97m") {
			t.Errorf("Expected \\x1b[97m (brightWhite) in output: %q", out)
		}
		if !strings.Contains(out, "\x1b[92m") {
			t.Errorf("Expected \\x1b[92m (brightGreen) in output: %q", out)
		}
		if !strings.Contains(out, "\x1b[93m") {
			t.Errorf("Expected \\x1b[93m (brightYellow) in output: %q", out)
		}
		if !strings.Contains(out, "\x1b[91m") {
			t.Errorf("Expected \\x1b[91m (brightRed) in output: %q", out)
		}

		// Verify every ANSI color is cleanly reset with \x1b[39m
		if !strings.Contains(out, "\x1b[39m") {
			t.Errorf("Expected \\x1b[39m reset sequence in output: %q", out)
		}
	})

	t.Run("Powerline ColorLevel 1 (Default nord-aurora theme)", func(t *testing.T) {
		settings := types.Settings{
			Powerline: types.PowerlineConfig{
				Enabled:   true,
				Theme:     "nord-aurora",
				Separator: "\uE0B0",
			},
			General: types.GeneralConfig{
				ColorLevel: 1, // Default ANSI 16
			},
			Lines: [][]types.WidgetItem{
				allQuotaWidgets,
			},
		}

		ctx := types.RenderContext{
			Data: statusData,
		}

		lines := RenderStatusLines(settings, ctx)
		if len(lines) != 1 {
			t.Fatalf("expected 1 line, got %d", len(lines))
		}

		out := lines[0]

		// Must not contain 256-color or truecolor sequences
		if strings.Contains(out, "\x1b[38;5;") || strings.Contains(out, "\x1b[48;5;") {
			t.Errorf("Unexpected ANSI 256 escape code in ColorLevel 1 output: %q", out)
		}
		if strings.Contains(out, "\x1b[38;2;") || strings.Contains(out, "\x1b[48;2;") {
			t.Errorf("Unexpected Truecolor escape code in ColorLevel 1 output: %q", out)
		}

		// Verify clean resets for background & foreground
		if !strings.Contains(out, "\x1b[49m\x1b[39m") {
			t.Errorf("Expected \\x1b[49m\\x1b[39m resets in Powerline output: %q", out)
		}
	})
}
