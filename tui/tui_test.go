package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yuys13/agystatusline/renderer"
	"github.com/yuys13/agystatusline/types"
	"github.com/yuys13/agystatusline/widgets"
)

func TestMain(m *testing.M) {
	widgets.RegisterAll()
	os.Exit(m.Run())
}

func TestInitialModel(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")

	if len(m.settings.Lines) == 0 {
		t.Errorf("Expected non-empty default lines, got %d", len(m.settings.Lines))
	}

	if m.activeMenu != "main" {
		t.Errorf("Expected initial menu %q, got %q", "main", m.activeMenu)
	}
}

func TestTUI_UpdateQuit(t *testing.T) {
	settings := types.DefaultSettings()

	tests := []struct {
		name         string
		activeMenu   string
		keyMsg       tea.KeyMsg
		wantQuitting bool
	}{
		{
			name:         "Press 'q' on main menu exits",
			activeMenu:   "main",
			keyMsg:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")},
			wantQuitting: true,
		},
		{
			name:         "Press 'q' on lines menu does NOT exit",
			activeMenu:   "lines",
			keyMsg:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")},
			wantQuitting: false,
		},
		{
			name:         "Press 'ctrl+c' on lines menu exits",
			activeMenu:   "lines",
			keyMsg:       tea.KeyMsg{Type: tea.KeyCtrlC},
			wantQuitting: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(settings, "/tmp/settings.toml")
			m.activeMenu = tt.activeMenu
			updatedModel, _ := m.Update(tt.keyMsg)
			newModel := updatedModel.(Model)
			if newModel.quitting != tt.wantQuitting {
				t.Errorf("Expected quitting to be %v, got %v", tt.wantQuitting, newModel.quitting)
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

func TestTUI_NavigateToLines(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")

	// Set cursor to 0 ("Edit Lines" or replacement for Toggle Minimalist Mode)
	m.cursor = 0
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")} // Enter key
	updatedModel, _ := m.Update(msg)
	newModel := updatedModel.(Model)

	if newModel.activeMenu != "lines" {
		t.Errorf("Expected activeMenu to be 'lines' after pressing Enter on menu item 0, got %q", newModel.activeMenu)
	}
	if newModel.cursor != 0 {
		t.Errorf("Expected cursor to reset to 0, got %d", newModel.cursor)
	}
}

func TestTUI_LinesOperations(t *testing.T) {
	t.Run("AddLine", func(t *testing.T) {
		settings := types.DefaultSettings()
		initialLinesCount := len(settings.Lines)
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "lines"
		m.cursor = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
		updatedModel, _ := m.Update(msg)
		newModel := updatedModel.(Model)

		if len(newModel.settings.Lines) != initialLinesCount+1 {
			t.Errorf("Expected lines count to be %d, got %d", initialLinesCount+1, len(newModel.settings.Lines))
		}
		if newModel.cursor != initialLinesCount {
			t.Errorf("Expected cursor to be at the new line index %d, got %d", initialLinesCount, newModel.cursor)
		}
	})

	t.Run("DeleteLine", func(t *testing.T) {
		settings := types.DefaultSettings()
		settings.Lines = append(settings.Lines, []types.WidgetItem{})
		initialLinesCount := len(settings.Lines)
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "lines"
		m.cursor = initialLinesCount - 1

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
		updatedModel, _ := m.Update(msg)
		newModel := updatedModel.(Model)

		if len(newModel.settings.Lines) != initialLinesCount-1 {
			t.Errorf("Expected lines count to be %d after deletion, got %d", initialLinesCount-1, len(newModel.settings.Lines))
		}
		if newModel.cursor != initialLinesCount-2 {
			t.Errorf("Expected cursor to adjust to %d, got %d", initialLinesCount-2, newModel.cursor)
		}
	})

	t.Run("CannotDeleteLastLine", func(t *testing.T) {
		settings := types.DefaultSettings()
		settings.Lines = [][]types.WidgetItem{
			{{Type: "model"}},
		}
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "lines"
		m.cursor = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
		updatedModel, _ := m.Update(msg)
		newModel := updatedModel.(Model)

		if len(newModel.settings.Lines) != 1 {
			t.Errorf("Expected lines count to remain 1, got %d", len(newModel.settings.Lines))
		}
	})

	t.Run("MoveLine", func(t *testing.T) {
		settings := types.DefaultSettings()
		settings.Lines = [][]types.WidgetItem{
			{{Type: "model"}},
			{},
		}
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "lines"
		m.cursor = 1

		// Switch to move mode ("m")
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
		newModel := updatedModel.(Model)
		if !newModel.moveMode {
			t.Fatalf("Expected moveMode to be true")
		}

		// Move up (swap line 1 and line 0)
		updatedModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyUp})
		newModel = updatedModel.(Model)

		if len(newModel.settings.Lines[0]) != 0 {
			t.Errorf("Expected line 0 to be empty after swapping, but got %d widgets", len(newModel.settings.Lines[0]))
		}
		if len(newModel.settings.Lines[1]) == 0 {
			t.Errorf("Expected line 1 to contain widgets after swapping, but it was empty")
		}
		if newModel.cursor != 0 {
			t.Errorf("Expected cursor to follow the moved item to 0, got %d", newModel.cursor)
		}

		// Toggle moveMode off using Enter
		updatedModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		newModel = updatedModel.(Model)
		if newModel.moveMode {
			t.Errorf("Expected moveMode to be false after pressing Enter")
		}
	})

	t.Run("NavigateToItems", func(t *testing.T) {
		settings := types.DefaultSettings()
		settings.Lines = append(settings.Lines, []types.WidgetItem{})
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "lines"
		m.cursor = 1

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		newModel := updatedModel.(Model)
		if newModel.activeMenu != "items" {
			t.Errorf("Expected activeMenu to transition to 'items', got %q", newModel.activeMenu)
		}
		if newModel.selectedLine != 1 {
			t.Errorf("Expected selectedLine to be 1, got %d", newModel.selectedLine)
		}
	})

	t.Run("BackToMain", func(t *testing.T) {
		settings := types.DefaultSettings()
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "lines"

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		newModel := updatedModel.(Model)
		if newModel.activeMenu != "main" {
			t.Errorf("Expected activeMenu to go back to 'main', got %q", newModel.activeMenu)
		}
	})
}

func TestTUI_ItemsOperations(t *testing.T) {
	baseSettings := func() types.Settings {
		s := types.DefaultSettings()
		s.Lines[0] = []types.WidgetItem{
			{Type: "model", Color: "brightMagenta"},
			{Type: "sandbox", Color: "brightBlack"},
			{Type: "git-branch", Color: "brightMagenta"},
			{Type: "git-changes", Color: "yellow"},
		}
		return s
	}

	t.Run("DeleteWidget", func(t *testing.T) {
		m := NewModel(baseSettings(), "/tmp/settings.toml")
		m.activeMenu = "items"
		m.selectedLine = 0
		m.cursor = 1 // Pointing to sandbox

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
		updatedModel, _ := m.Update(msg)
		newModel := updatedModel.(Model)

		if len(newModel.settings.Lines[0]) != 3 {
			t.Errorf("Expected widget count to be 3 after deletion, got %d", len(newModel.settings.Lines[0]))
		}
		if newModel.settings.Lines[0][1].Type != "git-branch" {
			t.Errorf("Expected widget at index 1 to be 'git-branch', got %q", newModel.settings.Lines[0][1].Type)
		}
	})

	t.Run("MoveWidget", func(t *testing.T) {
		m := NewModel(baseSettings(), "/tmp/settings.toml")
		m.activeMenu = "items"
		m.selectedLine = 0
		m.cursor = 0 // Pointing to model

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
		newModel := updatedModel.(Model)
		if !newModel.moveMode {
			t.Fatalf("Expected moveMode to be true")
		}

		// Move down (swap index 0 and 1)
		downMsg := tea.KeyMsg{Type: tea.KeyDown}
		updatedModel, _ = newModel.Update(downMsg)
		newModel = updatedModel.(Model)

		if newModel.settings.Lines[0][0].Type != "sandbox" {
			t.Errorf("Expected widget at index 0 to be 'sandbox' after swapping, got %q", newModel.settings.Lines[0][0].Type)
		}
		if newModel.settings.Lines[0][1].Type != "model" {
			t.Errorf("Expected widget at index 1 to be 'model' after swapping, got %q", newModel.settings.Lines[0][1].Type)
		}
		if newModel.cursor != 1 {
			t.Errorf("Expected cursor to follow the item to index 1, got %d", newModel.cursor)
		}
	})

	t.Run("BackToLines", func(t *testing.T) {
		m := NewModel(baseSettings(), "/tmp/settings.toml")
		m.activeMenu = "items"
		m.selectedLine = 0
		m.moveMode = false

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		newModel := updatedModel.(Model)
		if newModel.activeMenu != "lines" {
			t.Errorf("Expected activeMenu to go back to 'lines', got %q", newModel.activeMenu)
		}
		if newModel.cursor != 0 {
			t.Errorf("Expected cursor in lines menu to be the selected line (0), got %d", newModel.cursor)
		}
	})
}

func TestTUI_AddWidget(t *testing.T) {
	t.Run("OpenAddWidgetScreen", func(t *testing.T) {
		settings := types.DefaultSettings()
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "items"
		m.selectedLine = 0
		m.cursor = 1

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
		updatedModel, _ := m.Update(msg)
		newModel := updatedModel.(Model)

		if newModel.activeMenu != "add_widget" {
			t.Errorf("Expected activeMenu to transition to 'add_widget', got %q", newModel.activeMenu)
		}
		if newModel.itemIndex != 1 {
			t.Errorf("Expected itemIndex to save previous cursor (1), got %d", newModel.itemIndex)
		}
		if newModel.cursor != 0 {
			t.Errorf("Expected cursor to reset to 0, got %d", newModel.cursor)
		}
	})

	t.Run("InsertWidget", func(t *testing.T) {
		settings := types.DefaultSettings()
		initialLen := len(settings.Lines[0])
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "add_widget"
		m.selectedLine = 0
		m.itemIndex = 1
		m.cursor = 2 // context-bar

		// Press Enter to add
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		newModel := updatedModel.(Model)

		if newModel.activeMenu != "items" {
			t.Errorf("Expected activeMenu to return to 'items', got %q", newModel.activeMenu)
		}
		if len(newModel.settings.Lines[0]) != initialLen+1 {
			t.Errorf("Expected %d widgets in line 0, got %d", initialLen+1, len(newModel.settings.Lines[0]))
		}
		if newModel.settings.Lines[0][2].Type != "context-bar" {
			t.Errorf("Expected added widget at index 2 to be 'context-bar', got %q", newModel.settings.Lines[0][2].Type)
		}
		if newModel.cursor != 2 {
			t.Errorf("Expected cursor to point to the newly added widget (index 2), got %d", newModel.cursor)
		}
	})
}

func TestTUI_AddQuotaBarWidgets(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()

	t.Run("NavigateToAddWidget", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "items"
		m.selectedLine = 0
		m.cursor = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
		updatedModel, _ := m.Update(msg)
		newModel := updatedModel.(Model)

		if newModel.activeMenu != "add_widget" {
			t.Fatalf("Expected activeMenu to be %q, got %q", "add_widget", newModel.activeMenu)
		}
	})

	t.Run("PresetPresence", func(t *testing.T) {
		requiredPresets := []string{
			"quota-bar-5h",
			"quota-bar-7d",
			"quota-bar-argon-7d",
			"quota-bar-3p-5h",
			"quota-bar-3p-7d",
		}
		presetMap := make(map[string]bool, len(widgetTypes))
		for _, wt := range widgetTypes {
			presetMap[wt.item.Type] = true
		}
		for _, expected := range requiredPresets {
			if !presetMap[expected] {
				t.Errorf("Expected preset %q to be present in widgetTypes", expected)
			}
		}
	})

	t.Run("AddQuotaBarWidget", func(t *testing.T) {
		targetBarIdx := -1
		for i, wt := range widgetTypes {
			if wt.item.Type == "quota-bar-5h" {
				targetBarIdx = i
				break
			}
		}
		if targetBarIdx == -1 {
			t.Fatalf("Gemini 5h quota-bar widget type not found in widgetTypes")
		}

		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "add_widget"
		m.selectedLine = 0
		m.itemIndex = 0
		m.cursor = targetBarIdx

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		finalBarModel := updatedModel.(Model)

		if finalBarModel.activeMenu != "items" {
			t.Fatalf("Expected activeMenu to return to %q, got %q", "items", finalBarModel.activeMenu)
		}

		addedBarWidget := finalBarModel.settings.Lines[0][1]
		if addedBarWidget.Type != "quota-bar-5h" {
			t.Errorf("Expected widget type 'quota-bar-5h', got %q", addedBarWidget.Type)
		}
	})
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

func TestTUI_PowerlineSeparator_IndexInitialization(t *testing.T) {
	tests := []struct {
		name          string
		setupSettings func() types.Settings
		wantSepName   string
		wantSepValue  string
	}{
		{
			name: "Default Separator",
			setupSettings: func() types.Settings {
				return types.DefaultSettings()
			},
			wantSepName:  "Arrow (\uE0B0)",
			wantSepValue: "\uE0B0",
		},
		{
			name: "Preset Separator (exists in list)",
			setupSettings: func() types.Settings {
				s := types.DefaultSettings()
				s.Powerline.Separator = "\uE0B4"
				return s
			},
			wantSepName:  "Round (\uE0B4)",
			wantSepValue: "\uE0B4",
		},
		{
			name: "None Separator",
			setupSettings: func() types.Settings {
				s := types.DefaultSettings()
				s.Powerline.Separator = ""
				return s
			},
			wantSepName:  "None",
			wantSepValue: "",
		},
		{
			name: "None Separator with Powerline Enabled",
			setupSettings: func() types.Settings {
				s := types.DefaultSettings()
				s.Powerline.Enabled = true
				s.Powerline.Separator = ""
				return s
			},
			wantSepName:  "None",
			wantSepValue: "",
		},
		{
			name: "Custom Separator (NOT in list)",
			setupSettings: func() types.Settings {
				s := types.DefaultSettings()
				s.Powerline.Separator = "♦"
				return s
			},
			wantSepName:  "Custom (♦)",
			wantSepValue: "♦",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.setupSettings()
			m := NewModel(settings, "/tmp/settings.toml")

			if m.separatorIndex < 0 || m.separatorIndex >= len(separatorsList) {
				t.Fatalf("Expected valid separatorIndex [0, %d), got %d", len(separatorsList), m.separatorIndex)
			}
			gotSep := separatorsList[m.separatorIndex]
			if gotSep.name != tt.wantSepName {
				t.Errorf("Expected separator name %q, got %q", tt.wantSepName, gotSep.name)
			}
			if gotSep.value != tt.wantSepValue {
				t.Errorf("Expected separator value %q, got %q", tt.wantSepValue, gotSep.value)
			}
		})
	}
}

func TestTUI_PowerlineSeparator_PresetIndexes(t *testing.T) {
	tests := []struct {
		name      string
		separator string
		wantIndex int
	}{
		{name: "None", separator: "", wantIndex: 0},
		{name: "Arrow", separator: "\uE0B0", wantIndex: 1},
		{name: "Round", separator: "\uE0B4", wantIndex: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := types.DefaultSettings()
			s.Powerline.Separator = tt.separator
			m := NewModel(s, "/tmp/settings.toml")
			if m.separatorIndex != tt.wantIndex {
				t.Errorf("Expected separatorIndex %d, got %d", tt.wantIndex, m.separatorIndex)
			}
		})
	}
}

