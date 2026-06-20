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

