package tui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yuys13/agystatusline/renderer"
	"github.com/yuys13/agystatusline/types"
)

var (
	titleStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	errorTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	previewHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244"))
	selectedStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))
	moveModeStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208"))
	plainStyle         = lipgloss.NewStyle()
)

func (m Model) buildPreviewSettings() types.Settings {
	previewSettings := m.settings
	switch m.activeMenu {
	case menuSelectTheme:
		if m.cursor >= 0 && m.cursor < len(themesList) {
			previewSettings.Powerline.Theme = themesList[m.cursor]
		}
	case menuSelectSeparator:
		if m.cursor >= 0 && m.cursor < len(m.separators) {
			previewSettings.Powerline.Separator = m.separators[m.cursor].value
		}
	case menuSelectStartCap:
		if m.cursor >= 0 && m.cursor < len(m.startCaps) {
			previewSettings.Powerline.StartCaps = m.startCaps[m.cursor].value
		}
	case menuSelectEndCap:
		if m.cursor >= 0 && m.cursor < len(m.endCaps) {
			previewSettings.Powerline.EndCaps = m.endCaps[m.cursor].value
		}
	case menuSelectColorLevel:
		if m.cursor >= 0 && m.cursor < len(colorLevelsList) {
			previewSettings.General.ColorLevel = colorLevelsList[m.cursor].value
		}
	case menuAddWidget:
		if m.cursor >= 0 && m.cursor < len(widgetTypes) && m.selectedLine >= 0 && m.selectedLine < len(m.settings.Lines) {
			selectedType := widgetTypes[m.cursor]
			newWidgets, _ := insertWidget(m.settings.Lines[m.selectedLine], m.itemIndex, selectedType.item)

			// Clone Lines slice to prevent modifying the shared layout
			newLines := slices.Clone(previewSettings.Lines)
			newLines[m.selectedLine] = newWidgets
			previewSettings.Lines = newLines
		}
	}
	return previewSettings
}

