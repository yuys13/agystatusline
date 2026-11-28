package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yuys13/agystatusline/types"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			if m.activeMenu == menuMain || m.activeMenu == menuCorruptedConfig {
				m.quitting = true
				return m, tea.Quit
			}
		}

		switch m.activeMenu {
		case menuCorruptedConfig:
			return m.updateCorruptedConfig(msg)
		case menuMain:
			return m.updateMain(msg)
		case menuPowerline:
			return m.updatePowerline(msg)
		case menuLines:
			return m.updateLines(msg)
		case menuItems:
			return m.updateItems(msg)
		case menuAddWidget:
			return m.updateAddWidget(msg)
		case menuSelectTheme:
			return m.updateSelectTheme(msg)
		case menuSelectSeparator:
			return m.updateSelectSeparator(msg)
		case menuSelectStartCap:
			return m.updateSelectStartCap(msg)
		case menuSelectEndCap:
			return m.updateSelectEndCap(msg)
		case menuSelectColorLevel:
			return m.updateSelectColorLevel(msg)
		}
	}
	return m, nil
}

func (m Model) updateCorruptedConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < 1 {
			m.cursor++
		}

	case "enter", "\n":
		switch m.cursor {
		case 0: // Load Default Settings
			m.settings = types.DefaultSettings()
			m.activeMenu = menuMain
			m.cursor = 0

		case 1: // Exit
			m.quitting = true
			return m, tea.Quit
		}

	case "esc":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		maxItems := 5
		if m.cursor < maxItems-1 {
			m.cursor++
		}

	case "enter", "\n":
		switch m.cursor {
		case 0: // Edit Lines
			m.activeMenu = menuLines
			m.cursor = 0
			m.moveMode = false

		case 1: // Powerline Settings
			m.activeMenu = menuPowerline
			m.cursor = 0

		case 2: // Select Color Level
			m.activeMenu = menuSelectColorLevel
			m.cursor = m.colorLevelIndex

		case 3: // Save & Exit
			err := saveSettings(m.configPath, m.settings)
			if err == nil {
				m.saved = true
			}
			m.quitting = true
			return m, tea.Quit

		case 4: // Discard & Exit
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) updatePowerline(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		maxItems := 6
		if m.cursor < maxItems-1 {
			m.cursor++
		}

	case "enter", "\n":
		switch m.cursor {
		case 0: // Toggle Powerline Mode
			m.settings.Powerline.Enabled = !m.settings.Powerline.Enabled

		case 1: // Select Powerline Theme
			m.activeMenu = menuSelectTheme
			m.cursor = m.themeIndex

		case 2: // Select Powerline Separator
			m.activeMenu = menuSelectSeparator
			m.cursor = m.separatorIndex

		case 3: // Select Powerline Start Cap
			m.activeMenu = menuSelectStartCap
			m.cursor = m.startCapIndex

		case 4: // Select Powerline End Cap
			m.activeMenu = menuSelectEndCap
			m.cursor = m.endCapIndex

		case 5: // Back to Main Menu
			m.activeMenu = menuMain
			m.cursor = 1
		}

	case "esc":
		m.activeMenu = menuMain
		m.cursor = 1
	}
	return m, nil
}

func (m Model) updateLines(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	linesCount := len(m.settings.Lines)

	switch msg.String() {
	case "up", "k":
		if m.moveMode {
			if m.cursor > 0 && linesCount > 1 {
				m.settings.Lines[m.cursor], m.settings.Lines[m.cursor-1] = m.settings.Lines[m.cursor-1], m.settings.Lines[m.cursor]
				m.cursor--
			}
		} else {
			if m.cursor > 0 {
				m.cursor--
			}
		}

	case "down", "j":
		if m.moveMode {
			if m.cursor < linesCount-1 && linesCount > 1 {
				m.settings.Lines[m.cursor], m.settings.Lines[m.cursor+1] = m.settings.Lines[m.cursor+1], m.settings.Lines[m.cursor]
				m.cursor++
			}
		} else {
			if m.cursor < linesCount-1 {
				m.cursor++
			}
		}

	case "a":
		m.settings.Lines = append(m.settings.Lines, []types.WidgetItem{})
		m.cursor = len(m.settings.Lines) - 1
		m.moveMode = false

	case "d":
		if linesCount > 1 && m.cursor < linesCount {
			newLines := make([][]types.WidgetItem, 0, linesCount-1)
			newLines = append(newLines, m.settings.Lines[:m.cursor]...)
			newLines = append(newLines, m.settings.Lines[m.cursor+1:]...)
			m.settings.Lines = newLines
			if m.cursor >= len(m.settings.Lines) {
				m.cursor = len(m.settings.Lines) - 1
			}
			m.moveMode = false
		}

	case "m":
		if linesCount > 1 {
			m.moveMode = !m.moveMode
		}

	case "enter", "\n":
		if m.moveMode {
			m.moveMode = false
		} else if m.cursor < linesCount {
			m.selectedLine = m.cursor
			m.activeMenu = menuItems
			m.cursor = 0
			m.moveMode = false
		}

	case "esc":
		if m.moveMode {
			m.moveMode = false
		} else {
			m.activeMenu = menuMain
			m.cursor = 0
		}
	}
	return m, nil
}

