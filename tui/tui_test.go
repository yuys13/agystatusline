package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
	m := NewModel(settings, "/tmp/settings.json")

	// Send key event "q"
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	updatedModel, cmd := m.Update(msg)
	
	newModel := updatedModel.(Model)
	if !newModel.quitting {
		t.Errorf("Expected quitting to be true after pressing 'q'")
	}
	
	if cmd == nil {
		t.Log("Command is nil, which is expected for normal quitting")
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

func TestTUI_LivePreviewContextPercentages(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	// Append the new widgets to the first line to test their preview rendering
	settings.Lines[0] = append(settings.Lines[0],
		types.WidgetItem{ID: "test_used", Type: "context-used-pct"},
		types.WidgetItem{ID: "test_remaining", Type: "context-remaining-pct"},
	)
	m := NewModel(settings, "/tmp/settings.json")

	viewStr := m.View()
	if !strings.Contains(viewStr, "Used: 20.00%") {
		t.Errorf("Expected live preview to contain 'Used: 20.00%%', but it did not. View output:\n%s", viewStr)
	}
	if !strings.Contains(viewStr, "Remaining: 80.00%") {
		t.Errorf("Expected live preview to contain 'Remaining: 80.00%%', but it did not. View output:\n%s", viewStr)
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
	m := NewModel(settings, "/tmp/settings.json")
	m.activeMenu = "items"
	m.selectedLine = 0
	m.cursor = 2 // Pointing to context-length

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
	for i := 0; i < 4; i++ {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updatedModel.(Model)
	}
	if m.cursor != 4 {
		t.Fatalf("Expected cursor to be 4, got %d", m.cursor)
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
	if newModel.settings.Lines[0][2].Type != "separator" {
		t.Errorf("Expected added widget at index 2 to be 'separator', got %q", newModel.settings.Lines[0][2].Type)
	}
	if newModel.cursor != 2 {
		t.Errorf("Expected cursor to point to the newly added widget (index 2), got %d", newModel.cursor)
	}
}

func TestTUI_AddContextPctWidgets(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.json")
	m.activeMenu = "items"
	m.selectedLine = 0
	m.cursor = 0

	// Press "a" to trigger Add Widget screen
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	updatedModel, _ := m.Update(msg)
	newModel := updatedModel.(Model)

	// Navigate to the bottom (where Context Used % is at index 6)
	m = newModel
	for i := 0; i < 6; i++ {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updatedModel.(Model)
	}
	if m.cursor != 6 {
		t.Fatalf("Expected cursor to be 6, got %d", m.cursor)
	}

	// Press Enter to add
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\n")})
	newModel = updatedModel.(Model)

	if newModel.activeMenu != "items" {
		t.Errorf("Expected activeMenu to return to 'items', got %q", newModel.activeMenu)
	}
	addedWidget := newModel.settings.Lines[0][1]
	if addedWidget.Type != "context-used-pct" {
		t.Errorf("Expected added widget type to be 'context-used-pct', got %q", addedWidget.Type)
	}
	if addedWidget.Color != "brightBlack" {
		t.Errorf("Expected added widget color to be 'brightBlack', got %q", addedWidget.Color)
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

	// 2. ウィジェット追加リストにクォータウィジェットが含まれているか確認
	var foundG5h, foundGwk, found3p5h, found3pwk bool
	for _, wt := range widgetTypes {
		if wt.wType == "quota" {
			switch wt.metadata["key"] {
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
	}
	if !foundG5h || !foundGwk || !found3p5h || !found3pwk {
		t.Errorf("Expected all 4 quota presets in widgetTypes, got Gemini 5h:%t, Gemini Weekly:%t, 3P 5h:%t, 3P Weekly:%t",
			foundG5h, foundGwk, found3p5h, found3pwk)
	}

	// 3. 実際に Gemini 5h クォータウィジェットを追加してみる。
	targetIdx := -1
	for i, wt := range widgetTypes {
		if wt.wType == "quota" && wt.metadata["key"] == "gemini-5h" {
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
	if addedWidget.Metadata == nil || addedWidget.Metadata["key"] != "gemini-5h" {
		t.Errorf("Expected widget metadata key 'gemini-5h', got %v", addedWidget.Metadata)
	}
}

func TestTUI_LivePreviewQuota(t *testing.T) {
	widgets.RegisterAll()
	settings := types.DefaultSettings()
	// クォータウィジェットを追加
	settings.Lines[0] = append(settings.Lines[0],
		types.WidgetItem{
			ID:   "test_quota_g5h",
			Type: "quota",
			Metadata: map[string]string{
				"key": "gemini-5h",
			},
		},
		types.WidgetItem{
			ID:   "test_quota_3p",
			Type: "quota",
			Metadata: map[string]string{
				"key": "3p-5h",
			},
		},
		types.WidgetItem{
			ID:   "test_quota_g_wk",
			Type: "quota",
			Metadata: map[string]string{
				"key": "gemini-weekly",
			},
		},
		types.WidgetItem{
			ID:   "test_quota_3p_wk",
			Type: "quota",
			Metadata: map[string]string{
				"key": "3p-weekly",
			},
		},
	)
	m := NewModel(settings, "/tmp/settings.json")

	viewStr := m.View()
	// previewCtx のダミーデータから、それぞれ適切な値が表示されることを検証
	if !strings.Contains(viewStr, "gemini-5h: 50.19%") {
		t.Errorf("Expected live preview to contain 'gemini-5h: 50.19%%', but it did not. View:\n%s", viewStr)
	}
	if !strings.Contains(viewStr, "3p-5h: 100.00%") {
		t.Errorf("Expected live preview to contain '3p-5h: 100.00%%', but it did not. View:\n%s", viewStr)
	}
	if !strings.Contains(viewStr, "gemini-weekly: 90.91%") {
		t.Errorf("Expected live preview to contain 'gemini-weekly: 90.91%%', but it did not. View:\n%s", viewStr)
	}
	if !strings.Contains(viewStr, "3p-weekly: 100.00%") {
		t.Errorf("Expected live preview to contain '3p-weekly: 100.00%%', but it did not. View:\n%s", viewStr)
	}
}




