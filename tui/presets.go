package tui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/yuys13/agystatusline/types"
)

var themesList = []string{"nord", "nord-aurora", "monokai", "solarized", "minimal", "dracula", "catppuccin", "gruvbox", "onedark", "tokyonight"}

var colorLevelsList = []struct {
	name  string
	value int
}{
	{name: "ANSI 16 Colors (Default)", value: 1},
	{name: "ANSI 256 Colors", value: 2},
	{name: "Truecolor (24-bit)", value: 3},
}

type choiceItem struct {
	name  string
	value string
}

var defaultSeparators = []choiceItem{
	{name: "None", value: ""},
	{name: "Arrow (\uE0B0)", value: "\uE0B0"},
	{name: "Round (\uE0B4)", value: "\uE0B4"},
	{name: "Flame (\uE0C0)", value: "\uE0C0"},
	{name: "Hexagon (\uE0C6)", value: "\uE0C6"},
	{name: "Slanted (\uE0C8)", value: "\uE0C8"},
}

var defaultStartCaps = []choiceItem{
	{name: "None", value: ""},
	{name: "Triangle (\uE0B2)", value: "\uE0B2"},
	{name: "Round (\uE0B6)", value: "\uE0B6"},
	{name: "Lower Triangle (\uE0BA)", value: "\uE0BA"},
	{name: "Diagonal (\uE0BE)", value: "\uE0BE"},
	{name: "Flame (\uE0C2)", value: "\uE0C2 "},
	{name: "Hexagon (\uE0C7)", value: "\uE0C7 "},
	{name: "Slanted (\uE0CA)", value: "\uE0CA "},
}

var defaultEndCaps = []choiceItem{
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
	name string
	item types.WidgetItem
}{
	// Default Settings
	{name: "Agent State", item: types.WidgetItem{Type: "agent-state"}},
	{name: "Model", item: types.WidgetItem{Type: "model"}},
	{name: "Context Bar", item: types.WidgetItem{Type: "context-bar"}},
	{name: "Artifacts", item: types.WidgetItem{Type: "artifacts"}},
	{name: "Subagents", item: types.WidgetItem{Type: "subagents"}},
	{name: "Tasks", item: types.WidgetItem{Type: "tasks"}},
	{name: "Sandbox Enabled", item: types.WidgetItem{Type: "sandbox"}},

	// Git-related
	{name: "Git Repository", item: types.WidgetItem{Type: "git-repo"}},
	{name: "Git Branch", item: types.WidgetItem{Type: "git-branch"}},
	{name: "Git Ahead/Behind", item: types.WidgetItem{Type: "git-ahead-behind"}},
	{name: "Git Changes", item: types.WidgetItem{Type: "git-changes"}},

	// Quota
	{name: "Quota: All", item: types.WidgetItem{Type: "quota-all"}},
	{name: "Quota: Gemini 5h", item: types.WidgetItem{Type: "quota-5h"}},
	{name: "Quota: Gemini 7d", item: types.WidgetItem{Type: "quota-7d"}},
	{name: "Quota: Argon 7d", item: types.WidgetItem{Type: "quota-argon-7d"}},
	{name: "Quota: Claude 5h", item: types.WidgetItem{Type: "quota-3p-5h"}},
	{name: "Quota: Claude 7d", item: types.WidgetItem{Type: "quota-3p-7d"}},

	// Quota Bar
	{name: "Quota Bar: All", item: types.WidgetItem{Type: "quota-bar-all"}},
	{name: "Quota Bar: Gemini 5h", item: types.WidgetItem{Type: "quota-bar-5h"}},
	{name: "Quota Bar: Gemini 7d", item: types.WidgetItem{Type: "quota-bar-7d"}},
	{name: "Quota Bar: Argon 7d", item: types.WidgetItem{Type: "quota-bar-argon-7d"}},
	{name: "Quota Bar: Claude 5h", item: types.WidgetItem{Type: "quota-bar-3p-5h"}},
	{name: "Quota Bar: Claude 7d", item: types.WidgetItem{Type: "quota-bar-3p-7d"}},

	// Custom Text
	{name: "Custom Text", item: types.WidgetItem{Type: "custom-text", Text: "CUSTOM"}},
}

// initChoiceList clones defaults, locates current (optionally ignoring trailing spaces),
// and appends a Custom entry when current is not present in defaults.
func initChoiceList(defaults []choiceItem, current string, trimRightSpace bool) ([]choiceItem, int) {
	items := slices.Clone(defaults)
	target := current
	if trimRightSpace {
		target = strings.TrimRight(target, " ")
	}
	for i, item := range items {
		val := item.value
		if trimRightSpace {
			val = strings.TrimRight(val, " ")
		}
		if val == target {
			return items, i
		}
	}
	items = append(items, choiceItem{name: fmt.Sprintf("Custom (%s)", current), value: current})
	return items, len(items) - 1
}
