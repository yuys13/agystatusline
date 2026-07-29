package widgets

import (
	"github.com/yuys13/agystatusline/types"
	"testing"
)

func TestAgentStateWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("agent-state")
	if w == nil {
		t.Fatalf("Agent state widget not found")
	}

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

func TestAgentStateWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("agent-state")
	settings := types.DefaultSettings()

	states := []struct {
		state         string
		expectedColor string
		expectedText  string
	}{
		{"", "brightGreen", "● READY"},
		{"thinking", "brightYellow", "◆ THINKING"},
		{"working", "brightCyan", "⚙ WORKING"},
		{"tool_use", "brightMagenta", "🔧 TOOL"},
		{"custom_state", "white", "⏳ CUSTOM_STATE"},
	}

	for _, tc := range states {
		ctx := types.RenderContext{
			Data: types.StatusJSON{AgentState: tc.state},
		}
		item := types.WidgetItem{Type: "agent-state"}

		color := w.GetBodyColor(item, ctx)
		if color != tc.expectedColor {
			t.Errorf("For state %q expected color %s, got %s", tc.state, tc.expectedColor, color)
		}

		_, body, _ := w.Render(item, ctx, settings)
		if body != tc.expectedText {
			t.Errorf("For state %q expected text %q, got %q", tc.state, tc.expectedText, body)
		}
	}
}
