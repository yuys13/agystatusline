package widgets

import (
	"github.com/yuys13/agystatusline/types"
	"strings"
	"testing"
)

func TestContextBarWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("context-bar")
	if w == nil {
		t.Fatalf("Context bar widget not found")
	}

	settings := types.DefaultSettings()
	pct := 50.0
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			ContextWindow: &types.ContextWindowInfo{
				UsedPercentage: &pct,
			},
		},
	}
	item := types.WidgetItem{Type: "context-bar"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "ctx" || !strings.Contains(output, "50.0%") {
		t.Errorf("Expected title 'ctx' and body containing '50.0%%', got title '%s' and body '%s'", title, output)
	}
}

func TestContextBarWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("context-bar")
	settings := types.DefaultSettings()

	// Nil context window
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "context-bar"}
	title, body, _ := w.Render(item, ctxNil, settings)
	if title != "ctx" || body != "" {
		t.Errorf("Expected 'ctx' and empty body when ContextWindow is nil, got %q, %q", title, body)
	}

	// High percentage colors
	pctHigh := 95.0
	ctxHigh := types.RenderContext{
		Data: types.StatusJSON{
			ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pctHigh},
		},
	}
	if w.GetBodyColor(item, ctxHigh) != "brightRed" {
		t.Errorf("Expected brightRed for 95%% context bar")
	}

	pctMid := 65.0
	ctxMid := types.RenderContext{
		Data: types.StatusJSON{
			ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pctMid},
		},
	}
	if w.GetBodyColor(item, ctxMid) != "brightYellow" {
		t.Errorf("Expected brightYellow for 65%% context bar")
	}
}
