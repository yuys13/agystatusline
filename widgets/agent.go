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

	if item.Raw {
		return "", valStr, nil
	}

	return "sandbox", valStr, nil
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
	if item.Raw {
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
		switch v := ctx.Data.Subagents.(type) {
		case []any:
			count = len(v)
		case float64:
			count = int(v)
		case int:
			count = v
		}
	}

	countStr := strconv.Itoa(count)
	if item.Raw {
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
	if item.Raw {
		return "", countStr, nil
	}
	return "tasks", countStr, nil
}