func TestTUI_PowerlineSeparator_Navigation(t *testing.T) {
	s := types.DefaultSettings()
	s.Powerline.Separator = "♦"
	m := NewModel(s, "/tmp/settings.toml")
	m.activeMenu = "powerline"
	m.cursor = 2 // Powerline Separator option

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")}
	updatedModel, _ := m.Update(keyMsg)
	newModel := updatedModel.(Model)

	if newModel.activeMenu != "select_separator" {
		t.Errorf("Expected activeMenu to transition to 'select_separator', got %q", newModel.activeMenu)
	}
	if newModel.cursor != m.separatorIndex {
		t.Errorf("Expected cursor to be %d, got %d", m.separatorIndex, newModel.cursor)
	}
}

func TestTUI_PowerlineSubmenu(t *testing.T) {
	settings := types.DefaultSettings()

	t.Run("NavigateToSubmenu", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.cursor = 1 // Powerline Settings
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mPowerline := updatedModel.(Model)

		if mPowerline.activeMenu != "powerline" {
			t.Fatalf("Expected activeMenu to be 'powerline', got %q", mPowerline.activeMenu)
		}
		if mPowerline.cursor != 0 {
			t.Errorf("Expected cursor in powerline menu to start at 0, got %d", mPowerline.cursor)
		}
	})

	t.Run("EscBackToMain", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "powerline"
		m.cursor = 0

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		mMain := updatedModel.(Model)
		if mMain.activeMenu != "main" {
			t.Errorf("Expected activeMenu to return to 'main' on Esc, got %q", mMain.activeMenu)
		}
		if mMain.cursor != 1 {
			t.Errorf("Expected main menu cursor to be 1, got %d", mMain.cursor)
		}
	})

	t.Run("TogglePowerlineMode", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "powerline"
		m.cursor = 0
		initialEnabled := m.settings.Powerline.Enabled

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mToggle := updatedModel.(Model)
		if mToggle.settings.Powerline.Enabled != !initialEnabled {
			t.Errorf("Expected Powerline.Enabled to toggle from %v to %v", initialEnabled, !initialEnabled)
		}
	})

	t.Run("BackToMainMenuOption", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "powerline"
		m.cursor = 5 // "Back to Main Menu"

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mBack := updatedModel.(Model)
		if mBack.activeMenu != "main" {
			t.Errorf("Expected activeMenu to return to 'main' on 'Back to Main Menu', got %q", mBack.activeMenu)
		}
		if mBack.cursor != 1 {
			t.Errorf("Expected main menu cursor to be 1, got %d", mBack.cursor)
		}
	})

	t.Run("ViewRendering", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "powerline"

		viewStr := m.View()
		if !strings.Contains(viewStr, "Powerline Settings") || !strings.Contains(viewStr, "Toggle Powerline Mode") {
			t.Errorf("Expected viewPowerline output to contain Powerline Settings menu title and items, got:\n%s", viewStr)
		}
	})

	t.Run("CursorBoundaries", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "powerline"

		// Upper boundary at 0 with Up and 'k'
		m.cursor = 0
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
		if updatedModel.(Model).cursor != 0 {
			t.Errorf("Expected cursor to remain 0 on Up at upper bound, got %d", updatedModel.(Model).cursor)
		}

		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		if updatedModel.(Model).cursor != 0 {
			t.Errorf("Expected cursor to remain 0 on 'k' at upper bound, got %d", updatedModel.(Model).cursor)
		}

		// Down navigation with Down and 'j'
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		mNav := updatedModel.(Model)
		if mNav.cursor != 1 {
			t.Errorf("Expected cursor to move to 1 on Down key, got %d", mNav.cursor)
		}

		updatedModel, _ = mNav.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		mNav = updatedModel.(Model)
		if mNav.cursor != 2 {
			t.Errorf("Expected cursor to move to 2 on 'j' key, got %d", mNav.cursor)
		}

		// Lower boundary at 5 with Down and 'j'
		mNav.cursor = 5
		updatedModel, _ = mNav.Update(tea.KeyMsg{Type: tea.KeyDown})
		if updatedModel.(Model).cursor != 5 {
			t.Errorf("Expected cursor to remain 5 on Down at lower bound, got %d", updatedModel.(Model).cursor)
		}

		updatedModel, _ = mNav.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		if updatedModel.(Model).cursor != 5 {
			t.Errorf("Expected cursor to remain 5 on 'j' at lower bound, got %d", updatedModel.(Model).cursor)
		}
	})
}

