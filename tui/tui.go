package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yuys13/agystatusline/renderer"
	"github.com/yuys13/agystatusline/types"
)

type Model struct {
	settings        types.Settings
	configPath      string
	activeMenu      string
	cursor          int
	quitting        bool
	saved           bool
	themeIndex      int
	separatorIndex  int
	startCapIndex   int
	endCapIndex     int
	colorLevelIndex int
	selectedLine    int
	itemIndex       int
	moveMode        bool
}

var themesList = []string{"nord", "nord-aurora", "monokai", "solarized", "minimal", "dracula", "catppuccin", "gruvbox", "onedark", "tokyonight"}

var colorLevelsList = []struct {
	name  string
	value int
}{
	{name: "ANSI 16 Colors", value: 1},
	{name: "ANSI 256 Colors (Default)", value: 2},
	{name: "Truecolor (24-bit)", value: 3},
}

var separatorsList = []struct {
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

var startCapsList = []struct {
	name  string
	value string
}{
	{name: "None", value: ""},
	{name: "Triangle (\uE0B2)", value: "\uE0B2"},
	{name: "Round (\uE0B6)", value: "\uE0B6"},
	{name: "Lower Triangle (\uE0BA)", value: "\uE0BA"},
	{name: "Diagonal (\uE0BE)", value: "\uE0BE"},
	{name: "Flame (\uE0C2)", value: "\uE0C2 "},
	{name: "Hexagon (\uE0C7)", value: "\uE0C7 "},
	{name: "Slanted (\uE0CA)", value: "\uE0CA "},
}

var endCapsList = []struct {
	name  string
	value string
}{
	{name: "None", value: ""},
	{name: "Triangle (\uE0B0)", value: "\uE0B0"},
	{name: "Round (\uE0B4)", value: "\uE0B4"},
	{name: "Lower Triangle (\uE0B8)", value: "\uE0B8"},
	{name: "Diagonal (\uE0BC)", value: "\uE0BC"},
	{name: "Flame (\uE0C0)", value: "\uE0C0"},
	{name: "Hexagon (\uE0C6)", value: "\uE0C6"},
	{name: "Slanted (\uE0C8)", value: "\uE0C8"},
}

var widgetTypes = []struct {
	name            string
	wType           string
	color           string
	backgroundColor string
	customText      string
	metadata        map[string]string
}{
	{name: "Model", wType: "model", color: "brightMagenta"},
	{name: "Git Branch", wType: "git-branch", color: "brightMagenta"},
	{name: "Git Changes", wType: "git-changes", color: "yellow"},
	{name: "Quota Bar: 5h", wType: "quota-bar", color: "", metadata: map[string]string{"key": "gemini-5h"}},
	{name: "Quota Bar: Weekly", wType: "quota-bar", color: "", metadata: map[string]string{"key": "gemini-weekly"}},
	{name: "Quota Bar: 3P 5h", wType: "quota-bar", color: "", metadata: map[string]string{"key": "3p-5h"}},
	{name: "Quota Bar: 3P Weekly", wType: "quota-bar", color: "", metadata: map[string]string{"key": "3p-weekly"}},
	{name: "Sandbox Enabled", wType: "sandbox", color: "yellow"},
	{name: "Agent State", wType: "agent-state", color: "brightGreen"},
	{name: "Context Bar", wType: "context-bar", color: "brightWhite"},
	{name: "Artifacts", wType: "artifacts", color: "brightWhite"},
	{name: "Subagents", wType: "subagents", color: "brightWhite"},
	{name: "Tasks", wType: "tasks", color: "brightWhite"},
}

func NewModel(settings types.Settings, configPath string) Model {
	initialThemeIndex := 0
	for i, t := range themesList {
		if t == settings.Powerline.Theme {
			initialThemeIndex = i
			break
		}
	}

	initialSeparatorIndex := -1
	currentSep := "\uE0B0"
	if len(settings.Powerline.Separators) > 0 {
		currentSep = settings.Powerline.Separators[0]
	}
	for i, s := range separatorsList {
		if strings.TrimRight(s.value, " ") == strings.TrimRight(currentSep, " ") {
			initialSeparatorIndex = i
			break
		}
	}
	if initialSeparatorIndex == -1 {
		separatorsList = append(separatorsList, struct {
			name  string
			value string
		}{name: fmt.Sprintf("Custom (%s)", currentSep), value: currentSep})
		initialSeparatorIndex = len(separatorsList) - 1
	}

	initialStartCapIndex := -1
	currentStartCap := ""
	if len(settings.Powerline.StartCaps) > 0 {
		currentStartCap = settings.Powerline.StartCaps[0]
	}
	for i, s := range startCapsList {
		if s.value == currentStartCap {
			initialStartCapIndex = i
			break
		}
	}
	if initialStartCapIndex == -1 {
		startCapsList = append(startCapsList, struct {
			name  string
			value string
		}{name: fmt.Sprintf("Custom (%s)", currentStartCap), value: currentStartCap})
		initialStartCapIndex = len(startCapsList) - 1
	}

	initialEndCapIndex := -1
	currentEndCap := ""
	if len(settings.Powerline.EndCaps) > 0 {
		currentEndCap = settings.Powerline.EndCaps[0]
	}
	for i, s := range endCapsList {
		if s.value == currentEndCap {
			initialEndCapIndex = i
			break
		}
	}
	if initialEndCapIndex == -1 {
		endCapsList = append(endCapsList, struct {
			name  string
			value string
		}{name: fmt.Sprintf("Custom (%s)", currentEndCap), value: currentEndCap})
		initialEndCapIndex = len(endCapsList) - 1
	}

	initialColorLevelIndex := 1 // default to ANSI 256
	for i, cl := range colorLevelsList {
		if cl.value == settings.ColorLevel {
			initialColorLevelIndex = i
			break
		}
	}

	return Model{
		settings:        settings,
		configPath:      configPath,
		activeMenu:      "main",
		cursor:          0,
		themeIndex:      initialThemeIndex,
		separatorIndex:  initialSeparatorIndex,
		startCapIndex:   initialStartCapIndex,
		endCapIndex:     initialEndCapIndex,
		colorLevelIndex: initialColorLevelIndex,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			if m.activeMenu == "main" {
				m.quitting = true
				return m, tea.Quit
			}
		}

		switch m.activeMenu {
		case "main":
			return m.updateMain(msg)
		case "lines":
			return m.updateLines(msg)
		case "items":
			return m.updateItems(msg)
		case "add_widget":
			return m.updateAddWidget(msg)
		case "select_theme":
			return m.updateSelectTheme(msg)
		case "select_separator":
			return m.updateSelectSeparator(msg)
		case "select_start_cap":
			return m.updateSelectStartCap(msg)
		case "select_end_cap":
			return m.updateSelectEndCap(msg)
		case "select_color_level":
			return m.updateSelectColorLevel(msg)
		}
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
		maxItems := 9
		if m.cursor < maxItems-1 {
			m.cursor++
		}

	case "enter", "\n":
		switch m.cursor {
		case 0: // Edit Lines
			m.activeMenu = "lines"
			m.cursor = 0
			m.moveMode = false

		case 1: // Toggle Powerline Mode
			m.settings.Powerline.Enabled = !m.settings.Powerline.Enabled

		case 2: // Select Powerline Theme
			m.activeMenu = "select_theme"
			m.cursor = m.themeIndex

		case 3: // Select Powerline Separator
			m.activeMenu = "select_separator"
			m.cursor = m.separatorIndex

		case 4: // Select Powerline Start Cap
			m.activeMenu = "select_start_cap"
			m.cursor = m.startCapIndex

		case 5: // Select Powerline End Cap
			m.activeMenu = "select_end_cap"
			m.cursor = m.endCapIndex

		case 6: // Select Color Level
			m.activeMenu = "select_color_level"
			m.cursor = m.colorLevelIndex

		case 7: // Save & Exit
			err := saveSettings(m.configPath, m.settings)
			if err == nil {
				m.saved = true
			}
			m.quitting = true
			return m, tea.Quit

		case 8: // Discard & Exit
			m.quitting = true
			return m, tea.Quit
		}
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
			m.activeMenu = "items"
			m.cursor = 0
			m.moveMode = false
		}

	case "esc":
		if m.moveMode {
			m.moveMode = false
		} else {
			m.activeMenu = "main"
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
		m.activeMenu = "add_widget"
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
			m.activeMenu = "lines"
			m.cursor = m.selectedLine
		}
	}
	return m, nil
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
		id := fmt.Sprintf("w_%d", time.Now().UnixNano())
		newWidget := types.WidgetItem{
			ID:    id,
			Type:  selectedType.wType,
			Color: selectedType.color,
		}
		if selectedType.customText != "" {
			newWidget.CustomText = selectedType.customText
		}
		if len(selectedType.metadata) > 0 {
			newWidget.Metadata = selectedType.metadata
		}

		widgets := m.settings.Lines[m.selectedLine]
		insertIndex := 0
		if len(widgets) > 0 {
			insertIndex = m.itemIndex + 1
		}

		newWidgets := make([]types.WidgetItem, len(widgets)+1)
		copy(newWidgets[:insertIndex], widgets[:insertIndex])
		newWidgets[insertIndex] = newWidget
		copy(newWidgets[insertIndex+1:], widgets[insertIndex:])

		m.settings.Lines[m.selectedLine] = newWidgets
		m.activeMenu = "items"
		m.cursor = insertIndex

	case "esc":
		m.activeMenu = "items"
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
		m.activeMenu = "main"
		m.cursor = 2

	case "esc":
		m.activeMenu = "main"
		m.cursor = 2
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
		if m.cursor < len(separatorsList)-1 {
			m.cursor++
		}

	case "enter", "\n":
		m.separatorIndex = m.cursor
		m.settings.Powerline.Separators = []string{separatorsList[m.separatorIndex].value}
		m.activeMenu = "main"
		m.cursor = 3

	case "esc":
		m.activeMenu = "main"
		m.cursor = 3
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
		if m.cursor < len(startCapsList)-1 {
			m.cursor++
		}

	case "enter", "\n":
		m.startCapIndex = m.cursor
		if startCapsList[m.startCapIndex].value == "" {
			m.settings.Powerline.StartCaps = []string{}
		} else {
			m.settings.Powerline.StartCaps = []string{startCapsList[m.startCapIndex].value}
		}
		m.activeMenu = "main"
		m.cursor = 4

	case "esc":
		m.activeMenu = "main"
		m.cursor = 4
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
		if m.cursor < len(endCapsList)-1 {
			m.cursor++
		}

	case "enter", "\n":
		m.endCapIndex = m.cursor
		if endCapsList[m.endCapIndex].value == "" {
			m.settings.Powerline.EndCaps = []string{}
		} else {
			m.settings.Powerline.EndCaps = []string{endCapsList[m.endCapIndex].value}
		}
		m.activeMenu = "main"
		m.cursor = 5

	case "esc":
		m.activeMenu = "main"
		m.cursor = 5
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
		m.settings.ColorLevel = colorLevelsList[m.colorLevelIndex].value
		m.activeMenu = "main"
		m.cursor = 6

	case "esc":
		m.activeMenu = "main"
		m.cursor = 6
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		if m.saved {
			return "\x1b[0mConfiguration saved successfully. Exiting...\n"
		}
		return "\x1b[0mChanges discarded. Exiting...\n"
	}

	var s stringsBuilder

	// Render Live Preview at the top (rendered on all screens except maybe add_widget if we want, but let's keep it consistent)
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244")).Render("--- Live Preview ---"))
	s.WriteString("\n")

	width := 160
	inputTokens := float64(14200)

	usedPct := float64(20.0)
	remainingPct := float64(80.0)
	g5hFraction := 0.5019274
	g5hReset := 8891.0
	p35hFraction := 1.0
	p35hReset := 17956.0
	gWkFraction := 0.9090967
	gWkReset := 567440.0
	p3WkFraction := 1.0
	p3WkReset := 604756.0
	sandboxEnabled := true
	previewCtx := types.RenderContext{
		TerminalWidth: &width,
		IsPreview:     true,
		Minimalist:    m.settings.MinimalistMode,
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID:          "gemini-3.5-flash-medium",
				DisplayName: "Gemini 3.5 Flash (Medium)",
			},
			ContextWindow: &types.ContextWindowInfo{
				TotalInputTokens:    &inputTokens,
				UsedPercentage:      &usedPct,
				RemainingPercentage: &remainingPct,
			},
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &g5hFraction,
					ResetInSeconds:    &g5hReset,
				},
				"gemini-weekly": {
					RemainingFraction: &gWkFraction,
					ResetInSeconds:    &gWkReset,
				},
				"3p-5h": {
					RemainingFraction: &p35hFraction,
					ResetInSeconds:    &p35hReset,
				},
				"3p-weekly": {
					RemainingFraction: &p3WkFraction,
					ResetInSeconds:    &p3WkReset,
				},
			},
			Sandbox: &types.SandboxInfo{
				Enabled: &sandboxEnabled,
			},
		},
	}

	previewSettings := m.settings
	if m.activeMenu == "select_theme" {
		if m.cursor >= 0 && m.cursor < len(themesList) {
			previewSettings.Powerline.Theme = themesList[m.cursor]
		}
	} else if m.activeMenu == "select_separator" {
		if m.cursor >= 0 && m.cursor < len(separatorsList) {
			previewSettings.Powerline.Separators = []string{separatorsList[m.cursor].value}
		}
	} else if m.activeMenu == "select_start_cap" {
		if m.cursor >= 0 && m.cursor < len(startCapsList) {
			if startCapsList[m.cursor].value == "" {
				previewSettings.Powerline.StartCaps = []string{}
			} else {
				previewSettings.Powerline.StartCaps = []string{startCapsList[m.cursor].value}
			}
		}
	} else if m.activeMenu == "select_end_cap" {
		if m.cursor >= 0 && m.cursor < len(endCapsList) {
			if endCapsList[m.cursor].value == "" {
				previewSettings.Powerline.EndCaps = []string{}
			} else {
				previewSettings.Powerline.EndCaps = []string{endCapsList[m.cursor].value}
			}
		}
	} else if m.activeMenu == "select_color_level" {
		if m.cursor >= 0 && m.cursor < len(colorLevelsList) {
			previewSettings.ColorLevel = colorLevelsList[m.cursor].value
		}
	} else if m.activeMenu == "add_widget" {
		if m.cursor >= 0 && m.cursor < len(widgetTypes) && m.selectedLine >= 0 && m.selectedLine < len(m.settings.Lines) {
			selectedType := widgetTypes[m.cursor]
			tempWidget := types.WidgetItem{
				ID:    "temp_preview_add",
				Type:  selectedType.wType,
				Color: selectedType.color,
			}
			if selectedType.customText != "" {
				tempWidget.CustomText = selectedType.customText
			}
			if len(selectedType.metadata) > 0 {
				tempWidget.Metadata = selectedType.metadata
			}

			widgets := m.settings.Lines[m.selectedLine]
			insertIndex := 0
			if len(widgets) > 0 {
				insertIndex = m.itemIndex + 1
			}

			newWidgets := make([]types.WidgetItem, len(widgets)+1)
			copy(newWidgets[:insertIndex], widgets[:insertIndex])
			newWidgets[insertIndex] = tempWidget
			copy(newWidgets[insertIndex+1:], widgets[insertIndex:])

			// We need to create a copy of Lines slice to prevent modifying the shared layout
			newLines := make([][]types.WidgetItem, len(previewSettings.Lines))
			for i, line := range previewSettings.Lines {
				if i == m.selectedLine {
					newLines[i] = newWidgets
				} else {
					newLines[i] = line
				}
			}
			previewSettings.Lines = newLines
		}
	}

	previewLines := renderer.RenderStatusLines(previewSettings, previewCtx)
	for _, line := range previewLines {
		s.WriteString("\x1b[0m" + line)
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// Render the active menu screen below the preview
	switch m.activeMenu {
	case "main":
		m.viewMain(&s)
	case "lines":
		m.viewLines(&s)
	case "items":
		m.viewItems(&s)
	case "add_widget":
		m.viewAddWidget(&s)
	case "select_theme":
		m.viewSelectTheme(&s)
	case "select_separator":
		m.viewSelectSeparator(&s)
	case "select_start_cap":
		m.viewSelectStartCap(&s)
	case "select_end_cap":
		m.viewSelectEndCap(&s)
	case "select_color_level":
		m.viewSelectColorLevel(&s)
	}

	return s.String()
}

