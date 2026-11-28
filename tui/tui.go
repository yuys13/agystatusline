package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yuys13/agystatusline/types"
)

const (
	menuMain             = "main"
	menuCorruptedConfig  = "corrupted_config"
	menuPowerline        = "powerline"
	menuLines            = "lines"
	menuItems            = "items"
	menuAddWidget        = "add_widget"
	menuSelectTheme      = "select_theme"
	menuSelectSeparator  = "select_separator"
	menuSelectStartCap   = "select_start_cap"
	menuSelectEndCap     = "select_end_cap"
	menuSelectColorLevel = "select_color_level"
)

type Model struct {
	settings        types.Settings
	configPath      string
	loadError       string
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
	separators      []choiceItem
	startCaps       []choiceItem
	endCaps         []choiceItem
}

func NewModel(settings types.Settings, configPath string, loadError ...string) Model {
	settings.ApplyDefaults()
	initialThemeIndex := 0
	for i, t := range themesList {
		if t == settings.Powerline.Theme {
			initialThemeIndex = i
			break
		}
	}

	separators, initialSeparatorIndex := initChoiceList(defaultSeparators, settings.Powerline.Separator, true)
	startCaps, initialStartCapIndex := initChoiceList(defaultStartCaps, settings.Powerline.StartCaps, false)
	endCaps, initialEndCapIndex := initChoiceList(defaultEndCaps, settings.Powerline.EndCaps, false)

	initialColorLevelIndex := 0 // default to ANSI 16
	for i, cl := range colorLevelsList {
		if cl.value == settings.General.ColorLevel {
			initialColorLevelIndex = i
			break
		}
	}

	initialMenu := menuMain
	var errMsg string
	if len(loadError) > 0 && loadError[0] != "" {
		errMsg = loadError[0]
		initialMenu = menuCorruptedConfig
	}

	return Model{
		settings:        settings,
		configPath:      configPath,
		loadError:       errMsg,
		activeMenu:      initialMenu,
		cursor:          0,
		themeIndex:      initialThemeIndex,
		separatorIndex:  initialSeparatorIndex,
		startCapIndex:   initialStartCapIndex,
		endCapIndex:     initialEndCapIndex,
		colorLevelIndex: initialColorLevelIndex,
		separators:      separators,
		startCaps:       startCaps,
		endCaps:         endCaps,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// RunTUI launches the Bubble Tea program to edit settings interactively.
func RunTUI(settings types.Settings, configPath string, loadError ...string) error {
	p := tea.NewProgram(NewModel(settings, configPath, loadError...))
	_, err := p.Run()
	return err
}