func TestTUI_SelectThemeMenu(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	settings.Powerline.Theme = "nord"

	t.Run("OpenThemeMenu", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "powerline"
		m.cursor = 1 // Select Powerline Theme

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mTheme := updatedModel.(Model)

		if mTheme.activeMenu != "select_theme" {
			t.Fatalf("Expected activeMenu to transition to 'select_theme', got %q", mTheme.activeMenu)
		}
		if mTheme.cursor != 0 {
			t.Errorf("Expected cursor to start at current theme index 0, got %d", mTheme.cursor)
		}

		// Move cursor down (nord -> nord-aurora)
		updatedModel, _ = mTheme.Update(tea.KeyMsg{Type: tea.KeyDown})
		mTheme = updatedModel.(Model)
		if mTheme.cursor != 1 {
			t.Errorf("Expected cursor to move to 1, got %d", mTheme.cursor)
		}
	})

	t.Run("DynamicPreviewUpdate", func(t *testing.T) {
		mTheme0 := NewModel(settings, "/tmp/settings.toml")
		mTheme0.activeMenu = "select_theme"
		mTheme0.cursor = 0
		viewNord := mTheme0.View()

		mTheme1 := NewModel(settings, "/tmp/settings.toml")
		mTheme1.activeMenu = "select_theme"
		mTheme1.cursor = 1
		viewNordAurora := mTheme1.View()

		linesNord := strings.Split(viewNord, "\n")
		linesNordAurora := strings.Split(viewNordAurora, "\n")

		if len(linesNord) < 5 || len(linesNordAurora) < 5 {
			t.Fatalf("Expected view outputs to have enough lines")
		}
		previewNordPart := strings.Join(linesNord[:4], "\n")
		previewNordAuroraPart := strings.Join(linesNordAurora[:4], "\n")

		if previewNordPart == previewNordAuroraPart {
			t.Errorf("Expected Live Preview to be updated dynamically for theme cursor")
		}
	})

	t.Run("ConfirmTheme", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_theme"
		m.cursor = 1 // nord-aurora

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mConfirmed := updatedModel.(Model)

		if mConfirmed.activeMenu != "powerline" {
			t.Errorf("Expected activeMenu to return to 'powerline', got %q", mConfirmed.activeMenu)
		}
		if mConfirmed.cursor != 1 {
			t.Errorf("Expected powerline menu cursor to be 1, got %d", mConfirmed.cursor)
		}
		if mConfirmed.settings.Powerline.Theme != "nord-aurora" {
			t.Errorf("Expected theme to be 'nord-aurora', got %q", mConfirmed.settings.Powerline.Theme)
		}
		if mConfirmed.themeIndex != 1 {
			t.Errorf("Expected themeIndex to be updated to 1, got %d", mConfirmed.themeIndex)
		}
	})

	t.Run("CancelTheme", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_theme"
		m.cursor = 1 // nord-aurora

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		mCancelled := updatedModel.(Model)

		if mCancelled.activeMenu != "powerline" {
			t.Errorf("Expected activeMenu to return to 'powerline' on Esc, got %q", mCancelled.activeMenu)
		}
		if mCancelled.cursor != 1 {
			t.Errorf("Expected powerline menu cursor to return to 1, got %d", mCancelled.cursor)
		}
		if mCancelled.settings.Powerline.Theme != "nord" {
			t.Errorf("Expected theme to remain 'nord', got %q", mCancelled.settings.Powerline.Theme)
		}
	})
}