func (m Model) viewMain(s *stringsBuilder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("agystatusline Configuration Menu"))
	s.WriteString("\n\n")

	menuItems := []string{
		"Edit Lines",
		fmt.Sprintf("Toggle Powerline Mode       [%t]", m.settings.Powerline.Enabled),
		fmt.Sprintf("Select Powerline Theme      [%s]", m.settings.Powerline.Theme),
		fmt.Sprintf("Select Powerline Separator  [%s]", separatorsList[m.separatorIndex].name),
		fmt.Sprintf("Select Powerline Start Cap  [%s]", startCapsList[m.startCapIndex].name),
		fmt.Sprintf("Select Powerline End Cap    [%s]", endCapsList[m.endCapIndex].name),
		fmt.Sprintf("Select Color Level          [%s]", colorLevelsList[m.colorLevelIndex].name),
		"Save & Exit",
		"Discard & Exit",
	}

	for i, item := range menuItems {
		cursorStr := " "
		style := lipgloss.NewStyle()
		if m.cursor == i {
			cursorStr = ">"
			style = style.Bold(true).Foreground(lipgloss.Color("226"))
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursorStr, style.Render(item)))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to toggle/select, q to quit)\n")
}

func (m Model) viewLines(s *stringsBuilder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Select Line to Edit Items"))
	s.WriteString("\n\n")

	for i, line := range m.settings.Lines {
		cursorStr := " "
		style := lipgloss.NewStyle()
		if m.cursor == i {
			if m.moveMode {
				cursorStr = "M"
				style = style.Bold(true).Foreground(lipgloss.Color("208"))
			} else {
				cursorStr = ">"
				style = style.Bold(true).Foreground(lipgloss.Color("226"))
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

		s.WriteString(fmt.Sprintf("%s Line %d: %s\n", cursorStr, i+1, style.Render(widgetStr)))
	}

	s.WriteString("\n(a: add line, d: delete line, m: toggle move mode, Enter: edit widgets, Esc: back)\n")
}

func (m Model) viewItems(s *stringsBuilder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render(fmt.Sprintf("Editing Line %d Items", m.selectedLine+1)))
	s.WriteString("\n\n")

	widgets := m.settings.Lines[m.selectedLine]
	if len(widgets) == 0 {
		s.WriteString("  (No widgets in this line)\n")
	} else {
		for i, w := range widgets {
			cursorStr := " "
			style := lipgloss.NewStyle()
			if m.cursor == i {
				if m.moveMode {
					cursorStr = "M"
					style = style.Bold(true).Foreground(lipgloss.Color("208"))
				} else {
					cursorStr = ">"
					style = style.Bold(true).Foreground(lipgloss.Color("226"))
				}
			}

			detailStr := fmt.Sprintf("[%d] %s", i+1, w.Type)
			if w.Color != "" {
				detailStr += fmt.Sprintf(" (color: %s)", w.Color)
			}
			s.WriteString(fmt.Sprintf("%s %s\n", cursorStr, style.Render(detailStr)))
		}
	}

	s.WriteString("\n(a: add widget, d: delete widget, m: toggle move mode, Esc: back)\n")
}

func (m Model) viewAddWidget(s *stringsBuilder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Select Widget Type to Add"))
	s.WriteString("\n\n")

	for i, t := range widgetTypes {
		cursorStr := " "
		style := lipgloss.NewStyle()
		if m.cursor == i {
			cursorStr = ">"
			style = style.Bold(true).Foreground(lipgloss.Color("226"))
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursorStr, style.Render(t.name)))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to add, Esc: cancel)\n")
}

