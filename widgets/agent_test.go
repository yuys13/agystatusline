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
		t.Errorf("Expected default color %q, got %q", "yellow", w.GetDefaultColor())
	}

	if w.GetDisplayName() != "Sandbox" {
		t.Errorf("Expected display name %q, got %q", "Sandbox", w.GetDisplayName())
	}

	settings := types.DefaultSettings()
	trueVal := true
	falseVal := false

	tests := []struct {
		name             string
		item             types.WidgetItem
		ctx              types.RenderContext
		expectedTitle    string
		expectedBody     string
		expectedColor    string
		bodyContainsText string
	}{
		{
			name:          "Nil sandbox info",
			item:          types.WidgetItem{Type: "sandbox"},
			ctx:           types.RenderContext{Data: types.StatusJSON{}},
			expectedTitle: "",
			expectedBody:  "",
		},
		{
			name:          "Sandbox.Enabled is nil",
			item:          types.WidgetItem{Type: "sandbox"},
			ctx:           types.RenderContext{Data: types.StatusJSON{Sandbox: &types.SandboxInfo{}}},
			expectedTitle: "",
			expectedBody:  "",
		},
		{
			name:          "Sandbox.Enabled is true (normal)",
			item:          types.WidgetItem{Type: "sandbox"},
			ctx:           types.RenderContext{Data: types.StatusJSON{Sandbox: &types.SandboxInfo{Enabled: &trueVal}}},
			expectedTitle: "sandbox",
			expectedBody:  "on",
		},
		{
			name:          "Sandbox.Enabled is true (raw value)",
			item:          types.WidgetItem{Type: "sandbox", RawValue: &trueVal},
			ctx:           types.RenderContext{Data: types.StatusJSON{Sandbox: &types.SandboxInfo{Enabled: &trueVal}}},
			expectedTitle: "",
			expectedBody:  "on",
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
			item:          types.WidgetItem{Type: "sandbox", RawValue: &trueVal},
			ctx:           types.RenderContext{Data: types.StatusJSON{Sandbox: &types.SandboxInfo{Enabled: &falseVal}}},
			expectedTitle: "",
			expectedBody:  "off",
		},
		{
			name:             "Sandbox.Enabled is false with preserve colors",
			item:             types.WidgetItem{Type: "sandbox", PreserveColors: &trueVal},
			ctx:              types.RenderContext{Data: types.StatusJSON{Sandbox: &types.SandboxInfo{Enabled: &falseVal}}},
			bodyContainsText: "sandbox off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if tt.bodyContainsText != "" {
				if !strings.Contains(body, tt.bodyContainsText) {
					t.Errorf("Expected body to contain %q, got %q", tt.bodyContainsText, body)
				}
			} else {
				if title != tt.expectedTitle || body != tt.expectedBody {
					t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
				}
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

	settings := types.DefaultSettings()
	count5 := 5

	tests := []struct {
		name          string
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
	}{
		{
			name:          "Nil artifact count",
			ctx:           types.RenderContext{Data: types.StatusJSON{}},
			expectedTitle: "artifacts",
			expectedBody:  "0",
		},
		{
			name:          "Valid artifact count",
			ctx:           types.RenderContext{Data: types.StatusJSON{ArtifactCount: &count5}},
			expectedTitle: "artifacts",
			expectedBody:  "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := types.WidgetItem{Type: "artifacts"}
			title, body, err := w.Render(item, tt.ctx, settings)
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

	settings := types.DefaultSettings()

	tests := []struct {
		name          string
		subagentsData any
		expectedTitle string
		expectedBody  string
	}{
		{
			name:          "Float64 number 3",
			subagentsData: float64(3),
			expectedTitle: "subagents",
			expectedBody:  "3",
		},
		{
			name:          "Float64 number 5",
			subagentsData: float64(5),
			expectedTitle: "subagents",
			expectedBody:  "5",
		},
		{
			name:          "Slice of agents",
			subagentsData: []any{"a1", "a2"},
			expectedTitle: "subagents",
			expectedBody:  "2",
		},
		{
			name:          "Invalid string type",
			subagentsData: "invalid_string_data",
			expectedTitle: "subagents",
			expectedBody:  "0",
		},
		{
			name:          "Nil data",
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
			item := types.WidgetItem{Type: "subagents"}
			title, body, err := w.Render(item, ctx, settings)
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

	settings := types.DefaultSettings()
	count2 := 2
	count3 := 3

	tests := []struct {
		name          string
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
	}{
		{
			name:          "Nil task count",
			ctx:           types.RenderContext{Data: types.StatusJSON{}},
			expectedTitle: "tasks",
			expectedBody:  "0",
		},
		{
			name:          "Task count 2",
			ctx:           types.RenderContext{Data: types.StatusJSON{TaskCount: &count2}},
			expectedTitle: "tasks",
			expectedBody:  "2",
		},
		{
			name:          "Task count 3",
			ctx:           types.RenderContext{Data: types.StatusJSON{TaskCount: &count3}},
			expectedTitle: "tasks",
			expectedBody:  "3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := types.WidgetItem{Type: "tasks"}
			title, body, err := w.Render(item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
		})
	}
}
