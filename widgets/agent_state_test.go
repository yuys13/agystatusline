package widgets

import (
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestAgentStateWidget_Normal(t *testing.T) {
	w := initTestWidget(t, "agent-state")
	settings := types.DefaultSettings()
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			AgentState: "thinking",
		},
	}
	item := types.WidgetItem{Type: "agent-state"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "◆ THINKING" {
		t.Errorf("Expected body '◆ THINKING', got title '%s' and body '%s'", title, output)
	}
}

func TestAgentStateWidget_StateEmpty_Ready(t *testing.T) {
	w := initTestWidget(t, "agent-state")
	settings := types.DefaultSettings()
	ctx := types.RenderContext{
		Data: types.StatusJSON{AgentState: ""},
	}
	item := types.WidgetItem{Type: "agent-state"}

	color := w.GetBodyColor(item, ctx)
	if color != "brightGreen" {
		t.Errorf("For state '' expected color brightGreen, got %s", color)
	}

	_, body, err := w.Render(item, ctx, settings)
	if err != nil || body != "● READY" {
		t.Errorf("For state '' expected text '● READY', got %q, err=%v", body, err)
	}
}

func TestAgentStateWidget_StateThinking(t *testing.T) {
	w := initTestWidget(t, "agent-state")
	settings := types.DefaultSettings()
	ctx := types.RenderContext{
		Data: types.StatusJSON{AgentState: "thinking"},
	}
	item := types.WidgetItem{Type: "agent-state"}

	color := w.GetBodyColor(item, ctx)
	if color != "brightYellow" {
		t.Errorf("For state 'thinking' expected color brightYellow, got %s", color)
	}

	_, body, err := w.Render(item, ctx, settings)
	if err != nil || body != "◆ THINKING" {
		t.Errorf("For state 'thinking' expected text '◆ THINKING', got %q, err=%v", body, err)
	}
}

func TestAgentStateWidget_StateWorking(t *testing.T) {
	w := initTestWidget(t, "agent-state")
	settings := types.DefaultSettings()
	ctx := types.RenderContext{
		Data: types.StatusJSON{AgentState: "working"},
	}
	item := types.WidgetItem{Type: "agent-state"}

	color := w.GetBodyColor(item, ctx)
	if color != "brightCyan" {
		t.Errorf("For state 'working' expected color brightCyan, got %s", color)
	}

	_, body, err := w.Render(item, ctx, settings)
	if err != nil || body != "⚙ WORKING" {
		t.Errorf("For state 'working' expected text '⚙ WORKING', got %q, err=%v", body, err)
	}
}

func TestAgentStateWidget_StateToolUse(t *testing.T) {
	w := initTestWidget(t, "agent-state")
	settings := types.DefaultSettings()
	ctx := types.RenderContext{
		Data: types.StatusJSON{AgentState: "tool_use"},
	}
	item := types.WidgetItem{Type: "agent-state"}

	color := w.GetBodyColor(item, ctx)
	if color != "brightMagenta" {
		t.Errorf("For state 'tool_use' expected color brightMagenta, got %s", color)
	}

	_, body, err := w.Render(item, ctx, settings)
	if err != nil || body != "🔧 TOOL" {
		t.Errorf("For state 'tool_use' expected text '🔧 TOOL', got %q, err=%v", body, err)
	}
}

func TestAgentStateWidget_CustomState(t *testing.T) {
	w := initTestWidget(t, "agent-state")
	settings := types.DefaultSettings()
	ctx := types.RenderContext{
		Data: types.StatusJSON{AgentState: "custom_state"},
	}
	item := types.WidgetItem{Type: "agent-state"}

	color := w.GetBodyColor(item, ctx)
	if color != "white" {
		t.Errorf("For state 'custom_state' expected color white, got %s", color)
	}

	_, body, err := w.Render(item, ctx, settings)
	if err != nil || body != "⏳ CUSTOM_STATE" {
		t.Errorf("For state 'custom_state' expected text '⏳ CUSTOM_STATE', got %q, err=%v", body, err)
	}
}

func TestAgentStateWidget_Interface(t *testing.T) {
	w := initTestWidget(t, "agent-state")
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "agent-state"}

	if nameStr := w.GetDisplayName(); nameStr == "" {
		t.Errorf("GetDisplayName() returned empty for agent-state")
	}
	if defaultColor := w.GetDefaultColor(); defaultColor == "" {
		t.Errorf("GetDefaultColor() returned empty for agent-state")
	}
	if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
		t.Errorf("GetBodyColor() returned empty for agent-state")
	}
}
