package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yuys13/agystatusline/renderer"
	"github.com/yuys13/agystatusline/types"
)

type Model struct {
	settings     types.Settings
	configPath   string
	activeMenu   string
	cursor       int
	quitting     bool
	saved        bool
	themeIndex   int
	selectedLine int
	itemIndex    int
	moveMode     bool
}

var themesList = []string{"nord", "nord-aurora", "monokai", "solarized", "minimal", "dracula", "catppuccin", "gruvbox", "onedark", "tokyonight"}

var widgetTypes = []struct {
	name            string
	wType           string
	color           string
	backgroundColor string
	customText      string
	metadata        map[string]string
}{
	{name: "Model", wType: "model", color: "cyan"},
	{name: "Context Length", wType: "context-length", color: "brightBlack"},
	{name: "Git Branch", wType: "git-branch", color: "magenta"},
	{name: "Git Changes", wType: "git-changes", color: "yellow"},
	{name: "Separator", wType: "separator"},
	{name: "Custom Text", wType: "custom-text", color: "white", customText: "Custom Text"},
	{name: "Context Used %", wType: "context-used-pct", color: "brightBlack"},
	{name: "Context Remaining %", wType: "context-remaining-pct", color: "brightBlack"},
	{name: "Quota: Gemini 5h", wType: "quota", color: "brightBlack", metadata: map[string]string{"key": "gemini-5h"}},
	{name: "Quota: Gemini Weekly", wType: "quota", color: "brightBlack", metadata: map[string]string{"key": "gemini-weekly"}},
	{name: "Quota: 3P 5h", wType: "quota", color: "brightBlack", metadata: map[string]string{"key": "3p-5h"}},
	{name: "Quota: 3P Weekly", wType: "quota", color: "brightBlack", metadata: map[string]string{"key": "3p-weekly"}},
}


func NewModel(settings types.Settings, configPath string) Model {
	initialThemeIndex := 0
	for i, t := range themesList {
		if t == settings.Powerline.Theme {
			initialThemeIndex = i
			break
		}
	}

	return Model{
		settings:   settings,
		configPath: configPath,
		activeMenu: "main",
		cursor:     0,
		themeIndex: initialThemeIndex,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
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
		maxItems := 5
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
			m.themeIndex = (m.themeIndex + 1) % len(themesList)
			m.settings.Powerline.Theme = themesList[m.themeIndex]

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

		var newWidgets []types.WidgetItem
		if insertIndex >= len(widgets) {
			newWidgets = append(widgets, newWidget)
		} else {
			newWidgets = append(widgets[:insertIndex], append([]types.WidgetItem{newWidget}, widgets[insertIndex:]...)...)
		}

		m.settings.Lines[m.selectedLine] = newWidgets
		m.activeMenu = "items"
		m.cursor = insertIndex

	case "esc":
		m.activeMenu = "items"
		m.cursor = m.itemIndex
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
		},
	}


	previewLines := renderer.RenderStatusLines(m.settings, previewCtx)
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
