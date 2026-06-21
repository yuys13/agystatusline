package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yuys13/agystatusline/renderer"
	"github.com/yuys13/agystatusline/types"
	"github.com/yuys13/agystatusline/widgets"
)

func TestInitialModel(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.json")

	if m.settings.Version != 3 {
		t.Errorf("Expected settings version 3, got %d", m.settings.Version)
	}

	if m.activeMenu != "main" {
		t.Errorf("Expected initial menu 'main', got '%s'", m.activeMenu)
	}
}

func TestTUI_UpdateQuit(t *testing.T) {
	settings := types.DefaultSettings()

	// 1. Send key event "q" on main menu: should exit
	m := NewModel(settings, "/tmp/settings.json")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	updatedModel, _ := m.Update(msg)
	newModel := updatedModel.(Model)
	if !newModel.quitting {
		t.Errorf("Expected quitting to be true after pressing 'q' on main menu")
	}

	// 2. Send key event "q" on lines menu: should NOT exit
	m2 := NewModel(settings, "/tmp/settings.json")
	m2.activeMenu = "lines"
	updatedModel2, _ := m2.Update(msg)
	newModel2 := updatedModel2.(Model)
	if newModel2.quitting {
		t.Errorf("Expected quitting to be false after pressing 'q' on lines menu")
	}

	// 3. Send key event "ctrl+c" on lines menu: should exit
	ctrlCMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedModel3, _ := m2.Update(ctrlCMsg)
	newModel3 := updatedModel3.(Model)
	if !newModel3.quitting {
		t.Errorf("Expected quitting to be true after pressing 'ctrl+c' on lines menu")
	}
}

func TestTUI_LivePreviewModelName(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.json")

	viewStr := m.View()
	expectedModelName := "Gemini 3.5 Flash (Medium)"
	if !strings.Contains(viewStr, expectedModelName) {
		t.Errorf("Expected live preview to contain model name %q, but it did not. View output:\n%s", expectedModelName, viewStr)
	}
}

