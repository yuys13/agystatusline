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
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")
	m.activeMenu = "lines"
	m.cursor = 0

	// 1. Test Add Line ("a")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	updatedModel, _ := m.Update(msg)
	newModel := updatedModel.(Model)

	initialLinesCount := len(settings.Lines)
	if len(newModel.settings.Lines) != initialLinesCount+1 {
		t.Errorf("Expected lines count to be %d, got %d", initialLinesCount+1, len(newModel.settings.Lines))
	}
	if newModel.cursor != initialLinesCount {
		t.Errorf("Expected cursor to be at the new line index %d, got %d", initialLinesCount, newModel.cursor)
	}

	// 2. Test Delete Line ("d")
	m = newModel
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
	updatedModel, _ = m.Update(msg)
	newModel = updatedModel.(Model)

	if len(newModel.settings.Lines) != initialLinesCount {
		t.Errorf("Expected lines count to be %d after deletion, got %d", initialLinesCount, len(newModel.settings.Lines))
	}
	if newModel.cursor != initialLinesCount-1 {
		t.Errorf("Expected cursor to adjust to %d, got %d", initialLinesCount-1, newModel.cursor)
	}

	// 3. Test Cannot Delete Last Line
	// Delete until 1 line is left
	for len(newModel.settings.Lines) > 1 {
		m = newModel
		m.cursor = len(m.settings.Lines) - 1 // Fix: delete the last line so the first line with widgets remains
		updatedModel, _ = m.Update(msg)
		newModel = updatedModel.(Model)
	}
	if len(newModel.settings.Lines) != 1 {
		t.Fatalf("Setup failed: expected 1 line, got %d", len(newModel.settings.Lines))
	}

	// Try to delete the last line
	m = newModel
	m.cursor = 0
	updatedModel, _ = m.Update(msg)
	newModel = updatedModel.(Model)
	if len(newModel.settings.Lines) != 1 {
		t.Errorf("Expected lines count to remain 1, got %d", len(newModel.settings.Lines))
	}

	// 4. Test Move Line
	// First add lines to have at least 2 lines
	m = newModel
	m.activeMenu = "lines"
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}) // Add to 2 lines
	newModel = updatedModel.(Model)

	// Switch to move mode ("m")
	m = newModel
	m.cursor = 1
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	newModel = updatedModel.(Model)
	if !newModel.moveMode {
		t.Errorf("Expected moveMode to be true")
	}

	// Move up (swap line 1 and line 0)
	m = newModel
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = m.Update(upMsg)
	newModel = updatedModel.(Model)

	// Since we swapped line 1 (which was empty) and line 0 (which has widgets),
	// line 0 should now be empty and line 1 should have widgets.
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
	m = newModel
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	newModel = updatedModel.(Model)
	if newModel.moveMode {
		t.Errorf("Expected moveMode to be false after pressing Enter")
	}

	// 5. Test Enter to navigate to widgets editor
	m = newModel
	m.cursor = 1
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	newModel = updatedModel.(Model)
	if newModel.activeMenu != "items" {
		t.Errorf("Expected activeMenu to transition to 'items', got %q", newModel.activeMenu)
	}
	if newModel.selectedLine != 1 {
		t.Errorf("Expected selectedLine to be 1, got %d", newModel.selectedLine)
	}

	// 6. Test Esc to go back to main menu
	m = newModel
	m.activeMenu = "lines"
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	newModel = updatedModel.(Model)
	if newModel.activeMenu != "main" {
		t.Errorf("Expected activeMenu to go back to 'main', got %q", newModel.activeMenu)
	}
}

