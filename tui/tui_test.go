package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yuys13/agystatusline/types"
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