func TestTUI_LayoutAndBorders(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.json")

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
	m := NewModel(settings, "/tmp/settings.json")

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
	m := NewModel(settings, "/tmp/settings.json")
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
		{ID: "1", Type: "model", Color: "brightMagenta"},
		{ID: "2", Type: "separator"},
		{ID: "3", Type: "sandbox", Color: "brightBlack"},
		{ID: "4", Type: "separator"},
		{ID: "5", Type: "git-branch", Color: "brightMagenta"},
		{ID: "6", Type: "separator"},
		{ID: "7", Type: "git-changes", Color: "yellow"},
	}
	m := NewModel(settings, "/tmp/settings.json")
	m.activeMenu = "items"
	m.selectedLine = 0
	m.cursor = 2 // Pointing to sandbox

	// 1. Delete Widget ("d")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
	updatedModel, _ := m.Update(msg)
	newModel := updatedModel.(Model)

	if len(newModel.settings.Lines[0]) != 6 {
		t.Errorf("Expected widget count to be 6 after deletion, got %d", len(newModel.settings.Lines[0]))
	}
	if newModel.settings.Lines[0][2].Type != "separator" {
		t.Errorf("Expected widget at index 2 to be 'separator', got %q", newModel.settings.Lines[0][2].Type)
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
	t.Logf("Before swap: index 0 type = %q, index 1 type = %q", newModel.settings.Lines[0][0].Type, newModel.settings.Lines[0][1].Type)
	updatedModel, _ = newModel.Update(downMsg)
	newModel = updatedModel.(Model)
	t.Logf("After swap: index 0 type = %q, index 1 type = %q, cursor = %d", newModel.settings.Lines[0][0].Type, newModel.settings.Lines[0][1].Type, newModel.cursor)

	if newModel.settings.Lines[0][0].Type != "separator" {
		t.Errorf("Expected widget at index 0 to be 'separator' after swapping, got %q", newModel.settings.Lines[0][0].Type)
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
	m := NewModel(settings, "/tmp/settings.json")
	m.activeMenu = "items"
	m.selectedLine = 0
	m.cursor = 1 // Pointing to separator (index 1)

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
	for range 3 {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updatedModel.(Model)
	}
	if m.cursor != 3 {
		t.Fatalf("Expected cursor to be 3, got %d", m.cursor)
	}

	// Press Enter to add
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	newModel = updatedModel.(Model)

	if newModel.activeMenu != "items" {
		t.Errorf("Expected activeMenu to return to 'items', got %q", newModel.activeMenu)
	}
	if len(newModel.settings.Lines[0]) != 8 {
		t.Errorf("Expected 8 widgets in line 0, got %d", len(newModel.settings.Lines[0]))
	}
	if newModel.settings.Lines[0][2].Type != "custom-text" {
		t.Errorf("Expected added widget at index 2 to be 'custom-text', got %q", newModel.settings.Lines[0][2].Type)
	}
	if newModel.cursor != 2 {
		t.Errorf("Expected cursor to point to the newly added widget (index 2), got %d", newModel.cursor)
	}
}

func TestTUI_AddQuotaWidgets(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.json")
	m.activeMenu = "items"
	m.selectedLine = 0
	m.cursor = 0

	// 1. "a" キーで追加画面に遷移
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	updatedModel, _ := m.Update(msg)
	newModel := updatedModel.(Model)

	if newModel.activeMenu != "add_widget" {
		t.Fatalf("Expected activeMenu to be 'add_widget', got %s", newModel.activeMenu)
	}

	// 2. ウィジェット追加リストにクォータウィジェット（通常・%・リセット）が含まれているか確認
	var foundG5h, foundGwk, found3p5h, found3pwk bool
	var foundG5hP, foundGwkP, found3p5hP, found3pwkP bool
	var foundG5hR, foundGwkR, found3p5hR, found3pwkR bool
	var foundG5hB, foundGwkB, found3p5hB, found3pwkB bool
	for _, wt := range widgetTypes {
		if wt.wType == "quota" {
			key := wt.metadata["key"]
			display := wt.metadata["display"]
			if display == "reset" {
				switch key {
				case "gemini-5h":
					foundG5hR = true
				case "gemini-weekly":
					foundGwkR = true
				case "3p-5h":
					found3p5hR = true
				case "3p-weekly":
					found3pwkR = true
				}
			} else if display == "quota" {
				switch key {
				case "gemini-5h":
					foundG5hP = true
				case "gemini-weekly":
					foundGwkP = true
				case "3p-5h":
					found3p5hP = true
				case "3p-weekly":
					found3pwkP = true
				}
			} else {
				switch key {
				case "gemini-5h":
					foundG5h = true
				case "gemini-weekly":
					foundGwk = true
				case "3p-5h":
					found3p5h = true
				case "3p-weekly":
					found3pwk = true
				}
			}
		} else if wt.wType == "quota-bar" {
			key := wt.metadata["key"]
			switch key {
			case "gemini-5h":
				foundG5hB = true
			case "gemini-weekly":
				foundGwkB = true
			case "3p-5h":
				found3p5hB = true
			case "3p-weekly":
				found3pwkB = true
			}
		}
	}
	if !foundG5h || !foundGwk || !found3p5h || !found3pwk {
		t.Errorf("Expected all 4 quota presets in widgetTypes, got Gemini 5h:%t, Gemini Weekly:%t, 3P 5h:%t, 3P Weekly:%t",
			foundG5h, foundGwk, found3p5h, found3pwk)
	}
	if !foundG5hP || !foundGwkP || !found3p5hP || !found3pwkP {
		t.Errorf("Expected all 4 quota percent presets in widgetTypes, got Gemini 5h Percent:%t, Gemini Weekly Percent:%t, 3P 5h Percent:%t, 3P Weekly Percent:%t",
			foundG5hP, foundGwkP, found3p5hP, found3pwkP)
	}
	if !foundG5hR || !foundGwkR || !found3p5hR || !found3pwkR {
		t.Errorf("Expected all 4 quota reset presets in widgetTypes, got Gemini 5h Reset:%t, Gemini Weekly Reset:%t, 3P 5h Reset:%t, 3P Weekly Reset:%t",
			foundG5hR, foundGwkR, found3p5hR, found3pwkR)
	}
	if !foundG5hB || !foundGwkB || !found3p5hB || !found3pwkB {
		t.Errorf("Expected all 4 quota-bar presets in widgetTypes, got Gemini 5h Bar:%t, Gemini Weekly Bar:%t, 3P 5h Bar:%t, 3P Weekly Bar:%t",
			foundG5hB, foundGwkB, found3p5hB, found3pwkB)
	}

	// 3. 実際に Gemini 5h クォータウィジェットを追加してみる（デフォルト: display指定なし）。
	targetIdx := -1
	for i, wt := range widgetTypes {
		if wt.wType == "quota" && wt.metadata["key"] == "gemini-5h" && wt.metadata["display"] == "" {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		t.Fatalf("Gemini 5h quota widget type not found in widgetTypes")
	}

	// cursor を targetIdx まで移動させる
	m = newModel
	m.cursor = targetIdx

	// Enter を押してウィジェットを追加
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	finalModel := updatedModel.(Model)

	if finalModel.activeMenu != "items" {
		t.Fatalf("Expected activeMenu to return to 'items', got %s", finalModel.activeMenu)
	}

	// 追加されたウィジェットを検証
	addedWidget := finalModel.settings.Lines[0][1]
	if addedWidget.Type != "quota" {
		t.Errorf("Expected widget type 'quota', got '%s'", addedWidget.Type)
	}
	if addedWidget.Metadata == nil || addedWidget.Metadata["key"] != "gemini-5h" || addedWidget.Metadata["display"] != "" {
		t.Errorf("Expected widget metadata key 'gemini-5h' and no display format, got %v", addedWidget.Metadata)
	}

	// 4. 実際に Gemini 5h クォータ % ウィジェットを追加してみる。
	m = finalModel
	m.activeMenu = "items"
	m.cursor = 1

	// "a" キーで追加画面に遷移
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updatedModel.(Model)

	targetPercentIdx := -1
	for i, wt := range widgetTypes {
		if wt.wType == "quota" && wt.metadata["key"] == "gemini-5h" && wt.metadata["display"] == "quota" {
			targetPercentIdx = i
			break
		}
	}
	if targetPercentIdx == -1 {
		t.Fatalf("Gemini 5h quota percent widget type not found in widgetTypes")
	}

	m.cursor = targetPercentIdx
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	finalPercentModel := updatedModel.(Model)

	addedPercentWidget := finalPercentModel.settings.Lines[0][2]
	if addedPercentWidget.Type != "quota" {
		t.Errorf("Expected widget type 'quota', got '%s'", addedPercentWidget.Type)
	}
	if addedPercentWidget.Metadata == nil || addedPercentWidget.Metadata["key"] != "gemini-5h" || addedPercentWidget.Metadata["display"] != "quota" {
		t.Errorf("Expected widget metadata key 'gemini-5h' and display 'quota', got %v", addedPercentWidget.Metadata)
	}

	// 5. 実際に Gemini 5h クォータ Bar ウィジェットを追加してみる。
	m = finalPercentModel
	m.activeMenu = "items"
	m.cursor = 2

	// "a" キーで追加画面に遷移
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updatedModel.(Model)

	targetBarIdx := -1
	for i, wt := range widgetTypes {
		if wt.wType == "quota-bar" && wt.metadata["key"] == "gemini-5h" {
			targetBarIdx = i
			break
		}
	}
	if targetBarIdx == -1 {
		t.Fatalf("Gemini 5h quota-bar widget type not found in widgetTypes")
	}

	m.cursor = targetBarIdx
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	finalBarModel := updatedModel.(Model)

	addedBarWidget := finalBarModel.settings.Lines[0][3]
	if addedBarWidget.Type != "quota-bar" {
		t.Errorf("Expected widget type 'quota-bar', got '%s'", addedBarWidget.Type)
	}
	if addedBarWidget.Color != "" {
		t.Errorf("Expected widget color to be empty, got '%s'", addedBarWidget.Color)
	}
	if addedBarWidget.Metadata == nil || addedBarWidget.Metadata["key"] != "gemini-5h" {
		t.Errorf("Expected widget metadata key 'gemini-5h', got %v", addedBarWidget.Metadata)
	}
}

func TestTUI_LivePreviewQuota(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	settings.Lines[0] = []types.WidgetItem{}
	// クォータウィジェットを追加
	settings.Lines[0] = append(settings.Lines[0],
		types.WidgetItem{
			ID:   "test_quota_g5h",
			Type: "quota",
			Metadata: map[string]string{
				"key": "gemini-5h", // default (percentage + reset)
			},
		},
		types.WidgetItem{
			ID:   "test_quota_g5h_pct",
			Type: "quota",
			Metadata: map[string]string{
				"key":     "gemini-5h",
				"display": "quota", // quota % only
			},
		},
		types.WidgetItem{
			ID:   "test_quota_3p",
			Type: "quota",
			Metadata: map[string]string{
				"key": "3p-5h", // default (percentage + reset)
			},
		},
	)
	m := NewModel(settings, "/tmp/settings.json")

	viewStr := renderer.StripAnsi(m.View())
	// previewCtx のダミーデータから、それぞれ適切な値が表示されることを検証
	// default (both): gemini-5h 50.19% (2h 28m), 3p-5h 100.00% (4h 59m)
	// display:quota (% only): gemini-5h 50.19%
	if !strings.Contains(viewStr, "gemini-5h 50.19% (2h 28m)") {
		t.Errorf("Expected live preview to contain 'gemini-5h 50.19%% (2h 28m)', but it did not. View:\n%s", viewStr)
	}
	if !strings.Contains(viewStr, "gemini-5h 50.19%") {
		t.Errorf("Expected live preview to contain 'gemini-5h 50.19%%', but it did not. View:\n%s", viewStr)
	}
	if !strings.Contains(viewStr, "3p-5h 100.00% (4h 59m)") {
		t.Errorf("Expected live preview to contain '3p-5h 100.00%% (4h 59m)', but it did not. View:\n%s", viewStr)
	}
}

func TestTUI_PowerlineSeparator(t *testing.T) {
	// 1. Default Separator Test
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.json")
	if m.separatorIndex != 1 {
		t.Errorf("Expected default separatorIndex to be 1, got %d", m.separatorIndex)
	}

	// 2. Custom Separator (exists in list) Test
	settings.Powerline.Separators = []string{"\uE0B4"} // Round
	m2 := NewModel(settings, "/tmp/settings.json")
	if m2.separatorIndex != 2 {
		t.Errorf("Expected separatorIndex to be 2 for '\\uE0B4', got %d", m2.separatorIndex)
	}

	// 2.5. None Separator Test
	settings.Powerline.Separators = []string{""}
	mNone := NewModel(settings, "/tmp/settings.json")
	if mNone.separatorIndex != 0 {
		t.Errorf("Expected separatorIndex to be 0 for None, got %d", mNone.separatorIndex)
	}

	// 3. Custom Separator (NOT in list) Test
	settings.Powerline.Separators = []string{"♦"}
	m3 := NewModel(settings, "/tmp/settings.json")
	if m3.separatorIndex == -1 {
		t.Errorf("Expected new custom separator to be added to list, but got index -1")
	}
	customSepName := "Custom (♦)"
	if separatorsList[m3.separatorIndex].name != customSepName {
		t.Errorf("Expected custom separator name to be %q, got %q", customSepName, separatorsList[m3.separatorIndex].name)
	}

	// 4. Test Navigation to Separator Selection Menu in main menu
	m4 := NewModel(settings, "/tmp/settings.json")
	m4.activeMenu = "main"
	m4.cursor = 3 // Select Powerline Separator index

	// Simulate pressing Enter to open separator selection
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")}
	updatedModel, _ := m4.Update(msg)
	newModel := updatedModel.(Model)

	if newModel.activeMenu != "select_separator" {
		t.Errorf("Expected activeMenu to transition to 'select_separator', got %q", newModel.activeMenu)
	}
	if newModel.cursor != m4.separatorIndex {
		t.Errorf("Expected cursor in 'select_separator' menu to match current separatorIndex %d, got %d", m4.separatorIndex, newModel.cursor)
	}
}

func TestTUI_SelectThemeMenu(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	// Let's set default theme to "nord" (index 0)
	settings.Powerline.Theme = "nord"
	m := NewModel(settings, "/tmp/settings.json")
	m.activeMenu = "main"
	m.cursor = 2 // Select Powerline Theme

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
	mTheme0 := NewModel(settings, "/tmp/settings.json")
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
		t.Errorf("Expected Live Preview to be updated dynamically for theme cursor (previewNordPart should differ from previewNordAuroraPart)")
	}

	// Press Enter to confirm selection
	updatedModel, _ = mTheme.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mConfirmed := updatedModel.(Model)

	if mConfirmed.activeMenu != "main" {
		t.Errorf("Expected activeMenu to return to 'main', got %q", mConfirmed.activeMenu)
	}
	if mConfirmed.cursor != 2 {
		t.Errorf("Expected main menu cursor to remain 2, got %d", mConfirmed.cursor)
	}
	if mConfirmed.settings.Powerline.Theme != "nord-aurora" {
		t.Errorf("Expected theme to be 'nord-aurora', got %q", mConfirmed.settings.Powerline.Theme)
	}
	if mConfirmed.themeIndex != 1 {
		t.Errorf("Expected themeIndex to be updated to 1, got %d", mConfirmed.themeIndex)
	}

	// Test Esc key to cancel theme selection
	mCancel := mTheme // cursor at 1 (nord-aurora)
	updatedModel, _ = mCancel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mCancelled := updatedModel.(Model)

	if mCancelled.activeMenu != "main" {
		t.Errorf("Expected activeMenu to return to 'main' on Esc, got %q", mCancelled.activeMenu)
	}
	if mCancelled.cursor != 2 {
		t.Errorf("Expected main menu cursor to return to 2, got %d", mCancelled.cursor)
	}
	// Settings should not have changed
	if mCancelled.settings.Powerline.Theme != "nord" {
		t.Errorf("Expected theme to remain 'nord', got %q", mCancelled.settings.Powerline.Theme)
	}
}

func TestTUI_SelectSeparatorMenu(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	// Let's set default separator to Arrow (index 1)
	settings.Powerline.Separators = []string{"\uE0B0"}
	m := NewModel(settings, "/tmp/settings.json")
	m.activeMenu = "main"
	m.cursor = 3 // Select Powerline Separator

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

	// Verify preview shows Round separator (\uE0B4) before confirmation
	viewSepSelected := mSep.View()
	if !strings.Contains(viewSepSelected, "\uE0B4") {
		t.Errorf("Expected Live Preview to show temporary Round separator '\\uE0B4' while in select_separator menu, but it did not. View:\n%s", viewSepSelected)
	}
	linesSep := strings.Split(viewSepSelected, "\n")
	previewSepPart := strings.Join(linesSep[:4], "\n")
	if strings.Contains(previewSepPart, "\uE0B0") {
		t.Errorf("Expected Live Preview to NOT show Arrow separator '\\uE0B0' when cursor is at Round, but it did. Preview:\n%s", previewSepPart)
	}

	// Press Enter to confirm selection
	updatedModel, _ = mSep.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mConfirmed := updatedModel.(Model)

	if mConfirmed.activeMenu != "main" {
		t.Errorf("Expected activeMenu to return to 'main', got %q", mConfirmed.activeMenu)
	}
	if mConfirmed.cursor != 3 {
		t.Errorf("Expected main menu cursor to remain 3, got %d", mConfirmed.cursor)
	}
	if len(mConfirmed.settings.Powerline.Separators) != 1 || mConfirmed.settings.Powerline.Separators[0] != "\uE0B4" {
		t.Errorf("Expected separator to be updated to '\\uE0B4', got %v", mConfirmed.settings.Powerline.Separators)
	}
	if mConfirmed.separatorIndex != 2 {
		t.Errorf("Expected separatorIndex to be updated to 2, got %d", mConfirmed.separatorIndex)
	}

	// Test Esc key to cancel separator selection
	mCancel := mSep // cursor at 2 (Round)
	updatedModel, _ = mCancel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mCancelled := updatedModel.(Model)

	if mCancelled.activeMenu != "main" {
		t.Errorf("Expected activeMenu to return to 'main' on Esc, got %q", mCancelled.activeMenu)
	}
	if mCancelled.cursor != 3 {
		t.Errorf("Expected main menu cursor to return to 3, got %d", mCancelled.cursor)
	}
	// Settings should not have changed
	if len(mCancelled.settings.Powerline.Separators) != 1 || mCancelled.settings.Powerline.Separators[0] != "\uE0B0" {
		t.Errorf("Expected separator to remain '\\uE0B0', got %v", mCancelled.settings.Powerline.Separators)
	}
}

func TestTUI_SelectStartCapMenu(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	settings.Powerline.StartCaps = []string{"\uE0B2"}
	m := NewModel(settings, "/tmp/settings.json")
	m.activeMenu = "main"
	m.cursor = 4 // Select Powerline Start Cap

	// Enter opens start cap selection
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mCap := updatedModel.(Model)

	if mCap.activeMenu != "select_start_cap" {
		t.Fatalf("Expected activeMenu to transition to 'select_start_cap', got %q", mCap.activeMenu)
	}
	if mCap.cursor != 1 { // index of \uE0B2 in presets
		t.Errorf("Expected cursor to start at current start cap index 1, got %d", mCap.cursor)
	}

	// Move cursor to "Round" (index 2, \uE0B6)
	updatedModel, _ = mCap.Update(tea.KeyMsg{Type: tea.KeyDown})
	mCap = updatedModel.(Model)
	if mCap.cursor != 2 {
		t.Errorf("Expected cursor to move to 2, got %d", mCap.cursor)
	}

	// Verify preview shows Round start cap (\uE0B6) before confirmation
	viewCapSelected := mCap.View()
	if !strings.Contains(viewCapSelected, "\uE0B6") {
		t.Errorf("Expected Live Preview to show temporary Round start cap '\\uE0B6' while in select_start_cap menu. View:\n%s", viewCapSelected)
	}

	// Press Enter to confirm selection
	updatedModel, _ = mCap.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mConfirmed := updatedModel.(Model)

	if mConfirmed.activeMenu != "main" {
		t.Errorf("Expected activeMenu to return to 'main', got %q", mConfirmed.activeMenu)
	}
	if mConfirmed.cursor != 4 {
		t.Errorf("Expected main menu cursor to remain 4, got %d", mConfirmed.cursor)
	}
	if len(mConfirmed.settings.Powerline.StartCaps) != 1 || mConfirmed.settings.Powerline.StartCaps[0] != "\uE0B6" {
		t.Errorf("Expected start cap to be updated to '\\uE0B6', got %v", mConfirmed.settings.Powerline.StartCaps)
	}
	if mConfirmed.startCapIndex != 2 {
		t.Errorf("Expected startCapIndex to be updated to 2, got %d", mConfirmed.startCapIndex)
	}

	// Test Esc key to cancel selection
	mCancel := mCap // cursor at 2 (Round)
	updatedModel, _ = mCancel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mCancelled := updatedModel.(Model)

	if mCancelled.activeMenu != "main" {
		t.Errorf("Expected activeMenu to return to 'main' on Esc, got %q", mCancelled.activeMenu)
	}
	// Settings should not have changed
	if len(mCancelled.settings.Powerline.StartCaps) != 1 || mCancelled.settings.Powerline.StartCaps[0] != "\uE0B2" {
		t.Errorf("Expected start cap to remain '\\uE0B2', got %v", mCancelled.settings.Powerline.StartCaps)
	}
}

func TestTUI_SelectEndCapMenu(t *testing.T) {
	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	settings.Powerline.EndCaps = []string{"\uE0B0"}
	m := NewModel(settings, "/tmp/settings.json")
	m.activeMenu = "main"
	m.cursor = 5 // Select Powerline End Cap

	// Enter opens end cap selection
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mCap := updatedModel.(Model)

	if mCap.activeMenu != "select_end_cap" {
		t.Fatalf("Expected activeMenu to transition to 'select_end_cap', got %q", mCap.activeMenu)
	}
	if mCap.cursor != 1 { // index of \uE0B0 in presets
		t.Errorf("Expected cursor to start at current end cap index 1, got %d", mCap.cursor)
	}

	// Move cursor to "Round" (index 2, \uE0B4)
	updatedModel, _ = mCap.Update(tea.KeyMsg{Type: tea.KeyDown})
	mCap = updatedModel.(Model)
	if mCap.cursor != 2 {
		t.Errorf("Expected cursor to move to 2, got %d", mCap.cursor)
	}

	// Verify preview shows Round end cap (\uE0B4) before confirmation
	viewCapSelected := mCap.View()
	if !strings.Contains(viewCapSelected, "\uE0B4") {
		t.Errorf("Expected Live Preview to show temporary Round end cap '\\uE0B4' while in select_end_cap menu. View:\n%s", viewCapSelected)
	}

	// Press Enter to confirm selection
	updatedModel, _ = mCap.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	mConfirmed := updatedModel.(Model)

	if mConfirmed.activeMenu != "main" {
		t.Errorf("Expected activeMenu to return to 'main', got %q", mConfirmed.activeMenu)
	}
	if mConfirmed.cursor != 5 {
		t.Errorf("Expected main menu cursor to remain 5, got %d", mConfirmed.cursor)
	}
	if len(mConfirmed.settings.Powerline.EndCaps) != 1 || mConfirmed.settings.Powerline.EndCaps[0] != "\uE0B4" {
		t.Errorf("Expected end cap to be updated to '\\uE0B4', got %v", mConfirmed.settings.Powerline.EndCaps)
	}
	if mConfirmed.endCapIndex != 2 {
		t.Errorf("Expected endCapIndex to be updated to 2, got %d", mConfirmed.endCapIndex)
	}

	// Test Esc key to cancel selection
	mCancel := mCap // cursor at 2 (Round)
	updatedModel, _ = mCancel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mCancelled := updatedModel.(Model)

	if mCancelled.activeMenu != "main" {
		t.Errorf("Expected activeMenu to return to 'main' on Esc, got %q", mCancelled.activeMenu)
	}
	// Settings should not have changed
	if len(mCancelled.settings.Powerline.EndCaps) != 1 || mCancelled.settings.Powerline.EndCaps[0] != "\uE0B0" {
		t.Errorf("Expected end cap to remain '\\uE0B0', got %v", mCancelled.settings.Powerline.EndCaps)
	}
}

func TestTUI_LivePreviewAddWidget(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.json")

	// Set state to "add_widget" menu
	m.activeMenu = "add_widget"
	m.selectedLine = 0
	m.itemIndex = 0 // Insert after the first widget (index 0, Model widget)

	// Select "Custom Text" widget which is index 3 in widgetTypes
	m.cursor = 3

	viewStr := m.View()

	// The custom-text widget preview has "Custom Text" as its value.
	// If the preview updates dynamically, it should contain "Custom Text" in the preview section.
	lines := strings.Split(viewStr, "\n")
	if len(lines) < 5 {
		t.Fatalf("Expected view outputs to have enough lines")
	}
	previewPart := strings.Join(lines[:4], "\n")

	if !strings.Contains(previewPart, "Custom Text") {
		t.Errorf("Expected Live Preview to dynamically display the currently selected widget type 'Custom Text' in the preview part, but it did not. Preview part:\n%s", previewPart)
	}
}

func TestTUI_SelectColorLevel(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	settings.ColorLevel = 2 // ANSI 256 colors
	m := NewModel(settings, "/tmp/settings.json")

	// 1. Initial state
	if m.colorLevelIndex != 1 {
		t.Errorf("Expected initial colorLevelIndex to be 1 (ANSI 256 Colors), got %d", m.colorLevelIndex)
	}

	// 2. Select Color Level menu on main menu
	// cursor index for color level will be 6
	// (0: lines, 1: powerline enabled, 2: theme, 3: separator, 4: start cap, 5: end cap, 6: color level, 7: save, 8: discard)
	m.cursor = 6
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
	if mSelected.settings.ColorLevel != 3 {
		t.Errorf("Expected settings.ColorLevel to be updated to 3, got %d", mSelected.settings.ColorLevel)
	}
	if mSelected.colorLevelIndex != 2 {
		t.Errorf("Expected colorLevelIndex to be updated to 2, got %d", mSelected.colorLevelIndex)
	}
	if mSelected.cursor != 6 {
		t.Errorf("Expected cursor in main menu to remain at 6, got %d", mSelected.cursor)
	}

	// 4. Cancel selection with Esc
	mSelectCancel := mSelect
	mSelectCancel.cursor = 0 // ANSI 16 colors
	updatedModel, _ = mSelectCancel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mCancelled := updatedModel.(Model)

	if mCancelled.activeMenu != "main" {
		t.Errorf("Expected activeMenu to return to 'main' on Esc, got %q", mCancelled.activeMenu)
	}
	if mCancelled.settings.ColorLevel != 2 {
		t.Errorf("Expected settings.ColorLevel to remain 2, got %d", mCancelled.settings.ColorLevel)
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
		types.WidgetItem{ID: "test_sandbox", Type: "sandbox"},
	)
	m := NewModel(settings, "/tmp/settings.json")

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
		{ID: "w1", Type: "model"},
		{ID: "w2", Type: "git-changes"},
		{ID: "w3", Type: "git-branch"},
	}
	widgetsSlice := make([]types.WidgetItem, 3, 10)
	copy(widgetsSlice, originalWidgets)

	settings := types.DefaultSettings()
	settings.Lines = [][]types.WidgetItem{widgetsSlice}

	m := NewModel(settings, "/tmp/settings.json")
	m.selectedLine = 0
	m.itemIndex = 0 // Insert after index 0 (so between w1 and w2)

	// --- Test 1: Live Preview should not corrupt original settings ---
	m.activeMenu = "add_widget"
	m.cursor = 0 // first widget type to add

	// Verify pre-conditions
	if m.settings.Lines[0][1].ID != "w2" || m.settings.Lines[0][2].ID != "w3" {
		t.Fatalf("Pre-condition failed: settings initialized incorrectly")
	}

	// Trigger preview render (which calls View and performs a temporary insert)
	_ = m.View()

	// Verify that the original settings line was not mutated by the preview logic
	if len(m.settings.Lines[0]) != 3 {
		t.Errorf("Expected original settings line length to remain 3 after preview, but got %d", len(m.settings.Lines[0]))
	}
	if m.settings.Lines[0][1].ID != "w2" {
		t.Errorf("Expected original widget 'w2' to be untouched, but got ID %q", m.settings.Lines[0][1].ID)
	}
	if m.settings.Lines[0][2].ID != "w3" {
		t.Errorf("Expected original widget 'w3' to be untouched, but got ID %q", m.settings.Lines[0][2].ID)
	}

	// --- Test 2: Actually adding the widget should correctly insert it without corruption ---
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := m.Update(enterMsg)
	newModel := updatedModel.(Model)

	newWidgets := newModel.settings.Lines[0]
	if len(newWidgets) != 4 {
		t.Fatalf("Expected new settings line length to be 4, got %d", len(newWidgets))
	}

	// Expected sequence: w1 -> new_widget -> w2 -> w3
	if newWidgets[0].ID != "w1" {
		t.Errorf("Expected index 0 to be 'w1', got %q", newWidgets[0].ID)
	}
	// The new widget should not have w2's or w3's ID
	if newWidgets[1].ID == "w2" || newWidgets[1].ID == "w3" || newWidgets[1].ID == "w1" {
		t.Errorf("Expected index 1 to be a newly added widget, got ID %q", newWidgets[1].ID)
	}
	if newWidgets[2].ID != "w2" {
		t.Errorf("Expected index 2 to be 'w2', got %q", newWidgets[2].ID)
	}
	if newWidgets[3].ID != "w3" {
		t.Errorf("Expected index 3 to be 'w3', got %q", newWidgets[3].ID)
	}

	// Check that the underlying original widgets slice was not mutated (no in-place overwrite)
	if widgetsSlice[1].ID != "w2" {
		t.Errorf("Expected original widgetsSlice elements to remain untouched, but index 1 got ID %q", widgetsSlice[1].ID)
	}
}

func TestTUI_NoASCIIInSeparators(t *testing.T) {
	for _, sep := range separatorsList {
		if sep.value == "/" {
			t.Errorf("Slash ASCII (/) separator should be removed")
		}
		if sep.value == "|" {
			t.Errorf("Bar ASCII (|) separator should be removed")
		}
	}
}