func TestTUI_SelectSeparatorMenu(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	settings.Powerline.Separator = "\uE0B0"

	t.Run("OpenSeparatorMenu", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "powerline"
		m.cursor = 2 // Select Powerline Separator

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mSep := updatedModel.(Model)

		if mSep.activeMenu != "select_separator" {
			t.Fatalf("Expected activeMenu to transition to 'select_separator', got %q", mSep.activeMenu)
		}
		if mSep.cursor != 1 {
			t.Errorf("Expected cursor to start at current separator index 1, got %d", mSep.cursor)
		}

		// Move cursor down (Arrow -> Round, index 2)
		updatedModel, _ = mSep.Update(tea.KeyMsg{Type: tea.KeyDown})
		mSep = updatedModel.(Model)
		if mSep.cursor != 2 {
			t.Errorf("Expected cursor to move to 2, got %d", mSep.cursor)
		}
	})

	t.Run("ViewRendering", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_separator"
		m.cursor = 2

		viewSepSelected := m.View()
		if !strings.Contains(viewSepSelected, "Select Powerline Separator") {
			t.Errorf("Expected viewSelectSeparator to render title 'Select Powerline Separator', got:\n%s", viewSepSelected)
		}
	})

	t.Run("ConfirmSeparator", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_separator"
		m.cursor = 2 // Round

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mConfirmed := updatedModel.(Model)

		if mConfirmed.activeMenu != "powerline" {
			t.Errorf("Expected activeMenu to return to 'powerline', got %q", mConfirmed.activeMenu)
		}
		if mConfirmed.cursor != 2 {
			t.Errorf("Expected powerline menu cursor to be 2, got %d", mConfirmed.cursor)
		}
	})

	t.Run("CancelSeparator", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_separator"
		m.cursor = 2 // Round

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		mCancelled := updatedModel.(Model)

		if mCancelled.activeMenu != "powerline" {
			t.Errorf("Expected activeMenu to return to 'powerline' on Esc, got %q", mCancelled.activeMenu)
		}
		if mCancelled.cursor != 2 {
			t.Errorf("Expected powerline menu cursor to return to 2, got %d", mCancelled.cursor)
		}
	})
}

func TestTUI_SelectStartCapMenu(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	settings.Powerline.StartCaps = "\uE0B2"

	t.Run("OpenAndConfirm", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "powerline"
		m.cursor = 3 // Select Powerline Start Cap

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mCap := updatedModel.(Model)

		if mCap.activeMenu != "select_start_cap" {
			t.Fatalf("Expected activeMenu to transition to 'select_start_cap', got %q", mCap.activeMenu)
		}

		updatedModel, _ = mCap.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mConfirmed := updatedModel.(Model)

		if mConfirmed.activeMenu != "powerline" {
			t.Errorf("Expected activeMenu to return to 'powerline', got %q", mConfirmed.activeMenu)
		}
		if mConfirmed.cursor != 3 {
			t.Errorf("Expected powerline menu cursor to be 3, got %d", mConfirmed.cursor)
		}
	})

	t.Run("Cancel", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_start_cap"
		m.cursor = 1

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		mCancelled := updatedModel.(Model)

		if mCancelled.activeMenu != "powerline" {
			t.Errorf("Expected activeMenu to return to 'powerline' on Esc, got %q", mCancelled.activeMenu)
		}
		if mCancelled.cursor != 3 {
			t.Errorf("Expected powerline menu cursor to return to 3, got %d", mCancelled.cursor)
		}
	})
}

func TestTUI_SelectEndCapMenu(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	settings.Powerline.EndCaps = "\uE0B0"

	t.Run("OpenAndConfirm", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "powerline"
		m.cursor = 4 // Select Powerline End Cap

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mCap := updatedModel.(Model)

		if mCap.activeMenu != "select_end_cap" {
			t.Fatalf("Expected activeMenu to transition to 'select_end_cap', got %q", mCap.activeMenu)
		}

		updatedModel, _ = mCap.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mConfirmed := updatedModel.(Model)

		if mConfirmed.activeMenu != "powerline" {
			t.Errorf("Expected activeMenu to return to 'powerline', got %q", mConfirmed.activeMenu)
		}
		if mConfirmed.cursor != 4 {
			t.Errorf("Expected powerline menu cursor to be 4, got %d", mConfirmed.cursor)
		}
	})

	t.Run("Cancel", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_end_cap"
		m.cursor = 1

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		mCancelled := updatedModel.(Model)

		if mCancelled.activeMenu != "powerline" {
			t.Errorf("Expected activeMenu to return to 'powerline' on Esc, got %q", mCancelled.activeMenu)
		}
		if mCancelled.cursor != 4 {
			t.Errorf("Expected powerline menu cursor to return to 4, got %d", mCancelled.cursor)
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

func TestTUI_SelectColorLevel(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	settings.General.ColorLevel = 2 // ANSI 256 colors

	t.Run("InitialState", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		if m.colorLevelIndex != 1 {
			t.Errorf("Expected initial colorLevelIndex to be 1 (ANSI 256 Colors), got %d", m.colorLevelIndex)
		}
	})

	t.Run("OpenColorLevelMenu", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.cursor = 2 // Select Color Level
		msgEnter := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")}
		updatedModel, _ := m.Update(msgEnter)
		mSelect := updatedModel.(Model)

		if mSelect.activeMenu != "select_color_level" {
			t.Errorf("Expected activeMenu to transition to 'select_color_level', got %q", mSelect.activeMenu)
		}
		if mSelect.cursor != 1 {
			t.Errorf("Expected cursor in sub-menu to be 1, got %d", mSelect.cursor)
		}
	})

	t.Run("SelectTruecolor", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_color_level"
		m.cursor = 2 // Truecolor (24-bit)

		msgEnter := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")}
		updatedModel, _ := m.Update(msgEnter)
		mSelected := updatedModel.(Model)

		if mSelected.activeMenu != "main" {
			t.Errorf("Expected activeMenu to return to 'main', got %q", mSelected.activeMenu)
		}
		if mSelected.settings.General.ColorLevel != 3 {
			t.Errorf("Expected settings.General.ColorLevel to be updated to 3, got %d", mSelected.settings.General.ColorLevel)
		}
		if mSelected.colorLevelIndex != 2 {
			t.Errorf("Expected colorLevelIndex to be updated to 2, got %d", mSelected.colorLevelIndex)
		}
		if mSelected.cursor != 2 {
			t.Errorf("Expected cursor in main menu to remain at 2, got %d", mSelected.cursor)
		}
	})

	t.Run("CancelWithEsc", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_color_level"
		m.cursor = 0 // ANSI 16 colors

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		mCancelled := updatedModel.(Model)

		if mCancelled.activeMenu != "main" {
			t.Errorf("Expected activeMenu to return to 'main' on Esc, got %q", mCancelled.activeMenu)
		}
		if mCancelled.settings.General.ColorLevel != 2 {
			t.Errorf("Expected settings.General.ColorLevel to remain 2, got %d", mCancelled.settings.General.ColorLevel)
		}
	})

	t.Run("LivePreview", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_color_level"
		m.cursor = 0 // ANSI 16 colors (should trigger 16-color rendering in preview)

		viewStr := m.View()
		if viewStr == "" {
			t.Errorf("Expected non-empty view string")
		}
	})
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