func (m Model) viewSelectTheme(s *stringsBuilder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Select Powerline Theme"))
	s.WriteString("\n\n")

	for i, t := range themesList {
		cursorStr := " "
		style := lipgloss.NewStyle()
		if m.cursor == i {
			cursorStr = ">"
			style = style.Bold(true).Foreground(lipgloss.Color("226"))
		}
		themeStr := t
		if t == m.settings.Powerline.Theme {
			themeStr += " (active)"
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursorStr, style.Render(themeStr)))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to select, Esc: cancel)\n")
}

func (m Model) viewSelectSeparator(s *stringsBuilder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Select Powerline Separator"))
	s.WriteString("\n\n")

	for i, sep := range separatorsList {
		cursorStr := " "
		style := lipgloss.NewStyle()
		if m.cursor == i {
			cursorStr = ">"
			style = style.Bold(true).Foreground(lipgloss.Color("226"))
		}
		sepStr := sep.name
		currentSep := "\uE0B0"
		if len(m.settings.Powerline.Separators) > 0 {
			currentSep = m.settings.Powerline.Separators[0]
		}
		if sep.value == currentSep {
			sepStr += " (active)"
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursorStr, style.Render(sepStr)))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to select, Esc: cancel)\n")
}

func (m Model) viewSelectStartCap(s *stringsBuilder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Select Powerline Start Cap"))
	s.WriteString("\n\n")

	for i, capVal := range startCapsList {
		cursorStr := " "
		style := lipgloss.NewStyle()
		if m.cursor == i {
			cursorStr = ">"
			style = style.Bold(true).Foreground(lipgloss.Color("226"))
		}
		capStr := capVal.name
		currentCap := ""
		if len(m.settings.Powerline.StartCaps) > 0 {
			currentCap = m.settings.Powerline.StartCaps[0]
		}
		if capVal.value == currentCap {
			capStr += " (active)"
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursorStr, style.Render(capStr)))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to select, Esc: cancel)\n")
}

