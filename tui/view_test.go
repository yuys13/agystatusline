package tui

import (
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/renderer"
	"github.com/yuys13/agystatusline/types"
	"github.com/yuys13/agystatusline/widgets"
)

func TestBuildPreviewSettings(t *testing.T) {
	t.Run("select_theme updates preview theme", func(t *testing.T) {
		m := NewModel(types.DefaultSettings(), "/tmp/settings.toml")
		m.activeMenu = menuSelectTheme
		m.cursor = 2 // monokai

		got := m.buildPreviewSettings()
		if got.Powerline.Theme != themesList[2] {
			t.Errorf("Powerline.Theme = %q, want %q", got.Powerline.Theme, themesList[2])
		}
	})

	t.Run("select_separator updates preview separator", func(t *testing.T) {
		m := NewModel(types.DefaultSettings(), "/tmp/settings.toml")
		m.activeMenu = menuSelectSeparator
		m.cursor = 2 // Round

		got := m.buildPreviewSettings()
		if got.Powerline.Separator != m.separators[2].value {
			t.Errorf("Powerline.Separator = %q, want %q", got.Powerline.Separator, m.separators[2].value)
		}
	})

	t.Run("select_start_cap updates preview start cap", func(t *testing.T) {
		m := NewModel(types.DefaultSettings(), "/tmp/settings.toml")
		m.activeMenu = menuSelectStartCap
		m.cursor = 2

		got := m.buildPreviewSettings()
		if got.Powerline.StartCaps != m.startCaps[2].value {
			t.Errorf("Powerline.StartCaps = %q, want %q", got.Powerline.StartCaps, m.startCaps[2].value)
		}
	})

	t.Run("select_end_cap updates preview end cap", func(t *testing.T) {
		m := NewModel(types.DefaultSettings(), "/tmp/settings.toml")
		m.activeMenu = menuSelectEndCap
		m.cursor = 2

		got := m.buildPreviewSettings()
		if got.Powerline.EndCaps != m.endCaps[2].value {
			t.Errorf("Powerline.EndCaps = %q, want %q", got.Powerline.EndCaps, m.endCaps[2].value)
		}
	})

	t.Run("select_color_level updates preview color level", func(t *testing.T) {
		m := NewModel(types.DefaultSettings(), "/tmp/settings.toml")
		m.activeMenu = menuSelectColorLevel
		m.cursor = 2 // Truecolor

		got := m.buildPreviewSettings()
		if got.General.ColorLevel != colorLevelsList[2].value {
			t.Errorf("General.ColorLevel = %d, want %d", got.General.ColorLevel, colorLevelsList[2].value)
		}
	})

	t.Run("add_widget inserts preview widget without mutating model settings", func(t *testing.T) {
		s := types.DefaultSettings()
		s.Lines = [][]types.WidgetItem{{{Type: "model"}}}
		m := NewModel(s, "/tmp/settings.toml")
		m.activeMenu = menuAddWidget
		m.selectedLine = 0
		m.itemIndex = 0
		m.cursor = 2 // context-bar

		got := m.buildPreviewSettings()
		if len(got.Lines[0]) != 2 {
			t.Fatalf("len(got.Lines[0]) = %d, want 2", len(got.Lines[0]))
		}
		if got.Lines[0][1].Type != widgetTypes[2].item.Type {
			t.Errorf("got.Lines[0][1].Type = %q, want %q", got.Lines[0][1].Type, widgetTypes[2].item.Type)
		}
		if len(m.settings.Lines[0]) != 1 {
			t.Errorf("m.settings.Lines[0] was mutated: len = %d, want 1", len(m.settings.Lines[0]))
		}
	})

	t.Run("main menu preserves settings", func(t *testing.T) {
		s := types.DefaultSettings()
		s.Powerline.Theme = "nord"
		m := NewModel(s, "/tmp/settings.toml")
		m.activeMenu = menuMain

		got := m.buildPreviewSettings()
		if got.Powerline.Theme != "nord" {
			t.Errorf("Powerline.Theme = %q, want %q", got.Powerline.Theme, "nord")
		}
	})
}

func TestWriteChoiceListView(t *testing.T) {
	items := []choiceItem{
		{name: "None", value: ""},
		{name: "Arrow (\uE0B0)", value: "\uE0B0"},
		{name: "Round (\uE0B4)", value: "\uE0B4"},
	}

	tests := []struct {
		name        string
		title       string
		cursor      int
		activeValue string
		wantLines   []string
	}{
		{
			name:        "renders title cursor active marker and footer",
			title:       "Select Powerline Separator",
			cursor:      1,
			activeValue: "\uE0B4",
			wantLines: []string{
				"Select Powerline Separator",
				"  None",
				"> Arrow (\uE0B0)",
				"  Round (\uE0B4) (active)",
				"(Use arrows/jk to navigate, Enter to select, Esc: cancel)",
			},
		},
		{
			name:        "active item matches cursor",
			title:       "Select Powerline Start Cap",
			cursor:      0,
			activeValue: "",
			wantLines: []string{
				"Select Powerline Start Cap",
				"> None (active)",
				"  Arrow (\uE0B0)",
				"  Round (\uE0B4)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(types.DefaultSettings(), "/tmp/settings.toml")
			m.cursor = tt.cursor

			var sb strings.Builder
			m.writeChoiceListView(&sb, tt.title, items, tt.activeValue)
			plain := renderer.StripAnsi(sb.String())

			for _, want := range tt.wantLines {
				if !strings.Contains(plain, want) {
					t.Errorf("output missing %q\ngot:\n%s", want, plain)
				}
			}
		})
	}
}