func TestTUI_WidgetSliceCorruption(t *testing.T) {
	widgets.RegisterAll()

	// 1. Setup a line with multiple widgets, ensuring capacity is larger than length.
	// This simulates slice sharing capacity.
	originalWidgets := []types.WidgetItem{
		{Type: "model"},
		{Type: "git-changes"},
		{Type: "git-branch"},
	}
	widgetsSlice := make([]types.WidgetItem, 3, 10)
	copy(widgetsSlice, originalWidgets)

	settings := types.DefaultSettings()
	settings.Lines = [][]types.WidgetItem{widgetsSlice}

	m := NewModel(settings, "/tmp/settings.toml")
	m.selectedLine = 0
	m.itemIndex = 0 // Insert after index 0 (so between model and git-changes)

	// --- Test 1: Live Preview should not corrupt original settings ---
	m.activeMenu = "add_widget"
	m.cursor = 0 // first widget type to add

	// Verify pre-conditions
	if m.settings.Lines[0][1].Type != "git-changes" || m.settings.Lines[0][2].Type != "git-branch" {
		t.Fatalf("Pre-condition failed: settings initialized incorrectly")
	}

	// Trigger preview render (which calls View and performs a temporary insert)
	_ = m.View()

	// Verify that the original settings line was not mutated by the preview logic
	if len(m.settings.Lines[0]) != 3 {
		t.Errorf("Expected original settings line length to remain 3 after preview, but got %d", len(m.settings.Lines[0]))
	}
	if m.settings.Lines[0][1].Type != "git-changes" {
		t.Errorf("Expected original widget 'git-changes' to be untouched, but got %q", m.settings.Lines[0][1].Type)
	}
	if m.settings.Lines[0][2].Type != "git-branch" {
		t.Errorf("Expected original widget 'git-branch' to be untouched, but got %q", m.settings.Lines[0][2].Type)
	}

	// --- Test 2: Actually adding the widget should correctly insert it without corruption ---
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := m.Update(enterMsg)
	newModel := updatedModel.(Model)

	newWidgets := newModel.settings.Lines[0]
	if len(newWidgets) != 4 {
		t.Fatalf("Expected new settings line length to be 4, got %d", len(newWidgets))
	}

	// Expected sequence: model -> newly_added -> git-changes -> git-branch
	if newWidgets[0].Type != "model" {
		t.Errorf("Expected index 0 to be 'model', got %q", newWidgets[0].Type)
	}
	if newWidgets[2].Type != "git-changes" {
		t.Errorf("Expected index 2 to be 'git-changes', got %q", newWidgets[2].Type)
	}
	if newWidgets[3].Type != "git-branch" {
		t.Errorf("Expected index 3 to be 'git-branch', got %q", newWidgets[3].Type)
	}

	// Check that the underlying original widgets slice was not mutated (no in-place overwrite)
	if widgetsSlice[1].Type != "git-changes" {
		t.Errorf("Expected original widgetsSlice elements to remain untouched, but index 1 got %q", widgetsSlice[1].Type)
	}
}

func TestTUI_NoASCIIInSeparators(t *testing.T) {
	for _, sep := range separatorsList {
		if strings.ContainsAny(sep.value, "/|") {
			t.Errorf("Separator %q contains forbidden ASCII characters: %q", sep.name, sep.value)
		}
	}
}

func TestTUI_SeparatorsList_Presets(t *testing.T) {
	expected := []struct {
		name  string
		value string
	}{
		{name: "None", value: ""},
		{name: "Arrow (\uE0B0)", value: "\uE0B0"},
		{name: "Round (\uE0B4)", value: "\uE0B4"},
		{name: "Flame (\uE0C0)", value: "\uE0C0"},
		{name: "Hexagon (\uE0C6)", value: "\uE0C6"},
		{name: "Slanted (\uE0C8)", value: "\uE0C8"},
	}

	if len(separatorsList) < len(expected) {
		t.Fatalf("Expected at least %d preset separators, got %d", len(expected), len(separatorsList))
	}

	for i, want := range expected {
		t.Run(want.name, func(t *testing.T) {
			got := separatorsList[i]
			if got.name != want.name {
				t.Errorf("Expected preset name %q, got %q", want.name, got.name)
			}
			if got.value != want.value {
				t.Errorf("Expected preset value %q, got %q", want.value, got.value)
			}
		})
	}
}

func TestTUI_WidgetTypesOrdering(t *testing.T) {
	expectedTypes := []string{
		"agent-state",
		"model",
		"context-bar",
		"artifacts",
		"subagents",
		"tasks",
		"sandbox",
		"git-repo",
		"git-branch",
		"git-ahead-behind",
		"git-changes",
		"quota-all",
		"quota-5h",
		"quota-7d",
		"quota-argon-7d",
		"quota-3p-5h",
		"quota-3p-7d",
		"quota-bar-all",
		"quota-bar-5h",
		"quota-bar-7d",
		"quota-bar-argon-7d",
		"quota-bar-3p-5h",
		"quota-bar-3p-7d",
		"custom-text",
	}

	if len(widgetTypes) != len(expectedTypes) {
		t.Fatalf("Expected %d widget types, got %d", len(expectedTypes), len(widgetTypes))
	}

	for i, expected := range expectedTypes {
		if widgetTypes[i].item.Type != expected {
			t.Errorf("At index %d: expected widget type %q, got %q", i, expected, widgetTypes[i].item.Type)
		}
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

func TestSaveSettings_Success(t *testing.T) {
	settings := types.DefaultSettings()
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "config.toml")

	err := saveSettings(targetPath, settings)
	if err != nil {
		t.Fatalf("saveSettings() failed unexpectedly: %v", err)
	}

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Errorf("Expected settings file to exist at %q", targetPath)
	}

	// Verify temporary files are cleaned up
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read tempDir: %v", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".tmp") {
			t.Errorf("Expected no temporary files remaining in %q, but found %q", tempDir, f.Name())
		}
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

func TestTUI_InitAndMainChoices(t *testing.T) {
	settings := types.DefaultSettings()

	t.Run("Init", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		cmd := m.Init()
		if cmd != nil {
			t.Errorf("Expected Init() to return nil, got %v", cmd)
		}
	})

	t.Run("SaveAndExit", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.cursor = 3 // Save & Exit
		tempDir := t.TempDir()
		m.configPath = filepath.Join(tempDir, "settings.toml")

		updatedModel, quitCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mSave := updatedModel.(Model)
		if !mSave.saved {
			t.Errorf("Expected saved to be true after selecting Save & Exit")
		}
		if !mSave.quitting {
			t.Errorf("Expected quitting to be true after selecting Save & Exit")
		}
		if quitCmd == nil {
			t.Errorf("Expected Quit command after selecting Save & Exit")
		}
	})

	t.Run("DiscardAndExit", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.cursor = 4 // Discard & Exit
		m.saved = false

		updatedModel, quitCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mDiscard := updatedModel.(Model)
		if mDiscard.saved {
			t.Errorf("Expected saved to be false after selecting Discard & Exit")
		}
		if !mDiscard.quitting {
			t.Errorf("Expected quitting to be true after selecting Discard & Exit")
		}
		if quitCmd == nil {
			t.Errorf("Expected Quit command after selecting Discard & Exit")
		}
	})
}

