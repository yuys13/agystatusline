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
	settings   types.Settings
	configPath string
	activeMenu string
	cursor     int
	quitting   bool
	saved      bool
	themeIndex int
}

var themesList = []string{"nord", "nord-aurora", "monokai", "solarized", "minimal", "dracula", "catppuccin", "gruvbox", "onedark", "tokyonight"}

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

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			maxItems := 5
			if m.cursor < maxItems-1 {
				m.cursor++
			}

		case "enter":
			switch m.cursor {
			case 0: // Edit Lines / Widgets (Placeholder message in View)
				// Interactive sub-editors can be expanded later.
				// Currently we toggle minimalistMode as a dummy action for lines.
				m.settings.MinimalistMode = !m.settings.MinimalistMode

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

	// Render Live Preview at the top
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244")).Render("--- Live Preview ---"))
	s.WriteString("\n")

	// Render preview statusline using Nord dummy data
	width := 80
	inputTokens := float64(14200)
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
				TotalInputTokens: &inputTokens,
			},
		},
	}

	previewLines := renderer.RenderStatusLines(m.settings, previewCtx)
	for _, line := range previewLines {
		s.WriteString("\x1b[0m" + line)
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// Render Configuration Menu below the preview
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("agystatusline Configuration Menu"))
	s.WriteString("\n\n")

	menuItems := []string{
		fmt.Sprintf("Toggle Minimalist Mode      [%t]", m.settings.MinimalistMode),
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
	return s.String()
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