func TestTUI_LivePreviewModelName(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")

	viewStr := m.View()
	expectedModelName := "Gemini 3.5 Flash (Medium)"
	if !strings.Contains(viewStr, expectedModelName) {
		t.Errorf("Expected live preview to contain model name %q, but it did not. View output:\n%s", expectedModelName, viewStr)
	}
}

func TestTUI_LayoutAndBorders(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")

	viewStr := m.View()

	// 1. Verify preview is at the top (i.e. "--- Live Preview ---" is shown before "Configuration Menu")
	previewIdx := strings.Index(viewStr, "--- Live Preview ---")
	menuIdx := strings.Index(viewStr, "Configuration Menu")

	if previewIdx == -1 {
		t.Fatalf("Expected view to contain '--- Live Preview ---'")
	}
	if menuIdx == -1 {
		t.Fatalf("Expected view to contain 'Configuration Menu'")
	}
	if previewIdx > menuIdx {
		t.Errorf("Expected '--- Live Preview ---' to appear before 'Configuration Menu'")
	}

	// 2. Verify that there are no border characters around the preview
	borderChars := []string{"│", "─", "┌", "┐", "└", "┘"}
	for _, char := range borderChars {
		if strings.Contains(viewStr, char) {
			t.Errorf("Expected no border character %q in the view output, but found one", char)
		}
	}
}

func TestTUI_LivePreviewQuota(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	settings.Lines[0] = []types.WidgetItem{
		{
			Type: "quota",
			Key:  "gemini-5h",
		},
		{
			Type: "quota",
			Key:  "3p-5h",
		},
	}
	m := NewModel(settings, "/tmp/settings.toml")

	viewStr := renderer.StripAnsi(m.View())
	if !strings.Contains(viewStr, "Gemini 5h 50.19% (2h 28m)") {
		t.Errorf("Expected live preview to contain 'Gemini 5h 50.19%% (2h 28m)', but it did not. View:\n%s", viewStr)
	}
	if !strings.Contains(viewStr, "Claude 5h 100.00% (4h 59m)") {
		t.Errorf("Expected live preview to contain 'Claude 5h 100.00%% (4h 59m)', but it did not. View:\n%s", viewStr)
	}

	t.Run("quota-all and quota-bar-all in live preview", func(t *testing.T) {
		s := types.DefaultSettings()
		s.Lines = [][]types.WidgetItem{
			{{Type: "quota-all"}},
			{{Type: "quota-bar-all"}},
		}
		mAll := NewModel(s, "/tmp/settings.toml")
		vAll := renderer.StripAnsi(mAll.View())

		for _, want := range []string{
			"Gemini 5h 50.19% (2h 28m) · Gemini 7d 90.91% (6d 13h) · Argon 7d 100.00% (6d 23h) · Claude 5h 100.00% (4h 59m) · Claude 7d 100.00% (6d 23h)",
			"Gemini 5h █████····· 50.2% (2h 28m) · Gemini 7d █████████· 90.9% (6d 13h) · Argon 7d ██████████ 100.0% (6d 23h)",
		} {
			if !strings.Contains(vAll, want) {
				t.Errorf("Expected live preview to contain %q, but it did not. View:\n%s", want, vAll)
			}
		}
	})
}

