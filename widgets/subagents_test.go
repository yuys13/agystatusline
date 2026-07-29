package widgets

import (
	"github.com/yuys13/agystatusline/types"
	"testing"
)

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
