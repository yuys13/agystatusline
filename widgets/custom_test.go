package widgets

import (
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestWidgetInterfaces(t *testing.T) {
	RegisterAll()
	ctx := types.RenderContext{
		Data: types.StatusJSON{},
	}
	item := types.WidgetItem{
		Type: "model",
	}

	widgetsList := []string{
		"model", "git-branch", "git-changes", "quota", "custom-text",
		"sandbox", "agent-state", "context-bar", "quota-bar", "artifacts",
		"subagents", "tasks",
	}

	for _, name := range widgetsList {
		t.Run(name, func(t *testing.T) {
			w := GetWidget(name)
			if w == nil {
				t.Fatalf("Widget %q not registered", name)
			}
			if nameStr := w.GetDisplayName(); nameStr == "" {
				t.Errorf("GetDisplayName() returned empty for %q", name)
			}
			if defaultColor := w.GetDefaultColor(); defaultColor == "" {
				t.Errorf("GetDefaultColor() returned empty for %q", name)
			}
			if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
				t.Errorf("GetBodyColor() returned empty for %q", name)
			}
		})
	}
}

func TestCustomTextWidget_Formatting(t *testing.T) {
	RegisterAll()
	settings := types.DefaultSettings()

	w := GetWidget("custom-text")
	if w == nil {
		t.Fatalf("custom-text widget not found")
	}

	if w.GetDefaultColor() != "white" {
		t.Errorf("Expected default color %q, got %q", "white", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Custom Text" {
		t.Errorf("Expected display name %q, got %q", "Custom Text", w.GetDisplayName())
	}

	customCases := []struct {
		name     string
		text     string
		expected string
	}{
		{"Empty text", "", ""},
		{"Unicode Japanese text", "ステータスライン", "ステータスライン"},
		{"ANSI styled text", "\x1b[31mRed Text\x1b[0m", "\x1b[31mRed Text\x1b[0m"},
		{"RFC 2606 URL", "https://example.com/api/v1", "https://example.com/api/v1"},
	}

	for _, tc := range customCases {
		t.Run(tc.name, func(t *testing.T) {
			item := types.WidgetItem{Type: "custom-text", Text: tc.text}
			title, body, err := w.Render(item, types.RenderContext{}, settings)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if title != "" {
				t.Errorf("Expected empty title, got %q", title)
			}
			if body != tc.expected {
				t.Errorf("Expected body %q, got %q", tc.expected, body)
			}
			if color := w.GetBodyColor(item, types.RenderContext{}); color != "white" {
				t.Errorf("Expected body color 'white', got %q", color)
			}
		})
	}
}