func TestTUI_ItemsOperations(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Lines[0] = []types.WidgetItem{
		{Type: "model", Color: "brightMagenta"},
		{Type: "sandbox", Color: "brightBlack"},
		{Type: "git-branch", Color: "brightMagenta"},
		{Type: "git-changes", Color: "yellow"},
	}
	m := NewModel(settings, "/tmp/settings.toml")
	m.activeMenu = "items"
	m.selectedLine = 0
	m.cursor = 1 // Pointing to sandbox

	// 1. Delete Widget ("d")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
	updatedModel, _ := m.Update(msg)
	newModel := updatedModel.(Model)

	if len(newModel.settings.Lines[0]) != 3 {
		t.Errorf("Expected widget count to be 3 after deletion, got %d", len(newModel.settings.Lines[0]))
	}
	if newModel.settings.Lines[0][1].Type != "git-branch" {
		t.Errorf("Expected widget at index 1 to be 'git-branch', got %q", newModel.settings.Lines[0][1].Type)
	}

	// 2. Move Widget
	m = newModel
	m.cursor = 0 // Pointing to model
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	newModel = updatedModel.(Model)
	if !newModel.moveMode {
		t.Errorf("Expected moveMode to be true")
	}

	// Move down (swap index 0 and 1)
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = newModel.Update(downMsg)
	newModel = updatedModel.(Model)

	if newModel.settings.Lines[0][0].Type != "git-branch" {
		t.Errorf("Expected widget at index 0 to be 'git-branch' after swapping, got %q", newModel.settings.Lines[0][0].Type)
	}
	if newModel.settings.Lines[0][1].Type != "model" {
		t.Errorf("Expected widget at index 1 to be 'model' after swapping, got %q", newModel.settings.Lines[0][1].Type)
	}
	if newModel.cursor != 1 {
		t.Errorf("Expected cursor to follow the item to index 1, got %d", newModel.cursor)
	}

	// 3. Esc to go back to lines menu
	m = newModel
	m.activeMenu = "items"
	m.moveMode = false
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	newModel = updatedModel.(Model)
	if newModel.activeMenu != "lines" {
		t.Errorf("Expected activeMenu to go back to 'lines', got %q", newModel.activeMenu)
	}
	if newModel.cursor != 0 {
		t.Errorf("Expected cursor in lines menu to be the selected line (0), got %d", newModel.cursor)
	}
}

func TestTUI_AddWidget(t *testing.T) {
	settings := types.DefaultSettings()
	initialLen := len(settings.Lines[0])
	m := NewModel(settings, "/tmp/settings.toml")
	m.activeMenu = "items"
	m.selectedLine = 0
	m.cursor = 1

	// 1. Press "a" to trigger Add Widget screen
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

	// 2. Select a widget type and add it
	m = newModel
	for range 2 {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updatedModel.(Model)
	}
	if m.cursor != 2 {
		t.Fatalf("Expected cursor to be 2, got %d", m.cursor)
	}

	// Press Enter to add
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	newModel = updatedModel.(Model)

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
}