func TestTUI_UpdateMain_CursorNavigation(t *testing.T) {
	settings := types.DefaultSettings()

	tests := []struct {
		name        string
		initialCur  int
		key         tea.KeyMsg
		expectedCur int
	}{
		{
			name:        "Up key at upper bound (cursor 0)",
			initialCur:  0,
			key:         tea.KeyMsg{Type: tea.KeyUp},
			expectedCur: 0,
		},
		{
			name:        "k key at upper bound (cursor 0)",
			initialCur:  0,
			key:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")},
			expectedCur: 0,
		},
		{
			name:        "Down key at lower bound (cursor 4)",
			initialCur:  4,
			key:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCur: 4,
		},
		{
			name:        "j key at lower bound (cursor 4)",
			initialCur:  4,
			key:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")},
			expectedCur: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(settings, "/tmp/settings.toml")
			m.cursor = tt.initialCur
			updatedModel, _ := m.Update(tt.key)
			mResult := updatedModel.(Model)
			if mResult.cursor != tt.expectedCur {
				t.Errorf("Expected cursor %d, got %d", tt.expectedCur, mResult.cursor)
			}
		})
	}
}

func TestTUI_UpdateMain_SubmenuNavigation(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")
	m.cursor = 1

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mSubmenu := updatedModel.(Model)
	if mSubmenu.activeMenu != "powerline" {
		t.Errorf("Expected activeMenu to be 'powerline', got %q", mSubmenu.activeMenu)
	}
}

func TestTUI_UpdateMain_SaveFailure(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")
	m.cursor = 3

	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "invalid_parent_file")
	if err := os.WriteFile(invalidFile, []byte("data"), 0644); err != nil {
		t.Fatalf("Failed to create invalid parent file: %v", err)
	}
	m.configPath = filepath.Join(invalidFile, "settings.toml")

	updatedModel, quitCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mSaveErr := updatedModel.(Model)
	if mSaveErr.saved {
		t.Errorf("Expected saved to be false on save error")
	}
	if !mSaveErr.quitting {
		t.Errorf("Expected quitting to be true on save error exit")
	}
	if quitCmd == nil {
		t.Errorf("Expected Quit command when Save & Exit fails")
	}
}

func TestTUI_UpdateMain_NonKeyMessages(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")

	_, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if cmd != nil {
		t.Errorf("Expected nil command for non-KeyMsg input, got %v", cmd)
	}
}

func TestTUI_UpdateMain_UnknownMenu(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")
	m.activeMenu = "unknown_menu_type"

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	mResult := updatedModel.(Model)
	if mResult.activeMenu != "unknown_menu_type" {
		t.Errorf("Expected activeMenu to remain 'unknown_menu_type', got %q", mResult.activeMenu)
	}
	if cmd != nil {
		t.Errorf("Expected nil command for unknown activeMenu, got %v", cmd)
	}
}

func TestTUI_UpdateLines_BoundaryAndMoveMode(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")
	m.activeMenu = "lines"

	// 1. Move up boundary in moveMode (cursor == 0)
	m.cursor = 0
	m.moveMode = true
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	mResult := updatedModel.(Model)
	if mResult.cursor != 0 {
		t.Errorf("Expected cursor to remain 0 when moving up in moveMode at top, got %d", mResult.cursor)
	}

	// 2. Move up/down in moveMode with single line only
	singleLineSettings := types.DefaultSettings()
	singleLineSettings.Lines = [][]types.WidgetItem{{}}
	mSingle := NewModel(singleLineSettings, "/tmp/settings.toml")
	mSingle.activeMenu = "lines"
	mSingle.cursor = 0
	mSingle.moveMode = true

	updatedModel, _ = mSingle.Update(tea.KeyMsg{Type: tea.KeyUp})
	if updatedModel.(Model).cursor != 0 {
		t.Errorf("Expected cursor to remain 0 for single line move up")
	}

	updatedModel, _ = mSingle.Update(tea.KeyMsg{Type: tea.KeyDown})
	if updatedModel.(Model).cursor != 0 {
		t.Errorf("Expected cursor to remain 0 for single line move down")
	}

	// 3. Move up boundary when moveMode == false (cursor == 0)
	m.moveMode = false
	m.cursor = 0
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if updatedModel.(Model).cursor != 0 {
		t.Errorf("Expected cursor to remain 0 when moveMode=false at top")
	}

	// 4. Move down boundary in moveMode (cursor == len(lines)-1)
	linesCount := len(m.settings.Lines)
	m.cursor = linesCount - 1
	m.moveMode = true
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if updatedModel.(Model).cursor != linesCount-1 {
		t.Errorf("Expected cursor to remain %d when moving down in moveMode at bottom", linesCount-1)
	}

	// 5. Move down boundary when moveMode == false (cursor == len(lines)-1)
	m.moveMode = false
	m.cursor = linesCount - 1
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if updatedModel.(Model).cursor != linesCount-1 {
		t.Errorf("Expected cursor to remain %d when moveMode=false at bottom", linesCount-1)
	}

	// 6. Move down swap in moveMode (valid swap)
	multiLineSettings := types.DefaultSettings()
	multiLineSettings.Lines = [][]types.WidgetItem{
		{{Type: "model"}},
		{{Type: "sandbox"}},
	}
	mMulti := NewModel(multiLineSettings, "/tmp/settings.toml")
	mMulti.activeMenu = "lines"
	mMulti.cursor = 0
	mMulti.moveMode = true
	updatedModel, _ = mMulti.Update(tea.KeyMsg{Type: tea.KeyDown})
	mSwapped := updatedModel.(Model)
	if mSwapped.cursor != 1 {
		t.Errorf("Expected cursor to move to 1 after moving line down in moveMode, got %d", mSwapped.cursor)
	}

	// 7. Reset moveMode via Esc key
	m.moveMode = true
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mEsc := updatedModel.(Model)
	if mEsc.moveMode {
		t.Errorf("Expected moveMode to be false after pressing Esc")
	}
	if mEsc.activeMenu != "lines" {
		t.Errorf("Expected activeMenu to stay 'lines' when exiting moveMode with Esc, got %q", mEsc.activeMenu)
	}
}

