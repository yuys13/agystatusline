package widgets

import (
	"strconv"
	"strings"

	"github.com/yuys13/agystatusline/types"
)

// SandboxWidget displays the sandbox enabled status.
type SandboxWidget struct{}

func (s *SandboxWidget) GetDefaultColor() string { return "yellow" }
func (s *SandboxWidget) GetDisplayName() string  { return "Sandbox" }
func (s *SandboxWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	if ctx.Data.Sandbox != nil && ctx.Data.Sandbox.Enabled != nil && *ctx.Data.Sandbox.Enabled {
		return "brightGreen"
	}
	return "brightBlack"
}

func (s *SandboxWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	if ctx.Data.Sandbox == nil || ctx.Data.Sandbox.Enabled == nil {
		return "", "", nil
	}

	enabled := *ctx.Data.Sandbox.Enabled
	valStr := "off"
	if enabled {
		valStr = "on"
	}

	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		// statusline.sh style coloring:
		// sandbox is gray (ansi 90), ON is green and bold (ansi 92), off is gray (ansi 90)
		var bodyStr string
		if enabled {
			bodyStr = "\x1b[90msandbox\x1b[39m \x1b[92m\x1b[1mON\x1b[22m\x1b[39m"
		} else {
			bodyStr = "\x1b[90msandbox off\x1b[39m"
		}
		return "", bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", valStr, nil
	}

	if enabled {
		return "sandbox", "on", nil
	}
	return "sandbox", "off", nil
}

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

// ArtifactsWidget displays count of artifacts.
type ArtifactsWidget struct{}

func (a *ArtifactsWidget) GetDefaultColor() string { return "brightWhite" }
func (a *ArtifactsWidget) GetDisplayName() string  { return "Artifacts" }
func (a *ArtifactsWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightWhite"
}

func (a *ArtifactsWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	count := 0
	if ctx.Data.ArtifactCount != nil {
		count = *ctx.Data.ArtifactCount
	}

	countStr := strconv.Itoa(count)
	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		titleStr := "\x1b[90martifacts\x1b[39m"
		bodyStr := "\x1b[97m\x1b[1m" + countStr + "\x1b[22m\x1b[39m"
		return titleStr, bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", countStr, nil
	}
	return "artifacts", countStr, nil
}

// SubagentsWidget displays count of subagents.
type SubagentsWidget struct{}

func (s *SubagentsWidget) GetDefaultColor() string { return "brightWhite" }
func (s *SubagentsWidget) GetDisplayName() string  { return "Subagents" }
func (s *SubagentsWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightWhite"
}

func (s *SubagentsWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	count := 0
	if ctx.Data.Subagents != nil {
		if list, ok := ctx.Data.Subagents.([]any); ok {
			count = len(list)
		} else if num, ok := ctx.Data.Subagents.(float64); ok {
			count = int(num)
		}
	}

	countStr := strconv.Itoa(count)
	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		titleStr := "\x1b[90msubagents\x1b[39m"
		bodyStr := "\x1b[97m\x1b[1m" + countStr + "\x1b[22m\x1b[39m"
		return titleStr, bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", countStr, nil
	}
	return "subagents", countStr, nil
}

// TasksWidget displays count of background tasks.
type TasksWidget struct{}

func (t *TasksWidget) GetDefaultColor() string { return "brightWhite" }
func (t *TasksWidget) GetDisplayName() string  { return "Tasks" }
func (t *TasksWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightWhite"
}

func (t *TasksWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	count := 0
	if ctx.Data.TaskCount != nil {
		count = *ctx.Data.TaskCount
	}

	countStr := strconv.Itoa(count)
	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		titleStr := "\x1b[90mtasks\x1b[39m"
		bodyStr := "\x1b[97m\x1b[1m" + countStr + "\x1b[22m\x1b[39m"
		return titleStr, bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", countStr, nil
	}
	return "tasks", countStr, nil
}