func (m Model) updateItems(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	widgets := m.settings.Lines[m.selectedLine]
	widgetsCount := len(widgets)

	switch msg.String() {
	case "up", "k":
		if m.moveMode {
			if m.cursor > 0 && widgetsCount > 1 {
				m.settings.Lines[m.selectedLine][m.cursor], m.settings.Lines[m.selectedLine][m.cursor-1] = m.settings.Lines[m.selectedLine][m.cursor-1], m.settings.Lines[m.selectedLine][m.cursor]
				m.cursor--
			}
		} else {
			if m.cursor > 0 {
				m.cursor--
			}
		}

	case "down", "j":
		if m.moveMode {
			if m.cursor < widgetsCount-1 && widgetsCount > 1 {
				m.settings.Lines[m.selectedLine][m.cursor], m.settings.Lines[m.selectedLine][m.cursor+1] = m.settings.Lines[m.selectedLine][m.cursor+1], m.settings.Lines[m.selectedLine][m.cursor]
				m.cursor++
			}
		} else {
			if m.cursor < widgetsCount-1 {
				m.cursor++
			}
		}

	case "a":
		m.itemIndex = m.cursor
		m.activeMenu = menuAddWidget
		m.cursor = 0
		m.moveMode = false

	case "d":
		if widgetsCount > 0 && m.cursor < widgetsCount {
			newWidgets := make([]types.WidgetItem, 0, widgetsCount-1)
			newWidgets = append(newWidgets, widgets[:m.cursor]...)
			newWidgets = append(newWidgets, widgets[m.cursor+1:]...)
			m.settings.Lines[m.selectedLine] = newWidgets
			if m.cursor >= len(m.settings.Lines[m.selectedLine]) {
				m.cursor = len(m.settings.Lines[m.selectedLine]) - 1
			}
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.moveMode = false
		}

	case "m":
		if widgetsCount > 1 {
			m.moveMode = !m.moveMode
		}

	case "enter", "\n":
		if m.moveMode {
			m.moveMode = false
		}

	case "esc":
		if m.moveMode {
			m.moveMode = false
		} else {
			m.activeMenu = menuLines
			m.cursor = m.selectedLine
		}
	}
	return m, nil
}

// insertWidget returns a new slice with newWidget inserted after itemIndex (or at 0 when empty)
// without mutating the backing array of widgets.
func insertWidget(widgets []types.WidgetItem, itemIndex int, newWidget types.WidgetItem) ([]types.WidgetItem, int) {
	insertIndex := 0
	if len(widgets) > 0 {
		insertIndex = itemIndex + 1
	}

	newWidgets := make([]types.WidgetItem, len(widgets)+1)
	copy(newWidgets[:insertIndex], widgets[:insertIndex])
	newWidgets[insertIndex] = newWidget
	copy(newWidgets[insertIndex+1:], widgets[insertIndex:])
	return newWidgets, insertIndex
}

func (m Model) updateAddWidget(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(widgetTypes)-1 {
			m.cursor++
		}

	case "enter", "\n":
		selectedType := widgetTypes[m.cursor]
		newWidgets, insertIndex := insertWidget(m.settings.Lines[m.selectedLine], m.itemIndex, selectedType.item)

		m.settings.Lines[m.selectedLine] = newWidgets
		m.activeMenu = menuItems
		m.cursor = insertIndex

	case "esc":
		m.activeMenu = menuItems
		m.cursor = m.itemIndex
	}
	return m, nil
}

func (m Model) updateSelectTheme(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(themesList)-1 {
			m.cursor++
		}

	case "enter", "\n":
		m.themeIndex = m.cursor
		m.settings.Powerline.Theme = themesList[m.themeIndex]
		m.activeMenu = menuPowerline
		m.cursor = 1

	case "esc":
		m.activeMenu = menuPowerline
		m.cursor = 1
	}
	return m, nil
}

func (m Model) updateSelectSeparator(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.separators)-1 {
			m.cursor++
		}

	case "enter", "\n":
		m.separatorIndex = m.cursor
		m.settings.Powerline.Separator = m.separators[m.separatorIndex].value
		m.activeMenu = menuPowerline
		m.cursor = 2

	case "esc":
		m.activeMenu = menuPowerline
		m.cursor = 2
	}
	return m, nil
}

func (m Model) updateSelectStartCap(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.startCaps)-1 {
			m.cursor++
		}

	case "enter", "\n":
		m.startCapIndex = m.cursor
		m.settings.Powerline.StartCaps = m.startCaps[m.startCapIndex].value
		m.activeMenu = menuPowerline
		m.cursor = 3

	case "esc":
		m.activeMenu = menuPowerline
		m.cursor = 3
	}
	return m, nil
}

func (m Model) updateSelectEndCap(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.endCaps)-1 {
			m.cursor++
		}

	case "enter", "\n":
		m.endCapIndex = m.cursor
		m.settings.Powerline.EndCaps = m.endCaps[m.endCapIndex].value
		m.activeMenu = menuPowerline
		m.cursor = 4

	case "esc":
		m.activeMenu = menuPowerline
		m.cursor = 4
	}
	return m, nil
}

func (m Model) updateSelectColorLevel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(colorLevelsList)-1 {
			m.cursor++
		}

	case "enter", "\n":
		m.colorLevelIndex = m.cursor
		m.settings.General.ColorLevel = colorLevelsList[m.colorLevelIndex].value
		m.activeMenu = menuMain
		m.cursor = 2

	case "esc":
		m.activeMenu = menuMain
		m.cursor = 2
	}
	return m, nil
}