func TestTUI_UpdateItems_BoundaryAndMoveMode(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Lines[0] = []types.WidgetItem{
		{Type: "model"},
		{Type: "git-changes"},
		{Type: "sandbox"},
	}
	m := NewModel(settings, "/tmp/settings.toml")
	m.activeMenu = "items"
	m.selectedLine = 0

	// 1. Move up boundary in moveMode (cursor == 0)
	m.cursor = 0
	m.moveMode = true
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if updatedModel.(Model).cursor != 0 {
		t.Errorf("Expected cursor to remain 0 when moving item up at top")
	}

	// Move up swap in moveMode (valid swap when cursor == 1)
	m.cursor = 1
	m.moveMode = true
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	mItemSwapped := updatedModel.(Model)
	if mItemSwapped.cursor != 0 {
		t.Errorf("Expected cursor to follow item up to 0, got %d", mItemSwapped.cursor)
	}
	if mItemSwapped.settings.Lines[0][0].Type != "git-changes" {
		t.Errorf("Expected git-changes to move to index 0, got %q", mItemSwapped.settings.Lines[0][0].Type)
	}

	// Reset moveMode in items menu using Esc key
	m.moveMode = true
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if updatedModel.(Model).moveMode {
		t.Errorf("Expected Esc key to reset moveMode to false in items menu")
	}

	// 2. Move up boundary when moveMode == false (cursor == 0)
	m.moveMode = false
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if updatedModel.(Model).cursor != 0 {
		t.Errorf("Expected cursor to remain 0 when moveMode=false at top")
	}

	// 3. Move down boundary in moveMode (cursor == len(widgets)-1)
	m.cursor = 2
	m.moveMode = true
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if updatedModel.(Model).cursor != 2 {
		t.Errorf("Expected cursor to remain 2 when moving item down at bottom")
	}

	// 4. Move down boundary when moveMode == false (cursor == len(widgets)-1)
	m.moveMode = false
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if updatedModel.(Model).cursor != 2 {
		t.Errorf("Expected cursor to remain 2 when moveMode=false at bottom")
	}

	// 5. Delete last widget ('d') and test cursor clamping (m.cursor >= len)
	m.cursor = 2 // pointing to sandbox (last item)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	mDeleted := updatedModel.(Model)
	if len(mDeleted.settings.Lines[0]) != 2 {
		t.Fatalf("Expected 2 widgets remaining, got %d", len(mDeleted.settings.Lines[0]))
	}
	if mDeleted.cursor != 1 {
		t.Errorf("Expected cursor to clamp to last index (1), got %d", mDeleted.cursor)
	}

	// 6. Delete only remaining widget and test cursor clamping to 0 (len == 0)
	singleWidgetSettings := types.DefaultSettings()
	singleWidgetSettings.Lines[0] = []types.WidgetItem{{Type: "model"}}
	mSingle := NewModel(singleWidgetSettings, "/tmp/settings.toml")
	mSingle.activeMenu = "items"
	mSingle.selectedLine = 0
	mSingle.cursor = 0

	updatedModel, _ = mSingle.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	mEmpty := updatedModel.(Model)
	if len(mEmpty.settings.Lines[0]) != 0 {
		t.Fatalf("Expected 0 widgets remaining, got %d", len(mEmpty.settings.Lines[0]))
	}
	if mEmpty.cursor != 0 {
		t.Errorf("Expected cursor to clamp to 0 when line is empty, got %d", mEmpty.cursor)
	}

	// 7. Toggle moveMode off via 'm' key
	m.moveMode = true
	m.cursor = 0
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	if updatedModel.(Model).moveMode {
		t.Errorf("Expected 'm' key to toggle moveMode off when moveMode=true")
	}

	// 8. Toggle moveMode off via Enter key
	m.moveMode = true
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	if updatedModel.(Model).moveMode {
		t.Errorf("Expected Enter key to toggle moveMode off when moveMode=true")
	}
}

func TestTUI_UpdateAddWidget_Navigation(t *testing.T) {
	tests := []struct {
		name        string
		startCursor int
		key         tea.KeyMsg
		wantCursor  int
	}{
		{
			name:        "up when cursor > 0 decreases cursor",
			startCursor: 2,
			key:         tea.KeyMsg{Type: tea.KeyUp},
			wantCursor:  1,
		},
		{
			name:        "up boundary when cursor == 0 keeps cursor at 0",
			startCursor: 0,
			key:         tea.KeyMsg{Type: tea.KeyUp},
			wantCursor:  0,
		},
		{
			name:        "down when cursor < len-1 increases cursor",
			startCursor: 1,
			key:         tea.KeyMsg{Type: tea.KeyDown},
			wantCursor:  2,
		},
		{
			name:        "down boundary when cursor == len-1 keeps cursor at bottom",
			startCursor: len(widgetTypes) - 1,
			key:         tea.KeyMsg{Type: tea.KeyDown},
			wantCursor:  len(widgetTypes) - 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(types.DefaultSettings(), "/tmp/settings.toml")
			m.activeMenu = "add_widget"
			m.selectedLine = 0
			m.cursor = tt.startCursor

			updatedModel, _ := m.Update(tt.key)
			got := updatedModel.(Model).cursor
			if got != tt.wantCursor {
				t.Errorf("Expected cursor %d, got %d", tt.wantCursor, got)
			}
		})
	}
}

func TestTUI_UpdateAddWidget_EscKey(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")
	m.activeMenu = "add_widget"
	m.selectedLine = 0
	m.itemIndex = 3

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mEsc := updatedModel.(Model)
	if mEsc.activeMenu != "items" {
		t.Errorf("Expected activeMenu to return to 'items' on Esc, got %q", mEsc.activeMenu)
	}
	if mEsc.cursor != 3 {
		t.Errorf("Expected cursor to reset to itemIndex 3, got %d", mEsc.cursor)
	}
}

func TestTUI_UpdateAddWidget_EmptyLineInsertion(t *testing.T) {
	emptyLineSettings := types.DefaultSettings()
	emptyLineSettings.Lines[0] = []types.WidgetItem{}
	mEmpty := NewModel(emptyLineSettings, "/tmp/settings.toml")
	mEmpty.activeMenu = "add_widget"
	mEmpty.selectedLine = 0
	mEmpty.cursor = 0 // add first widget type

	updatedModel, _ := mEmpty.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mAdded := updatedModel.(Model)
	if len(mAdded.settings.Lines[0]) != 1 {
		t.Fatalf("Expected 1 widget in empty line after insert, got %d", len(mAdded.settings.Lines[0]))
	}
	if mAdded.cursor != 0 {
		t.Errorf("Expected cursor to be 0 for first widget insert in empty line, got %d", mAdded.cursor)
	}
}

func TestTUI_SelectSubmenus_BoundariesAndNoneCaps(t *testing.T) {
	settings := types.DefaultSettings()

	t.Run("ThemeBoundaries", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_theme"
		m.cursor = 0
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
		if updatedModel.(Model).cursor != 0 {
			t.Errorf("Expected select_theme cursor to stay 0 on Up")
		}
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if updatedModel.(Model).cursor != 1 {
			t.Errorf("Expected select_theme cursor to move to 1 on Down, got %d", updatedModel.(Model).cursor)
		}
		m.cursor = len(themesList) - 1
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if updatedModel.(Model).cursor != len(themesList)-1 {
			t.Errorf("Expected select_theme cursor to stay %d on Down at bottom", len(themesList)-1)
		}
	})

	t.Run("SeparatorBoundaries", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_separator"
		m.cursor = 0
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
		if updatedModel.(Model).cursor != 0 {
			t.Errorf("Expected select_separator cursor to stay 0 on Up")
		}
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if updatedModel.(Model).cursor != 1 {
			t.Errorf("Expected select_separator cursor to move to 1 on Down, got %d", updatedModel.(Model).cursor)
		}
		m.cursor = len(separatorsList) - 1
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if updatedModel.(Model).cursor != len(separatorsList)-1 {
			t.Errorf("Expected select_separator cursor to stay %d on Down at bottom", len(separatorsList)-1)
		}
	})

	t.Run("StartCapBoundariesAndNone", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_start_cap"
		m.cursor = 0
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
		if updatedModel.(Model).cursor != 0 {
			t.Errorf("Expected select_start_cap cursor to stay 0 on Up")
		}
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if updatedModel.(Model).cursor != 1 {
			t.Errorf("Expected select_start_cap cursor to move to 1 on Down, got %d", updatedModel.(Model).cursor)
		}
		m.cursor = len(startCapsList) - 1
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if updatedModel.(Model).cursor != len(startCapsList)-1 {
			t.Errorf("Expected select_start_cap cursor to stay %d on Down at bottom", len(startCapsList)-1)
		}

		// Test "None" cap selection (cursor = 0, value == "")
		m.cursor = 0 // "None" cap
		viewNoneStart := m.View()
		if viewNoneStart == "" {
			t.Errorf("Expected non-empty View string for None start cap preview")
		}

		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mNoneStart := updatedModel.(Model)
		if mNoneStart.settings.Powerline.StartCaps != "" {
			t.Errorf("Expected empty StartCaps string when selecting None, got %q", mNoneStart.settings.Powerline.StartCaps)
		}
	})

	t.Run("EndCapBoundariesAndNone", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_end_cap"
		m.cursor = 0
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
		if updatedModel.(Model).cursor != 0 {
			t.Errorf("Expected select_end_cap cursor to stay 0 on Up")
		}
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if updatedModel.(Model).cursor != 1 {
			t.Errorf("Expected select_end_cap cursor to move to 1 on Down, got %d", updatedModel.(Model).cursor)
		}
		m.cursor = len(endCapsList) - 1
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if updatedModel.(Model).cursor != len(endCapsList)-1 {
			t.Errorf("Expected select_end_cap cursor to stay %d on Down at bottom", len(endCapsList)-1)
		}

		// Test "None" cap selection (cursor = 0, value == "")
		m.cursor = 0 // "None" cap
		viewNoneEnd := m.View()
		if viewNoneEnd == "" {
			t.Errorf("Expected non-empty View string for None end cap preview")
		}

		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
		mNoneEnd := updatedModel.(Model)
		if mNoneEnd.settings.Powerline.EndCaps != "" {
			t.Errorf("Expected empty EndCaps string when selecting None, got %q", mNoneEnd.settings.Powerline.EndCaps)
		}
	})

	t.Run("ColorLevelBoundaries", func(t *testing.T) {
		m := NewModel(settings, "/tmp/settings.toml")
		m.activeMenu = "select_color_level"
		m.cursor = 0
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
		if updatedModel.(Model).cursor != 0 {
			t.Errorf("Expected select_color_level cursor to stay 0 on Up")
		}
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if updatedModel.(Model).cursor != 1 {
			t.Errorf("Expected select_color_level cursor to move to 1 on Down, got %d", updatedModel.(Model).cursor)
		}
		m.cursor = len(colorLevelsList) - 1
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if updatedModel.(Model).cursor != len(colorLevelsList)-1 {
			t.Errorf("Expected select_color_level cursor to stay %d on Down at bottom", len(colorLevelsList)-1)
		}
	})
}

