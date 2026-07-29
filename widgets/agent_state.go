package widgets

import (
	"strings"

	"github.com/yuys13/agystatusline/types"
)

// AgentStateWidget displays the active agent state.
type AgentStateWidget struct{}

func (a *AgentStateWidget) GetDefaultColor() string { return "brightGreen" }
func (a *AgentStateWidget) GetDisplayName() string  { return "Agent State" }

func (a *AgentStateWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	state := ctx.Data.AgentState
	if state == "" {
		state = "idle"
	}
	switch state {
	case "idle":
		return "brightGreen"
	case "thinking":
		return "brightYellow"
	case "working":
		return "brightCyan"
	case "tool_use":
		return "brightMagenta"
	}
	return "white"
}

func (a *AgentStateWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	state := ctx.Data.AgentState
	if state == "" {
		state = "idle"
	}

	var symbolText string
	switch state {
	case "idle":
		symbolText = "● READY"
	case "thinking":
		symbolText = "◆ THINKING"
	case "working":
		symbolText = "⚙ WORKING"
	case "tool_use":
		symbolText = "🔧 TOOL"
	default:
		symbolText = "⏳ " + strings.ToUpper(state)
	}

	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		boldCode := "\x1b[1m"
		resetCode := "\x1b[22m\x1b[39m"
		var colorCode string
		switch state {
		case "idle":
			colorCode = "\x1b[92m"
		case "thinking":
			colorCode = "\x1b[93m"
		case "working":
			colorCode = "\x1b[96m"
		case "tool_use":
			colorCode = "\x1b[95m"
		default:
			colorCode = "\x1b[97m"
		}
		return "", boldCode + colorCode + symbolText + resetCode, nil
	}

	return "", symbolText, nil
}