func TestTUI_AddQuotaBarWidgets(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")
	m.activeMenu = "items"
	m.selectedLine = 0
	m.cursor = 0

	// 1. "a" キーで追加画面に遷移
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	updatedModel, _ := m.Update(msg)
	newModel := updatedModel.(Model)

	if newModel.activeMenu != "add_widget" {
		t.Fatalf("Expected activeMenu to be %q, got %q", "add_widget", newModel.activeMenu)
	}

	// 2. ウィジェット追加リストにクォータバーウィジェットが含まれているか確認
	var foundG5hB, foundGwkB, found3p5hB, found3pwkB bool
	for _, wt := range widgetTypes {
		switch wt.item.Type {
		case "quota-bar-5h":
			foundG5hB = true
		case "quota-bar-7d":
			foundGwkB = true
		case "quota-bar-3p-5h":
			found3p5hB = true
		case "quota-bar-3p-7d":
			found3pwkB = true
		}
	}
	if !foundG5hB || !foundGwkB || !found3p5hB || !found3pwkB {
		t.Errorf("Expected all 4 quota-bar presets in widgetTypes, got Gemini 5h Bar:%t, Gemini 7d Bar:%t, 3P 5h Bar:%t, 3P 7d Bar:%t",
			foundG5hB, foundGwkB, found3p5hB, found3pwkB)
	}

	// 3. 実際に Gemini 5h クォータ Bar ウィジェットを追加してみる。
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

	m = newModel
	m.cursor = targetBarIdx
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	finalBarModel := updatedModel.(Model)

	if finalBarModel.activeMenu != "items" {
		t.Fatalf("Expected activeMenu to return to %q, got %q", "items", finalBarModel.activeMenu)
	}

	addedBarWidget := finalBarModel.settings.Lines[0][1]
	if addedBarWidget.Type != "quota-bar-5h" {
		t.Errorf("Expected widget type 'quota-bar-5h', got %q", addedBarWidget.Type)
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
	if !strings.Contains(viewStr, "gemini-5h 50.19% (2h 28m)") {
		t.Errorf("Expected live preview to contain 'gemini-5h 50.19%% (2h 28m)', but it did not. View:\n%s", viewStr)
	}
	if !strings.Contains(viewStr, "3p-5h 100.00% (4h 59m)") {
		t.Errorf("Expected live preview to contain '3p-5h 100.00%% (4h 59m)', but it did not. View:\n%s", viewStr)
	}
}

func TestTUI_PowerlineSeparator(t *testing.T) {
	tests := []struct {
		name               string
		setupSettings      func() types.Settings
		keyMsg             *tea.KeyMsg
		startMenu          string
		startCursor        int
		wantSeparatorIndex int
		wantCustomSepName  string
		wantActiveMenu     string
		wantCursor         func(m Model) int
	}{
		{
			name: "Default Separator",
			setupSettings: func() types.Settings {
				return types.DefaultSettings()
			},
			wantSeparatorIndex: 1,
		},
		{
			name: "Custom Separator (exists in list)",
			setupSettings: func() types.Settings {
				s := types.DefaultSettings()
				s.Powerline.Separator = "\uE0B4"
				return s
			},
			wantSeparatorIndex: 2,
		},
		{
			name: "None Separator",
			setupSettings: func() types.Settings {
				s := types.DefaultSettings()
				s.Powerline.Separator = ""
				return s
			},
			wantSeparatorIndex: 0,
		},
		{
			name: "Custom Separator (NOT in list)",
			setupSettings: func() types.Settings {
				s := types.DefaultSettings()
				s.Powerline.Separator = "♦"
				return s
			},
			wantCustomSepName: "Custom (♦)",
		},
		{
			name: "Navigation to Separator Selection Menu",
			setupSettings: func() types.Settings {
				s := types.DefaultSettings()
				s.Powerline.Separator = "♦"
				return s
			},
			startMenu:      "powerline",
			startCursor:    2,
			keyMsg:         &tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")},
			wantActiveMenu: "select_separator",
			wantCursor: func(m Model) int {
				return m.separatorIndex
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.setupSettings()
			m := NewModel(settings, "/tmp/settings.toml")

			if tt.wantCustomSepName != "" {
				if m.separatorIndex == -1 {
					t.Fatalf("Expected new custom separator to be added to list, but got index -1")
				}
				if separatorsList[m.separatorIndex].name != tt.wantCustomSepName {
					t.Errorf("Expected custom separator name to be %q, got %q", tt.wantCustomSepName, separatorsList[m.separatorIndex].name)
				}
			} else if tt.wantSeparatorIndex != 0 || tt.keyMsg == nil {
				if m.separatorIndex != tt.wantSeparatorIndex {
					t.Errorf("Expected separatorIndex to be %d, got %d", tt.wantSeparatorIndex, m.separatorIndex)
				}
			}

			if tt.keyMsg != nil {
				m.activeMenu = tt.startMenu
				m.cursor = tt.startCursor

				updatedModel, _ := m.Update(*tt.keyMsg)
				newModel := updatedModel.(Model)

				if tt.wantActiveMenu != "" && newModel.activeMenu != tt.wantActiveMenu {
					t.Errorf("Expected activeMenu to transition to %q, got %q", tt.wantActiveMenu, newModel.activeMenu)
				}
				if tt.wantCursor != nil {
					expectedCursor := tt.wantCursor(m)
					if newModel.cursor != expectedCursor {
						t.Errorf("Expected cursor to be %d, got %d", expectedCursor, newModel.cursor)
					}
				}
			}
		})
	}
}

func TestTUI_PowerlineSubmenu(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")

	// 1. Enter on main menu cursor 1 -> navigate to "powerline" submenu
	m.cursor = 1
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mPowerline := updatedModel.(Model)

	if mPowerline.activeMenu != "powerline" {
		t.Fatalf("Expected activeMenu to be 'powerline', got %q", mPowerline.activeMenu)
	}
	if mPowerline.cursor != 0 {
		t.Errorf("Expected cursor in powerline menu to start at 0, got %d", mPowerline.cursor)
	}

	// 2. Press Esc on powerline submenu -> return to main menu with cursor at 1
	updatedModel, _ = mPowerline.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mMain := updatedModel.(Model)
	if mMain.activeMenu != "main" {
		t.Errorf("Expected activeMenu to return to 'main' on Esc, got %q", mMain.activeMenu)
	}
	if mMain.cursor != 1 {
		t.Errorf("Expected main menu cursor to be 1, got %d", mMain.cursor)
	}

	// 3. Toggle Powerline Mode on powerline cursor 0
	initialEnabled := mPowerline.settings.Powerline.Enabled
	updatedModel, _ = mPowerline.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mToggle := updatedModel.(Model)
	if mToggle.settings.Powerline.Enabled != !initialEnabled {
		t.Errorf("Expected Powerline.Enabled to toggle from %v to %v", initialEnabled, !initialEnabled)
	}

	// 4. Select "Back to Main Menu" on powerline cursor 5
	mPowerline.cursor = 5
	updatedModel, _ = mPowerline.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mBack := updatedModel.(Model)
	if mBack.activeMenu != "main" {
		t.Errorf("Expected activeMenu to return to 'main' on 'Back to Main Menu', got %q", mBack.activeMenu)
	}
	if mBack.cursor != 1 {
		t.Errorf("Expected main menu cursor to be 1, got %d", mBack.cursor)
	}

	// 5. Test viewPowerline rendering and navigation boundaries
	mPowerlineNav := NewModel(settings, "/tmp/settings.toml")
	mPowerlineNav.activeMenu = "powerline"

	// View rendering check
	viewStr := mPowerlineNav.View()
	if !strings.Contains(viewStr, "Powerline Settings") || !strings.Contains(viewStr, "Toggle Powerline Mode") {
		t.Errorf("Expected viewPowerline output to contain Powerline Settings menu title and items, got:\n%s", viewStr)
	}

	// Boundary Up at 0
	mPowerlineNav.cursor = 0
	updatedModel, _ = mPowerlineNav.Update(tea.KeyMsg{Type: tea.KeyUp})
	mBoundary := updatedModel.(Model)
	if mBoundary.cursor != 0 {
		t.Errorf("Expected cursor to remain 0 when pressing Up at upper bound in powerline menu, got %d", mBoundary.cursor)
	}

	updatedModel, _ = mPowerlineNav.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	mBoundary = updatedModel.(Model)
	if mBoundary.cursor != 0 {
		t.Errorf("Expected cursor to remain 0 when pressing 'k' at upper bound in powerline menu, got %d", mBoundary.cursor)
	}

	// Down navigation and Boundary Down at 5
	updatedModel, _ = mPowerlineNav.Update(tea.KeyMsg{Type: tea.KeyDown})
	mNav := updatedModel.(Model)
	if mNav.cursor != 1 {
		t.Errorf("Expected cursor to move to 1 on Down key, got %d", mNav.cursor)
	}

	updatedModel, _ = mNav.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	mNav = updatedModel.(Model)
	if mNav.cursor != 2 {
		t.Errorf("Expected cursor to move to 2 on 'j' key, got %d", mNav.cursor)
	}

	mNav.cursor = 5
	updatedModel, _ = mNav.Update(tea.KeyMsg{Type: tea.KeyDown})
	mBoundary = updatedModel.(Model)
	if mBoundary.cursor != 5 {
		t.Errorf("Expected cursor to remain 5 when pressing Down at lower bound in powerline menu, got %d", mBoundary.cursor)
	}

	updatedModel, _ = mNav.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	mBoundary = updatedModel.(Model)
	if mBoundary.cursor != 5 {
		t.Errorf("Expected cursor to remain 5 when pressing 'j' at lower bound in powerline menu, got %d", mBoundary.cursor)
	}
}

func TestTUI_SelectThemeMenu(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	settings.Powerline.Theme = "nord"
	m := NewModel(settings, "/tmp/settings.toml")
	m.activeMenu = "powerline"
	m.cursor = 1 // Select Powerline Theme

	// Enter opens theme list selection
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

	// Verify that preview changes dynamically when cursor moves in select_theme menu
	mTheme0 := NewModel(settings, "/tmp/settings.toml")
	mTheme0.activeMenu = "select_theme"
	mTheme0.cursor = 0
	viewNord := mTheme0.View()

	viewNordAurora := mTheme.View()

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

	// Press Enter to confirm selection -> return to powerline submenu with cursor 1
	updatedModel, _ = mTheme.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
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

	// Test Esc key to cancel theme selection -> return to powerline submenu with cursor 1
	mCancel := mTheme // cursor at 1 (nord-aurora)
	updatedModel, _ = mCancel.Update(tea.KeyMsg{Type: tea.KeyEsc})
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
}

func TestTUI_SelectSeparatorMenu(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	settings.Powerline.Separator = "\uE0B0"
	m := NewModel(settings, "/tmp/settings.toml")
	m.activeMenu = "powerline"
	m.cursor = 2 // Select Powerline Separator

	// Enter opens separator list selection
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

	// Verify viewSelectSeparator rendering and Live Preview
	viewSepSelected := mSep.View()
	if !strings.Contains(viewSepSelected, "Select Powerline Separator") {
		t.Errorf("Expected viewSelectSeparator to render title 'Select Powerline Separator', got:\n%s", viewSepSelected)
	}

	// Press Enter to confirm selection
	updatedModel, _ = mSep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mConfirmed := updatedModel.(Model)

	if mConfirmed.activeMenu != "powerline" {
		t.Errorf("Expected activeMenu to return to 'powerline', got %q", mConfirmed.activeMenu)
	}
	if mConfirmed.cursor != 2 {
		t.Errorf("Expected powerline menu cursor to be 2, got %d", mConfirmed.cursor)
	}

	// Test Esc key to cancel separator selection
	mCancel := mSep // cursor at 2 (Round)
	updatedModel, _ = mCancel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mCancelled := updatedModel.(Model)

	if mCancelled.activeMenu != "powerline" {
		t.Errorf("Expected activeMenu to return to 'powerline' on Esc, got %q", mCancelled.activeMenu)
	}
	if mCancelled.cursor != 2 {
		t.Errorf("Expected powerline menu cursor to return to 2, got %d", mCancelled.cursor)
	}
}

func TestTUI_SelectStartCapMenu(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	settings.Powerline.StartCaps = "\uE0B2"
	m := NewModel(settings, "/tmp/settings.toml")
	m.activeMenu = "powerline"
	m.cursor = 3 // Select Powerline Start Cap

	// Enter opens start cap selection
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mCap := updatedModel.(Model)

	if mCap.activeMenu != "select_start_cap" {
		t.Fatalf("Expected activeMenu to transition to 'select_start_cap', got %q", mCap.activeMenu)
	}

	// Press Enter to confirm selection
	updatedModel, _ = mCap.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mConfirmed := updatedModel.(Model)

	if mConfirmed.activeMenu != "powerline" {
		t.Errorf("Expected activeMenu to return to 'powerline', got %q", mConfirmed.activeMenu)
	}
	if mConfirmed.cursor != 3 {
		t.Errorf("Expected powerline menu cursor to be 3, got %d", mConfirmed.cursor)
	}

	// Test Esc key to cancel selection
	mCancel := mCap
	updatedModel, _ = mCancel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mCancelled := updatedModel.(Model)

	if mCancelled.activeMenu != "powerline" {
		t.Errorf("Expected activeMenu to return to 'powerline' on Esc, got %q", mCancelled.activeMenu)
	}
	if mCancelled.cursor != 3 {
		t.Errorf("Expected powerline menu cursor to return to 3, got %d", mCancelled.cursor)
	}
}

func TestTUI_SelectEndCapMenu(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	settings.Powerline.EndCaps = "\uE0B0"
	m := NewModel(settings, "/tmp/settings.toml")
	m.activeMenu = "powerline"
	m.cursor = 4 // Select Powerline End Cap

	// Enter opens end cap selection
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mCap := updatedModel.(Model)

	if mCap.activeMenu != "select_end_cap" {
		t.Fatalf("Expected activeMenu to transition to 'select_end_cap', got %q", mCap.activeMenu)
	}

	// Press Enter to confirm selection
	updatedModel, _ = mCap.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mConfirmed := updatedModel.(Model)

	if mConfirmed.activeMenu != "powerline" {
		t.Errorf("Expected activeMenu to return to 'powerline', got %q", mConfirmed.activeMenu)
	}
	if mConfirmed.cursor != 4 {
		t.Errorf("Expected powerline menu cursor to be 4, got %d", mConfirmed.cursor)
	}

	// Test Esc key to cancel selection
	mCancel := mCap
	updatedModel, _ = mCancel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mCancelled := updatedModel.(Model)

	if mCancelled.activeMenu != "powerline" {
		t.Errorf("Expected activeMenu to return to 'powerline' on Esc, got %q", mCancelled.activeMenu)
	}
	if mCancelled.cursor != 4 {
		t.Errorf("Expected powerline menu cursor to return to 4, got %d", mCancelled.cursor)
	}
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

	if !strings.Contains(previewPart, "5h") || !strings.Contains(previewPart, "50.2%") {
		t.Errorf("Expected Live Preview to dynamically display the currently selected widget type 'Quota Bar: 5h' in the preview part, but it did not. Preview part:\n%s", previewPart)
	}
}

func TestTUI_SelectColorLevel(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	settings.General.ColorLevel = 2 // ANSI 256 colors
	m := NewModel(settings, "/tmp/settings.toml")

	// 1. Initial state
	if m.colorLevelIndex != 1 {
		t.Errorf("Expected initial colorLevelIndex to be 1 (ANSI 256 Colors), got %d", m.colorLevelIndex)
	}

	// 2. Select Color Level menu on main menu
	// cursor index for color level will be 2
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

	// 3. Move cursor and select Truecolor
	mSelect.cursor = 2 // Truecolor (24-bit)
	updatedModel, _ = mSelect.Update(msgEnter)
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

	// 4. Cancel selection with Esc
	mSelectCancel := mSelect
	mSelectCancel.cursor = 0 // ANSI 16 colors
	updatedModel, _ = mSelectCancel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mCancelled := updatedModel.(Model)

	if mCancelled.activeMenu != "main" {
		t.Errorf("Expected activeMenu to return to 'main' on Esc, got %q", mCancelled.activeMenu)
	}
	if mCancelled.settings.General.ColorLevel != 2 {
		t.Errorf("Expected settings.General.ColorLevel to remain 2, got %d", mCancelled.settings.General.ColorLevel)
	}

	// 5. Test Live Preview during color level selection
	mSelectPreview := mSelect
	mSelectPreview.cursor = 0 // ANSI 16 colors (should trigger 16-color rendering in preview)
	viewStr := mSelectPreview.View()
	if viewStr == "" {
		t.Errorf("Expected non-empty view string")
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
		t.Run(sep.name, func(t *testing.T) {
			if sep.value == "/" {
				t.Errorf("Slash ASCII (/) separator should be removed")
			}
			if sep.value == "|" {
				t.Errorf("Bar ASCII (|) separator should be removed")
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
		"git-branch",
		"git-changes",
		"quota-5h",
		"quota-7d",
		"quota-3p-5h",
		"quota-3p-7d",
		"quota-bar-5h",
		"quota-bar-7d",
		"quota-bar-3p-5h",
		"quota-bar-3p-7d",
		"custom-text",
	}

	for _, expected := range expectedTypes {
		t.Run("HasWidget_"+expected, func(t *testing.T) {
			var found bool
			for _, wt := range widgetTypes {
				if wt.item.Type == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected widget type %q to be present in widgetTypes", expected)
			}
		})
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

func TestSaveSettings(t *testing.T) {
	settings := types.DefaultSettings()

	tests := []struct {
		name           string
		setupPath      func(dir string) string
		wantErr        bool
		checkFileExist bool
	}{
		{
			name: "Successful save",
			setupPath: func(dir string) string {
				return filepath.Join(dir, "config.toml")
			},
			wantErr:        false,
			checkFileExist: true,
		},
		{
			name: "Target path is directory error handling",
			setupPath: func(dir string) string {
				invalidPath := filepath.Join(dir, "a_dir")
				_ = os.Mkdir(invalidPath, 0755)
				return invalidPath
			},
			wantErr:        true,
			checkFileExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			targetPath := tt.setupPath(tempDir)

			err := saveSettings(targetPath, settings)
			if (err != nil) != tt.wantErr {
				t.Fatalf("saveSettings() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkFileExist {
				if _, err := os.Stat(targetPath); os.IsNotExist(err) {
					t.Errorf("Expected settings file to exist at %q", targetPath)
				}
			}

			// Verify temporary files are cleaned up
			files, _ := os.ReadDir(tempDir)
			for _, f := range files {
				if strings.HasSuffix(f.Name(), ".tmp") {
					t.Errorf("Expected no temporary files remaining in %q, but found %q", tempDir, f.Name())
				}
			}
		})
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
	m := NewModel(settings, "/tmp/settings.toml")

	cmd := m.Init()
	if cmd != nil {
		t.Errorf("Expected Init() to return nil, got %v", cmd)
	}

	// Main menu item 3: Save & Exit
	m.cursor = 3
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

	// Main menu item 4: Discard & Exit
	m.cursor = 4
	m.saved = false
	updatedModel, quitCmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
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
}

func TestTUI_UpdateMain_BoundaryAndEdgeCases(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")

	// 1. Up key at cursor == 0 (upper boundary limit)
	m.cursor = 0
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	mResult := updatedModel.(Model)
	if mResult.cursor != 0 {
		t.Errorf("Expected cursor to remain 0 when pressing Up at upper bound, got %d", mResult.cursor)
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	mResult = updatedModel.(Model)
	if mResult.cursor != 0 {
		t.Errorf("Expected cursor to remain 0 when pressing 'k' at upper bound, got %d", mResult.cursor)
	}

	// 2. Down key at cursor == 4 (lower boundary limit, maxItems - 1)
	m.cursor = 4
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	mResult = updatedModel.(Model)
	if mResult.cursor != 4 {
		t.Errorf("Expected cursor to remain 4 when pressing Down at lower bound, got %d", mResult.cursor)
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	mResult = updatedModel.(Model)
	if mResult.cursor != 4 {
		t.Errorf("Expected cursor to remain 4 when pressing 'j' at lower bound, got %d", mResult.cursor)
	}

	// 3. Item 1 Enter: Navigate to Powerline Submenu
	m.cursor = 1
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mSubmenu := updatedModel.(Model)
	if mSubmenu.activeMenu != "powerline" {
		t.Errorf("Expected activeMenu to be 'powerline', got %q", mSubmenu.activeMenu)
	}

	// 4. Item 3 Enter: Save & Exit with save error
	m.cursor = 3
	// Set invalid configPath where parent directory is a regular file
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

	// 5. Non-KeyMsg message handling in Update
	_, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if cmd != nil {
		t.Errorf("Expected nil command for non-KeyMsg input, got %v", cmd)
	}

	// 6. Unknown activeMenu handling in Update
	mUnknown := m
	mUnknown.activeMenu = "unknown_menu_type"
	updatedModel, cmd = mUnknown.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	mResult = updatedModel.(Model)
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

func TestTUI_UpdateAddWidget_NavigationAndEdgeCases(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")
	m.activeMenu = "add_widget"
	m.selectedLine = 0

	// 1. Navigation up when cursor > 0
	m.cursor = 2
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if updatedModel.(Model).cursor != 1 {
		t.Errorf("Expected cursor to decrease to 1, got %d", updatedModel.(Model).cursor)
	}

	// 2. Navigation up boundary when cursor == 0
	m.cursor = 0
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if updatedModel.(Model).cursor != 0 {
		t.Errorf("Expected cursor to stay 0, got %d", updatedModel.(Model).cursor)
	}

	// 3. Navigation down boundary when cursor == len(widgetTypes)-1
	m.cursor = len(widgetTypes) - 1
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if updatedModel.(Model).cursor != len(widgetTypes)-1 {
		t.Errorf("Expected cursor to stay %d at bottom, got %d", len(widgetTypes)-1, updatedModel.(Model).cursor)
	}

	// 4. Esc key handling (cancels and returns to items menu)
	m.itemIndex = 3
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mEsc := updatedModel.(Model)
	if mEsc.activeMenu != "items" {
		t.Errorf("Expected activeMenu to return to 'items' on Esc, got %q", mEsc.activeMenu)
	}
	if mEsc.cursor != 3 {
		t.Errorf("Expected cursor to reset to itemIndex 3, got %d", mEsc.cursor)
	}

	// 5. Empty line insertion (len(widgets) == 0)
	emptyLineSettings := types.DefaultSettings()
	emptyLineSettings.Lines[0] = []types.WidgetItem{}
	mEmpty := NewModel(emptyLineSettings, "/tmp/settings.toml")
	mEmpty.activeMenu = "add_widget"
	mEmpty.selectedLine = 0
	mEmpty.cursor = 0 // add first widget type

	updatedModel, _ = mEmpty.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
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
	m := NewModel(settings, "/tmp/settings.toml")

	// 1. Theme selection boundary checks
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

	// 2. Separator selection boundary checks
	m.activeMenu = "select_separator"
	m.cursor = 0
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
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

	// 3. StartCap selection boundary & "None" cap
	m.activeMenu = "select_start_cap"
	m.cursor = 0
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
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

	// 4. EndCap selection boundary & "None" cap
	m.activeMenu = "select_end_cap"
	m.cursor = 0
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
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

	// 5. ColorLevel selection boundary checks
	m.activeMenu = "select_color_level"
	m.cursor = 0
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
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