func TestTUI_CustomCaps_Initialization(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Powerline.StartCaps = "[CUSTOM_START]"
	settings.Powerline.EndCaps = "[CUSTOM_END]"

	m := NewModel(settings, "/tmp/settings.toml")

	if m.startCapIndex == -1 || startCapsList[m.startCapIndex].value != "[CUSTOM_START]" {
		t.Errorf("Expected custom start cap to be initialized and appended to startCapsList")
	}
	if m.endCapIndex == -1 || endCapsList[m.endCapIndex].value != "[CUSTOM_END]" {
		t.Errorf("Expected custom end cap to be initialized and appended to endCapsList")
	}
}

func TestSaveSettings_ErrorPaths(t *testing.T) {
	settings := types.DefaultSettings()

	t.Run("Target path is directory error handling", func(t *testing.T) {
		tempDir := t.TempDir()
		invalidPath := filepath.Join(tempDir, "a_dir")
		if err := os.Mkdir(invalidPath, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		err := saveSettings(invalidPath, settings)
		if err == nil {
			t.Errorf("Expected saveSettings to fail when target path is a directory")
		}
	})

	t.Run("MkdirAll error simulation", func(t *testing.T) {
		tempDir := t.TempDir()
		parentFile := filepath.Join(tempDir, "file_blocking_mkdir")
		if err := os.WriteFile(parentFile, []byte("data"), 0644); err != nil {
			t.Fatalf("Failed to create file blocking mkdir: %v", err)
		}
		targetPath := filepath.Join(parentFile, "config.toml")

		err := saveSettings(targetPath, settings)
		if err == nil {
			t.Errorf("Expected saveSettings to fail when parent path is a regular file")
		}
	})

	t.Run("OpenFile error simulation and cleanup execution", func(t *testing.T) {
		tempDir := t.TempDir()
		readOnlyDir := filepath.Join(tempDir, "read_only_dir")
		if err := os.Mkdir(readOnlyDir, 0555); err != nil {
			t.Fatalf("Failed to create read only dir: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chmod(readOnlyDir, 0755)
		})

		targetPath := filepath.Join(readOnlyDir, "config.toml")
		err := saveSettings(targetPath, settings)
		if err == nil {
			t.Errorf("Expected saveSettings to fail when target directory is read-only")
		}
	})
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

func TestTUI_CorruptedConfigPrompt(t *testing.T) {
	settings := types.DefaultSettings()
	configPath := "/tmp/corrupted_settings.toml"
	loadErr := "settings.toml is not valid TOML"

	t.Run("initial state and view rendering", func(t *testing.T) {
		m := NewModel(settings, configPath, loadErr)
		if m.activeMenu != "corrupted_config" {
			t.Fatalf("Expected initial activeMenu 'corrupted_config', got %q", m.activeMenu)
		}
		if m.cursor != 0 {
			t.Errorf("Expected initial cursor 0, got %d", m.cursor)
		}

		viewStr := renderer.StripAnsi(m.View())
		for _, want := range []string{
			"Configuration File Corrupted",
			configPath,
			loadErr,
			"Load Default Settings",
			"Exit",
		} {
			if !strings.Contains(viewStr, want) {
				t.Errorf("Expected corrupted config view to contain %q, got:\n%s", want, viewStr)
			}
		}
	})

	tests := []struct {
		name           string
		keys           []tea.KeyMsg
		wantActiveMenu string
		wantCursor     int
		wantQuitting   bool
		wantSaved      bool
	}{
		{
			name: "navigate down and up within bounds",
			keys: []tea.KeyMsg{
				{Type: tea.KeyDown},
				{Type: tea.KeyDown}, // boundary at 1
			},
			wantActiveMenu: "corrupted_config",
			wantCursor:     1,
			wantQuitting:   false,
		},
		{
			name: "navigate down then up with j and k",
			keys: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("j")},
				{Type: tea.KeyRunes, Runes: []rune("k")},
				{Type: tea.KeyUp}, // boundary at 0
			},
			wantActiveMenu: "corrupted_config",
			wantCursor:     0,
			wantQuitting:   false,
		},
		{
			name: "select Load Default Settings transitions to main menu",
			keys: []tea.KeyMsg{
				{Type: tea.KeyEnter},
			},
			wantActiveMenu: "main",
			wantCursor:     0,
			wantQuitting:   false,
		},
		{
			name: "select Exit quits without saving",
			keys: []tea.KeyMsg{
				{Type: tea.KeyDown},
				{Type: tea.KeyEnter},
			},
			wantActiveMenu: "corrupted_config",
			wantCursor:     1,
			wantQuitting:   true,
			wantSaved:      false,
		},
		{
			name: "press Esc quits without saving",
			keys: []tea.KeyMsg{
				{Type: tea.KeyEsc},
			},
			wantActiveMenu: "corrupted_config",
			wantCursor:     0,
			wantQuitting:   true,
			wantSaved:      false,
		},
		{
			name: "press q quits without saving",
			keys: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune("q")},
			},
			wantActiveMenu: "corrupted_config",
			wantCursor:     0,
			wantQuitting:   true,
			wantSaved:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var curr tea.Model = NewModel(settings, configPath, loadErr)
			for _, k := range tt.keys {
				curr, _ = curr.Update(k)
			}
			got := curr.(Model)
			if got.activeMenu != tt.wantActiveMenu {
				t.Errorf("Expected activeMenu %q, got %q", tt.wantActiveMenu, got.activeMenu)
			}
			if got.cursor != tt.wantCursor {
				t.Errorf("Expected cursor %d, got %d", tt.wantCursor, got.cursor)
			}
			if got.quitting != tt.wantQuitting {
				t.Errorf("Expected quitting %v, got %v", tt.wantQuitting, got.quitting)
			}
			if got.saved != tt.wantSaved {
				t.Errorf("Expected saved %v, got %v", tt.wantSaved, got.saved)
			}
		})
	}
}