func (m Model) viewSelectEndCap(s *stringsBuilder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Select Powerline End Cap"))
	s.WriteString("\n\n")

	for i, capVal := range endCapsList {
		cursorStr := " "
		style := lipgloss.NewStyle()
		if m.cursor == i {
			cursorStr = ">"
			style = style.Bold(true).Foreground(lipgloss.Color("226"))
		}
		capStr := capVal.name
		currentCap := ""
		if len(m.settings.Powerline.EndCaps) > 0 {
			currentCap = m.settings.Powerline.EndCaps[0]
		}
		if capVal.value == currentCap {
			capStr += " (active)"
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursorStr, style.Render(capStr)))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to select, Esc: cancel)\n")
}

func (m Model) viewSelectColorLevel(s *stringsBuilder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Select Color Level"))
	s.WriteString("\n\n")

	for i, cl := range colorLevelsList {
		cursorStr := " "
		style := lipgloss.NewStyle()
		if m.cursor == i {
			cursorStr = ">"
			style = style.Bold(true).Foreground(lipgloss.Color("226"))
		}
		clStr := cl.name
		if cl.value == m.settings.ColorLevel {
			clStr += " (active)"
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursorStr, style.Render(clStr)))
	}

	s.WriteString("\n(Use arrows/jk to navigate, Enter to select, Esc: cancel)\n")
}

type stringsBuilder struct {
	str string
}

func (sb *stringsBuilder) WriteString(s string) {
	sb.str += s
}

func (sb *stringsBuilder) String() string {
	return sb.str
}

func saveSettings(path string, settings types.Settings) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	tempPath := fmt.Sprintf("%s.%d.%d.tmp", path, os.Getpid(), time.Now().UnixNano())
	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		os.Remove(tempPath)
	}()

	bytes, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(bytes)
	if err != nil {
		return err
	}
	file.Close()

	return os.Rename(tempPath, path)
}

// RunTUI launches the Bubble Tea program to edit settings interactively.
func RunTUI(settings types.Settings, configPath string) error {
	p := tea.NewProgram(NewModel(settings, configPath))
	_, err := p.Run()
	return err
}
