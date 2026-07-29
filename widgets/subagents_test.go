package widgets

import (
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestSubagentsWidget_Normal(t *testing.T) {
	w := initTestWidget(t, "subagents")
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

func TestSubagentsWidget_TypeSlice(t *testing.T) {
	w := initTestWidget(t, "subagents")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "subagents"}

	ctxSlice := types.RenderContext{
		Data: types.StatusJSON{Subagents: []any{"a1", "a2"}},
	}
	_, bodySlice, err := w.Render(item, ctxSlice, settings)
	if err != nil || bodySlice != "2" {
		t.Errorf("Expected '2' subagents, got %q, err=%v", bodySlice, err)
	}
}

func TestSubagentsWidget_TypeNumber(t *testing.T) {
	w := initTestWidget(t, "subagents")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "subagents"}

	ctxNum := types.RenderContext{
		Data: types.StatusJSON{Subagents: float64(5)},
	}
	_, bodyNum, err := w.Render(item, ctxNum, settings)
	if err != nil || bodyNum != "5" {
		t.Errorf("Expected '5' subagents, got %q, err=%v", bodyNum, err)
	}
}

func TestSubagentsWidget_InvalidType(t *testing.T) {
	w := initTestWidget(t, "subagents")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "subagents"}

	ctxString := types.RenderContext{
		Data: types.StatusJSON{
			Subagents: "invalid_string_data",
		},
	}
	title, body, err := w.Render(item, ctxString, settings)
	if err != nil || title != "subagents" || body != "0" {
		t.Errorf("Expected 'subagents' and '0' for string type Subagents, got title=%q, body=%q, err=%v", title, body, err)
	}
}

func TestSubagentsWidget_Interface(t *testing.T) {
	w := initTestWidget(t, "subagents")
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "subagents"}

	if nameStr := w.GetDisplayName(); nameStr == "" {
		t.Errorf("GetDisplayName() returned empty for subagents")
	}
	if defaultColor := w.GetDefaultColor(); defaultColor == "" {
		t.Errorf("GetDefaultColor() returned empty for subagents")
	}
	if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
		t.Errorf("GetBodyColor() returned empty for subagents")
	}
}