func (m Model) View() string {
	if m.quitting {
		if m.saved {
			return "\x1b[0mConfiguration saved successfully. Exiting...\n"
		}
		return "\x1b[0mChanges discarded. Exiting...\n"
	}

	var s strings.Builder

	if m.activeMenu == menuCorruptedConfig {
		m.viewCorruptedConfig(&s)
		return s.String()
	}

	// Render Live Preview at the top
	s.WriteString(previewHeaderStyle.Render("--- Live Preview ---"))
	s.WriteString("\n")

	previewCtx := DefaultPreviewContext(m.settings)
	previewSettings := m.buildPreviewSettings()

	previewLines := renderer.RenderStatusLines(previewSettings, previewCtx)
	for _, line := range previewLines {
		s.WriteString("\x1b[0m" + line)
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// Render the active menu screen below the preview
	switch m.activeMenu {
	case menuMain:
		m.viewMain(&s)
	case menuPowerline:
		m.viewPowerline(&s)
	case menuLines:
		m.viewLines(&s)
	case menuItems:
		m.viewItems(&s)
	case menuAddWidget:
		m.viewAddWidget(&s)
	case menuSelectTheme:
		m.viewSelectTheme(&s)
	case menuSelectSeparator:
		m.viewSelectSeparator(&s)
	case menuSelectStartCap:
		m.viewSelectStartCap(&s)
	case menuSelectEndCap:
		m.viewSelectEndCap(&s)
	case menuSelectColorLevel:
		m.viewSelectColorLevel(&s)
	}

	return s.String()
}

func (m Model) viewCorruptedConfig(s *strings.Builder) {
	s.WriteString(errorTitleStyle.Render("⚠️  Configuration File Corrupted"))
	s.WriteString("\n\n")
	fmt.Fprintf(s, "Failed to parse configuration file (%s): %s\n", m.configPath, m.loadError)
	s.WriteString("Would you like to load default settings or exit?\n\n")

	menuItems := []string{
		"🔄 Load Default Settings",
		"❌ Exit",
	}

	for i, item := range menuItems {
		cursorStr := " "
		style := plainStyle
		if m.cursor == i {
			cursorStr = ">"
			style = selectedStyle
		}
		fmt.Fprintf(s, "%s %s\n", cursorStr, style.Render(item))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to select, q/Esc to quit)\n")
}

func (m Model) viewMain(s *strings.Builder) {
	s.WriteString(titleStyle.Render("agystatusline Configuration Menu"))
	s.WriteString("\n\n")

	menuItems := []string{
		"📄 Edit Lines",
		fmt.Sprintf("⚡ Powerline Settings          [%s / %t]", m.settings.Powerline.Theme, m.settings.Powerline.Enabled),
		fmt.Sprintf("🎨 Select Color Level          [%s]", colorLevelsList[m.colorLevelIndex].name),
		"💾 Save & Exit",
		"❌ Discard & Exit",
	}

	for i, item := range menuItems {
		cursorStr := " "
		style := plainStyle
		if m.cursor == i {
			cursorStr = ">"
			style = selectedStyle
		}
		if item == "💾 Save & Exit" {
			s.WriteString("\n")
		}
		fmt.Fprintf(s, "%s %s\n", cursorStr, style.Render(item))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to select, q to quit)\n")
}

func (m Model) viewPowerline(s *strings.Builder) {
	s.WriteString(titleStyle.Render("Powerline Settings"))
	s.WriteString("\n\n")

	menuItems := []string{
		fmt.Sprintf("Toggle Powerline Mode       [%t]", m.settings.Powerline.Enabled),
		fmt.Sprintf("Select Powerline Theme      [%s]", m.settings.Powerline.Theme),
		fmt.Sprintf("Select Powerline Separator  [%s]", m.separators[m.separatorIndex].name),
		fmt.Sprintf("Select Powerline Start Cap  [%s]", m.startCaps[m.startCapIndex].name),
		fmt.Sprintf("Select Powerline End Cap    [%s]", m.endCaps[m.endCapIndex].name),
		"Back to Main Menu",
	}

	for i, item := range menuItems {
		cursorStr := " "
		style := plainStyle
		if m.cursor == i {
			cursorStr = ">"
			style = selectedStyle
		}
		if item == "Back to Main Menu" {
			s.WriteString("\n")
		}
		fmt.Fprintf(s, "%s %s\n", cursorStr, style.Render(item))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to toggle/select, Esc to go back)\n")
}

func (m Model) viewLines(s *strings.Builder) {
	s.WriteString(titleStyle.Render("Select Line to Edit Items"))
	s.WriteString("\n\n")

	for i, line := range m.settings.Lines {
		cursorStr := " "
		style := plainStyle
		if m.cursor == i {
			if m.moveMode {
				cursorStr = "M"
				style = moveModeStyle
			} else {
				cursorStr = ">"
				style = selectedStyle
			}
		}

		var widgetStr string
		if len(line) == 0 {
			widgetStr = "(empty)"
		} else {
			for _, w := range line {
				widgetStr += fmt.Sprintf("[%s] ", w.Type)
			}
		}

		fmt.Fprintf(s, "%s Line %d: %s\n", cursorStr, i+1, style.Render(widgetStr))
	}

	s.WriteString("\n(a: add line, d: delete line, m: toggle move mode, Enter: edit widgets, Esc: back)\n")
}

func (m Model) viewItems(s *strings.Builder) {
	s.WriteString(titleStyle.Render(fmt.Sprintf("Editing Line %d Items", m.selectedLine+1)))
	s.WriteString("\n\n")

	widgets := m.settings.Lines[m.selectedLine]
	if len(widgets) == 0 {
		s.WriteString("  (No widgets in this line)\n")
	} else {
		for i, w := range widgets {
			cursorStr := " "
			style := plainStyle
			if m.cursor == i {
				if m.moveMode {
					cursorStr = "M"
					style = moveModeStyle
				} else {
					cursorStr = ">"
					style = selectedStyle
				}
			}

			detailStr := fmt.Sprintf("[%d] %s", i+1, w.Type)
			if w.Color != "" {
				detailStr += fmt.Sprintf(" (color: %s)", w.Color)
			}
			fmt.Fprintf(s, "%s %s\n", cursorStr, style.Render(detailStr))
		}
	}

	s.WriteString("\n(a: add widget, d: delete widget, m: toggle move mode, Esc: back)\n")
}

func (m Model) viewAddWidget(s *strings.Builder) {
	s.WriteString(titleStyle.Render("Select Widget Type to Add"))
	s.WriteString("\n\n")

	for i, t := range widgetTypes {
		cursorStr := " "
		style := plainStyle
		if m.cursor == i {
			cursorStr = ">"
			style = selectedStyle
		}
		fmt.Fprintf(s, "%s %s\n", cursorStr, style.Render(t.name))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to add, Esc: cancel)\n")
}

func (m Model) viewSelectTheme(s *strings.Builder) {
	s.WriteString(titleStyle.Render("Select Powerline Theme"))
	s.WriteString("\n\n")

	for i, t := range themesList {
		cursorStr := " "
		style := plainStyle
		if m.cursor == i {
			cursorStr = ">"
			style = selectedStyle
		}
		themeStr := t
		if t == m.settings.Powerline.Theme {
			themeStr += " (active)"
		}
		fmt.Fprintf(s, "%s %s\n", cursorStr, style.Render(themeStr))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to select, Esc: cancel)\n")
}

func (m Model) writeChoiceListView(s *strings.Builder, title string, items []choiceItem, activeValue string) {
	s.WriteString(titleStyle.Render(title))
	s.WriteString("\n\n")

	for i, item := range items {
		cursorStr := " "
		style := plainStyle
		if m.cursor == i {
			cursorStr = ">"
			style = selectedStyle
		}
		label := item.name
		if item.value == activeValue {
			label += " (active)"
		}
		fmt.Fprintf(s, "%s %s\n", cursorStr, style.Render(label))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to select, Esc: cancel)\n")
}

func (m Model) viewSelectSeparator(s *strings.Builder) {
	m.writeChoiceListView(s, "Select Powerline Separator", m.separators, m.settings.Powerline.Separator)
}

func (m Model) viewSelectStartCap(s *strings.Builder) {
	m.writeChoiceListView(s, "Select Powerline Start Cap", m.startCaps, m.settings.Powerline.StartCaps)
}

func (m Model) viewSelectEndCap(s *strings.Builder) {
	m.writeChoiceListView(s, "Select Powerline End Cap", m.endCaps, m.settings.Powerline.EndCaps)
}

func (m Model) viewSelectColorLevel(s *strings.Builder) {
	s.WriteString(titleStyle.Render("Select Color Level"))
	s.WriteString("\n\n")

	for i, cl := range colorLevelsList {
		cursorStr := " "
		style := plainStyle
		if m.cursor == i {
			cursorStr = ">"
			style = selectedStyle
		}
		clStr := cl.name
		if cl.value == m.settings.General.ColorLevel {
			clStr += " (active)"
		}
		fmt.Fprintf(s, "%s %s\n", cursorStr, style.Render(clStr))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to select, Esc: cancel)\n")
}
