package widgets

import (
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestContextBarWidget_Normal(t *testing.T) {
	w := initTestWidget(t, "context-bar")
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

func TestContextBarWidget_EdgeCase_NilContextWindow(t *testing.T) {
	w := initTestWidget(t, "context-bar")
	settings := types.DefaultSettings()
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "context-bar"}

	title, body, err := w.Render(item, ctxNil, settings)
	if err != nil || title != "ctx" || body != "" {
		t.Errorf("Expected 'ctx' and empty body when ContextWindow is nil, got title=%q, body=%q, err=%v", title, body, err)
	}
}

func TestContextBarWidget_GetBodyColor_HighPercentage(t *testing.T) {
	w := initTestWidget(t, "context-bar")
	pctHigh := 95.0
	ctxHigh := types.RenderContext{
		Data: types.StatusJSON{
			ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pctHigh},
		},
	}
	item := types.WidgetItem{Type: "context-bar"}

	if w.GetBodyColor(item, ctxHigh) != "brightRed" {
		t.Errorf("Expected brightRed for 95%% context bar")
	}
}

func TestContextBarWidget_GetBodyColor_MidPercentage(t *testing.T) {
	w := initTestWidget(t, "context-bar")
	pctMid := 65.0
	ctxMid := types.RenderContext{
		Data: types.StatusJSON{
			ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pctMid},
		},
	}
	item := types.WidgetItem{Type: "context-bar"}

	if w.GetBodyColor(item, ctxMid) != "brightYellow" {
		t.Errorf("Expected brightYellow for 65%% context bar")
	}
}

func TestContextBarWidget_Interface(t *testing.T) {
	w := initTestWidget(t, "context-bar")
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "context-bar"}

	if nameStr := w.GetDisplayName(); nameStr == "" {
		t.Errorf("GetDisplayName() returned empty for context-bar")
	}
	if defaultColor := w.GetDefaultColor(); defaultColor == "" {
		t.Errorf("GetDefaultColor() returned empty for context-bar")
	}
	if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
		t.Errorf("GetBodyColor() returned empty for context-bar")
	}
}
