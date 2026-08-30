package widgets

import (
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
		t.Errorf("Expected default color %q, got %q", "yellow", w.GetDefaultColor())
	}

	if w.GetDisplayName() != "Sandbox" {
		t.Errorf("Expected display name %q, got %q", "Sandbox", w.GetDisplayName())
	}

	settings := types.DefaultSettings()
	trueVal := true
	falseVal := false

	tests := []struct {
		name          string
		item          types.WidgetItem
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
		expectedColor string
	}{
		{
			name:          "Nil sandbox info",
			item:          types.WidgetItem{Type: "sandbox"},
			ctx:           types.RenderContext{Data: types.StatusJSON{}},
			expectedTitle: "",
			expectedBody:  "",
			expectedColor: "brightBlack",
		},
		{
			name:          "Sandbox.Enabled is nil",
			item:          types.WidgetItem{Type: "sandbox"},
			ctx:           types.RenderContext{Data: types.StatusJSON{Sandbox: &types.SandboxInfo{}}},
			expectedTitle: "",
			expectedBody:  "",
			expectedColor: "brightBlack",
		},
		{
			name:          "Sandbox.Enabled is true (normal)",
			item:          types.WidgetItem{Type: "sandbox"},
			ctx:           types.RenderContext{Data: types.StatusJSON{Sandbox: &types.SandboxInfo{Enabled: &trueVal}}},
			expectedTitle: "sandbox",
			expectedBody:  "on",
			expectedColor: "brightGreen",
		},
		{
			name:          "Sandbox.Enabled is true (raw value)",
			item:          types.WidgetItem{Type: "sandbox", Raw: true},
			ctx:           types.RenderContext{Data: types.StatusJSON{Sandbox: &types.SandboxInfo{Enabled: &trueVal}}},
			expectedTitle: "",
			expectedBody:  "on",
			expectedColor: "brightGreen",
		},
		{
			name:          "Sandbox.Enabled is false (normal)",
			item:          types.WidgetItem{Type: "sandbox"},
			ctx:           types.RenderContext{Data: types.StatusJSON{Sandbox: &types.SandboxInfo{Enabled: &falseVal}}},
			expectedTitle: "sandbox",
			expectedBody:  "off",
			expectedColor: "brightBlack",
		},
		{
			name:          "Sandbox.Enabled is false (raw value)",
			item:          types.WidgetItem{Type: "sandbox", Raw: true},
			ctx:           types.RenderContext{Data: types.StatusJSON{Sandbox: &types.SandboxInfo{Enabled: &falseVal}}},
			expectedTitle: "",
			expectedBody:  "off",
			expectedColor: "brightBlack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
			if tt.expectedColor != "" {
				if color := w.GetBodyColor(tt.item, tt.ctx); color != tt.expectedColor {
					t.Errorf("Expected color %q, got %q", tt.expectedColor, color)
				}
			}
		})
	}
}

func TestAgentStateWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("agent-state")
	if w == nil {
		t.Fatalf("Agent state widget not found")
	}

	if w.GetDefaultColor() != "brightGreen" {
		t.Errorf("Expected default color %q, got %q", "brightGreen", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Agent State" {
		t.Errorf("Expected display name %q, got %q", "Agent State", w.GetDisplayName())
	}

	settings := types.DefaultSettings()

	tests := []struct {
		name          string
		state         string
		expectedTitle string
		expectedBody  string
		expectedColor string
	}{
		{
			name:          "Empty state (READY)",
			state:         "",
			expectedTitle: "",
			expectedBody:  "● READY",
			expectedColor: "brightGreen",
		},
		{
			name:          "Thinking state",
			state:         "thinking",
			expectedTitle: "",
			expectedBody:  "◆ THINKING",
			expectedColor: "brightYellow",
		},
		{
			name:          "Working state",
			state:         "working",
			expectedTitle: "",
			expectedBody:  "⚙ WORKING",
			expectedColor: "brightCyan",
		},
		{
			name:          "Tool use state",
			state:         "tool_use",
			expectedTitle: "",
			expectedBody:  "🔧 TOOL",
			expectedColor: "brightMagenta",
		},
		{
			name:          "Custom state",
			state:         "custom_state",
			expectedTitle: "",
			expectedBody:  "⏳ CUSTOM_STATE",
			expectedColor: "white",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := types.RenderContext{
				Data: types.StatusJSON{AgentState: tt.state},
			}
			item := types.WidgetItem{Type: "agent-state"}

			title, body, err := w.Render(item, ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
			if color := w.GetBodyColor(item, ctx); color != tt.expectedColor {
				t.Errorf("Expected color %q, got %q", tt.expectedColor, color)
			}
		})
	}
}

func TestArtifactsWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("artifacts")
	if w == nil {
		t.Fatalf("Artifacts widget not found")
	}

	if w.GetDefaultColor() != "brightWhite" {
		t.Errorf("Expected default color %q, got %q", "brightWhite", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Artifacts" {
		t.Errorf("Expected display name %q, got %q", "Artifacts", w.GetDisplayName())
	}

	settings := types.DefaultSettings()
	count5 := 5

	tests := []struct {
		name          string
		item          types.WidgetItem
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
	}{
		{
			name:          "Nil artifact count",
			item:          types.WidgetItem{Type: "artifacts"},
			ctx:           types.RenderContext{Data: types.StatusJSON{}},
			expectedTitle: "artifacts",
			expectedBody:  "0",
		},
		{
			name:          "Valid artifact count",
			item:          types.WidgetItem{Type: "artifacts"},
			ctx:           types.RenderContext{Data: types.StatusJSON{ArtifactCount: &count5}},
			expectedTitle: "artifacts",
			expectedBody:  "5",
		},
		{
			name:          "Raw artifact count",
			item:          types.WidgetItem{Type: "artifacts", Raw: true},
			ctx:           types.RenderContext{Data: types.StatusJSON{ArtifactCount: &count5}},
			expectedTitle: "",
			expectedBody:  "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
		})
	}
}

func TestSubagentsWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("subagents")
	if w == nil {
		t.Fatalf("Subagents widget not found")
	}

	if w.GetDefaultColor() != "brightWhite" {
		t.Errorf("Expected default color %q, got %q", "brightWhite", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Subagents" {
		t.Errorf("Expected display name %q, got %q", "Subagents", w.GetDisplayName())
	}

	settings := types.DefaultSettings()

	tests := []struct {
		name          string
		item          types.WidgetItem
		subagentsData any
		expectedTitle string
		expectedBody  string
	}{
		{
			name:          "Float64 number 3",
			item:          types.WidgetItem{Type: "subagents"},
			subagentsData: float64(3),
			expectedTitle: "subagents",
			expectedBody:  "3",
		},
		{
			name:          "Float64 number 5 raw",
			item:          types.WidgetItem{Type: "subagents", Raw: true},
			subagentsData: float64(5),
			expectedTitle: "",
			expectedBody:  "5",
		},
		{
			name:          "Slice of agents",
			item:          types.WidgetItem{Type: "subagents"},
			subagentsData: []any{"a1", "a2"},
			expectedTitle: "subagents",
			expectedBody:  "2",
		},
		{
			name:          "Int count 4",
			item:          types.WidgetItem{Type: "subagents"},
			subagentsData: 4,
			expectedTitle: "subagents",
			expectedBody:  "4",
		},
		{
			name:          "Invalid string type",
			item:          types.WidgetItem{Type: "subagents"},
			subagentsData: "invalid_string_data",
			expectedTitle: "subagents",
			expectedBody:  "0",
		},
		{
			name:          "Nil data",
			item:          types.WidgetItem{Type: "subagents"},
			subagentsData: nil,
			expectedTitle: "subagents",
			expectedBody:  "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := types.RenderContext{
				Data: types.StatusJSON{Subagents: tt.subagentsData},
			}
			title, body, err := w.Render(tt.item, ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
		})
	}
}

func TestTasksWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("tasks")
	if w == nil {
		t.Fatalf("Tasks widget not found")
	}

	if w.GetDefaultColor() != "brightWhite" {
		t.Errorf("Expected default color %q, got %q", "brightWhite", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Tasks" {
		t.Errorf("Expected display name %q, got %q", "Tasks", w.GetDisplayName())
	}

	settings := types.DefaultSettings()
	count2 := 2
	count3 := 3

	tests := []struct {
		name          string
		item          types.WidgetItem
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
	}{
		{
			name:          "Nil task count",
			item:          types.WidgetItem{Type: "tasks"},
			ctx:           types.RenderContext{Data: types.StatusJSON{}},
			expectedTitle: "tasks",
			expectedBody:  "0",
		},
		{
			name:          "Task count 2",
			item:          types.WidgetItem{Type: "tasks"},
			ctx:           types.RenderContext{Data: types.StatusJSON{TaskCount: &count2}},
			expectedTitle: "tasks",
			expectedBody:  "2",
		},
		{
			name:          "Task count 3 raw",
			item:          types.WidgetItem{Type: "tasks", Raw: true},
			ctx:           types.RenderContext{Data: types.StatusJSON{TaskCount: &count3}},
			expectedTitle: "",
			expectedBody:  "3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
		})
	}
}
