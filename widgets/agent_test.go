package widgets

import (
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestSandboxWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("sandbox")
	if w == nil {
		t.Fatalf("Sandbox widget not found in registry")
	}

	if w.GetDefaultColor() != "yellow" {
		t.Errorf("Expected default color 'yellow', got '%s'", w.GetDefaultColor())
	}

	if w.GetDisplayName() != "Sandbox" {
		t.Errorf("Expected display name 'Sandbox', got '%s'", w.GetDisplayName())
	}

	settings := types.DefaultSettings()

	// Case 1: Sandbox info is nil
	ctxNil := types.RenderContext{
		Data: types.StatusJSON{},
	}
	item := types.WidgetItem{Type: "sandbox"}
	titleNil, outNil, err := w.Render(item, ctxNil, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleNil != "" || outNil != "" {
		t.Errorf("Expected empty title/body when sandbox is nil, got title '%s' and body '%s'", titleNil, outNil)
	}

	// Case 2: Sandbox.Enabled is nil
	ctxNilEnabled := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{},
		},
	}
	titleNilEnabled, outNilEnabled, err := w.Render(item, ctxNilEnabled, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleNilEnabled != "" || outNilEnabled != "" {
		t.Errorf("Expected empty title/body when sandbox.enabled is nil, got title '%s' and body '%s'", titleNilEnabled, outNilEnabled)
	}

	// Case 3: Sandbox.Enabled is true (normal and raw)
	trueVal := true
	ctxTrue := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{
				Enabled: &trueVal,
			},
		},
	}
	titleTrue, outTrue, err := w.Render(item, ctxTrue, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleTrue != "sandbox" || outTrue != "on" {
		t.Errorf("Expected title 'sandbox' and body 'on', got title '%s' and body '%s'", titleTrue, outTrue)
	}

	itemRaw := types.WidgetItem{Type: "sandbox", RawValue: &trueVal}
	titleTrueRaw, outTrueRaw, err := w.Render(itemRaw, ctxTrue, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleTrueRaw != "" || outTrueRaw != "on" {
		t.Errorf("Expected title '' and body 'on', got title '%s' and body '%s'", titleTrueRaw, outTrueRaw)
	}

	// Case 4: Sandbox.Enabled is false (normal and raw)
	falseVal := false
	ctxFalse := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{
				Enabled: &falseVal,
			},
		},
	}
	titleFalse, outFalse, err := w.Render(item, ctxFalse, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleFalse != "sandbox" || outFalse != "off" {
		t.Errorf("Expected title 'sandbox' and body 'off', got title '%s' and body '%s'", titleFalse, outFalse)
	}

	titleFalseRaw, outFalseRaw, err := w.Render(itemRaw, ctxFalse, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleFalseRaw != "" || outFalseRaw != "off" {
		t.Errorf("Expected title '' and body 'off', got title '%s' and body '%s'", titleFalseRaw, outFalseRaw)
	}
}

func TestSandboxWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("sandbox")
	settings := types.DefaultSettings()

	// Sandbox nil
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "sandbox"}
	_, bodyNil, _ := w.Render(item, ctxNil, settings)
	if bodyNil != "" {
		t.Errorf("Expected empty body when sandbox is nil")
	}

	// Sandbox off
	falseVal := false
	ctxOff := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{Enabled: &falseVal},
		},
	}
	titleOff, bodyOff, _ := w.Render(item, ctxOff, settings)
	if titleOff != "sandbox" || bodyOff != "off" {
		t.Errorf("Expected 'sandbox' and 'off', got %q, %q", titleOff, bodyOff)
	}
	if w.GetBodyColor(item, ctxOff) != "brightBlack" {
		t.Errorf("Expected brightBlack for sandbox off")
	}

	// Preserve colors
	trueVal := true
	itemPreserve := types.WidgetItem{Type: "sandbox", PreserveColors: &trueVal}
	_, bodyPreserve, _ := w.Render(itemPreserve, ctxOff, settings)
	if !strings.Contains(bodyPreserve, "sandbox off") {
		t.Errorf("Expected preserve colors text to contain 'sandbox off'")
	}
}

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

func TestArtifactsWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("artifacts")
	if w == nil {
		t.Fatalf("Artifacts widget not found")
	}

	settings := types.DefaultSettings()
	count := 5
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			ArtifactCount: &count,
		},
	}
	item := types.WidgetItem{Type: "artifacts"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "artifacts" || output != "5" {
		t.Errorf("Expected title 'artifacts' and body '5', got title '%s' and body '%s'", title, output)
	}
}

func TestSubagentsWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("subagents")
	if w == nil {
		t.Fatalf("Subagents widget not found")
	}

	settings := types.DefaultSettings()
	count := 3
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Subagents: float64(count),
		},
	}
	item := types.WidgetItem{Type: "subagents"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "subagents" || output != "3" {
		t.Errorf("Expected title 'subagents' and body '3', got title '%s' and body '%s'", title, output)
	}
}

func TestSubagentsWidget_Types(t *testing.T) {
	RegisterAll()
	w := GetWidget("subagents")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "subagents"}

	// Array type
	ctxSlice := types.RenderContext{
		Data: types.StatusJSON{Subagents: []any{"a1", "a2"}},
	}
	_, bodySlice, _ := w.Render(item, ctxSlice, settings)
	if bodySlice != "2" {
		t.Errorf("Expected '2' subagents, got %q", bodySlice)
	}

	// Number type
	ctxNum := types.RenderContext{
		Data: types.StatusJSON{Subagents: float64(5)},
	}
	_, bodyNum, _ := w.Render(item, ctxNum, settings)
	if bodyNum != "5" {
		t.Errorf("Expected '5' subagents, got %q", bodyNum)
	}
}

func TestSubagentsWidget_InvalidType(t *testing.T) {
	RegisterAll()
	w := GetWidget("subagents")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "subagents"}

	// Subagents as unexpected string type
	ctxString := types.RenderContext{
		Data: types.StatusJSON{
			Subagents: "invalid_string_data",
		},
	}
	title, body, err := w.Render(item, ctxString, settings)
	if err != nil || title != "subagents" || body != "0" {
		t.Errorf("Expected 'subagents' and '0' for string type Subagents, got title=%q, body=%q", title, body)
	}
}

func TestTasksWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("tasks")
	if w == nil {
		t.Fatalf("Tasks widget not found")
	}

	settings := types.DefaultSettings()
	count := 2
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			TaskCount: &count,
		},
	}
	item := types.WidgetItem{Type: "tasks"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "tasks" || output != "2" {
		t.Errorf("Expected title 'tasks' and body '2', got title '%s' and body '%s'", title, output)
	}
}

func TestTasksWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("tasks")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "tasks"}

	// Nil task count
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	title, body, _ := w.Render(item, ctxNil, settings)
	if title != "tasks" || body != "0" {
		t.Errorf("Expected 'tasks' and '0' for nil TaskCount, got %q, %q", title, body)
	}

	// Valid task count
	taskCount := 3
	ctxVal := types.RenderContext{Data: types.StatusJSON{TaskCount: &taskCount}}
	_, bodyVal, _ := w.Render(item, ctxVal, settings)
	if bodyVal != "3" {
		t.Errorf("Expected '3' for TaskCount=3, got %q", bodyVal)
	}
}