func TestTUI_LivePreviewAddWidget(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")

	// Set state to "add_widget" menu
	m.activeMenu = "add_widget"
	m.selectedLine = 0
	m.itemIndex = 0 // Insert after the first widget (index 0, agent-state widget)

	// Select "Quota Bar: 5h" widget
	targetIdx := -1
	for i, wt := range widgetTypes {
		if wt.item.Type == "quota-bar-5h" {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		t.Fatalf("Quota Bar: 5h not found in widgetTypes")
	}
	m.cursor = targetIdx

	viewStr := m.View()

	lines := strings.Split(viewStr, "\n")
	if len(lines) < 5 {
		t.Fatalf("Expected view outputs to have enough lines")
	}
	previewPart := strings.Join(lines[:4], "\n")

	if !strings.Contains(previewPart, "Gemini 5h") || !strings.Contains(previewPart, "50.2%") {
		t.Errorf("Expected Live Preview to dynamically display the currently selected widget type 'Quota Bar: Gemini 5h' in the preview part, but it did not. Preview part:\n%s", previewPart)
	}
}

func TestTUI_LivePreviewSandbox(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	// Append sandbox widget to the first line to test its preview rendering
	settings.Lines[0] = append(settings.Lines[0],
		types.WidgetItem{Type: "sandbox"},
	)
	m := NewModel(settings, "/tmp/settings.toml")

	viewStr := renderer.StripAnsi(m.View())
	// Since sandbox.enabled will be configured as true in the preview context,
	// the preview output should contain "sandbox on"
	if !strings.Contains(viewStr, "sandbox on") {
		t.Errorf("Expected live preview to contain 'sandbox on', but it did not. View output:\n%s", viewStr)
	}
}

func TestTUI_MainMenuSaveExitSpacing(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")

	viewStr := m.View()

	colorLevelIdx := strings.Index(viewStr, "Select Color Level")

	saveExitIdx := strings.Index(viewStr, "Save & Exit")

	if colorLevelIdx == -1 {
		t.Fatalf("Expected view to contain 'Select Color Level'")
	}
	if saveExitIdx == -1 {
		t.Fatalf("Expected view to contain 'Save & Exit'")
	}

	if colorLevelIdx >= saveExitIdx {
		t.Fatalf("Expected 'Select Color Level' to appear before 'Save & Exit'")
	}

	between := viewStr[colorLevelIdx:saveExitIdx]
	newlineCount := strings.Count(between, "\n")
	if newlineCount < 2 {
		t.Errorf("Expected an empty line between 'Select Color Level' and 'Save & Exit', but found %d newlines", newlineCount)
	}
}

func TestTUI_ViewSubmenus(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()

	tests := []struct {
		name         string
		activeMenu   string
		selectedLine int
		wantSubstr   string
	}{
		{
			name:       "activeMenu = lines",
			activeMenu: "lines",
			wantSubstr: "Select Line to Edit Items",
		},
		{
			name:         "activeMenu = items",
			activeMenu:   "items",
			selectedLine: 0,
			wantSubstr:   "Editing Line 1 Items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(settings, "/tmp/settings.toml")
			m.activeMenu = tt.activeMenu
			m.selectedLine = tt.selectedLine

			viewStr := m.View()
			if !strings.Contains(viewStr, tt.wantSubstr) {
				t.Errorf("Expected %q in view when activeMenu=%q, got:\n%s", tt.wantSubstr, tt.activeMenu, viewStr)
			}
		})
	}
}

func TestTUI_View_QuittingScreensAndMoveIndicators(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")

	// 1. View when quitting and saved == true
	m.quitting = true
	m.saved = true
	viewSaved := m.View()
	if !strings.Contains(viewSaved, "Configuration saved successfully. Exiting...") {
		t.Errorf("Expected quitting view with saved=true to contain saved success message, got:\n%s", viewSaved)
	}

	// 2. View when quitting and saved == false
	m.quitting = true
	m.saved = false
	viewDiscarded := m.View()
	if !strings.Contains(viewDiscarded, "Changes discarded. Exiting...") {
		t.Errorf("Expected quitting view with saved=false to contain changes discarded message, got:\n%s", viewDiscarded)
	}

	// 3. viewLines moveMode indicator ("M")
	m.quitting = false
	m.activeMenu = "lines"
	m.cursor = 0
	m.moveMode = true
	viewLinesMove := m.View()
	if !strings.Contains(viewLinesMove, "M Line 1:") {
		t.Errorf("Expected viewLines to display 'M Line 1:' when moveMode=true, got:\n%s", viewLinesMove)
	}

	// 4. viewItems moveMode indicator ("M") and empty widget line message
	m.activeMenu = "items"
	m.selectedLine = 0
	m.cursor = 0
	m.moveMode = true
	viewItemsMove := m.View()
	if !strings.Contains(viewItemsMove, "M ") {
		t.Errorf("Expected viewItems to display 'M ' cursor when moveMode=true, got:\n%s", viewItemsMove)
	}

	// Empty line message in viewItems
	emptySettings := types.DefaultSettings()
	emptySettings.Lines[0] = []types.WidgetItem{}
	mEmpty := NewModel(emptySettings, "/tmp/settings.toml")
	mEmpty.activeMenu = "items"
	mEmpty.selectedLine = 0
	viewEmptyLine := mEmpty.View()
	if !strings.Contains(viewEmptyLine, "(No widgets in this line)") {
		t.Errorf("Expected viewItems to display '(No widgets in this line)' for empty line, got:\n%s", viewEmptyLine)
	}
}

func TestTUI_MainMenuItems(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")
	viewStr := renderer.StripAnsi(m.View())

	expectedItems := []string{
		"📄 Edit Lines",
		"⚡ Powerline Settings",
		"🎨 Select Color Level",
		"💾 Save & Exit",
		"❌ Discard & Exit",
	}

	for _, item := range expectedItems {
		if !strings.Contains(viewStr, item) {
			t.Errorf("Expected main menu view to contain %q, but it did not. View output:\n%s", item, viewStr)
		}
	}
}
