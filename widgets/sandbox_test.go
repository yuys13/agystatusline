package widgets

import (
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestSandboxWidget_DefaultColorAndDisplayName(t *testing.T) {
	w := initTestWidget(t, "sandbox")

	if w.GetDefaultColor() != "yellow" {
		t.Errorf("Expected default color 'yellow', got '%s'", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Sandbox" {
		t.Errorf("Expected display name 'Sandbox', got '%s'", w.GetDisplayName())
	}
}

func TestSandboxWidget_NilSandboxInfo(t *testing.T) {
	w := initTestWidget(t, "sandbox")
	settings := types.DefaultSettings()
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "sandbox"}

	titleNil, outNil, err := w.Render(item, ctxNil, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleNil != "" || outNil != "" {
		t.Errorf("Expected empty title/body when sandbox is nil, got title '%s' and body '%s'", titleNil, outNil)
	}
}

func TestSandboxWidget_NilSandboxEnabled(t *testing.T) {
	w := initTestWidget(t, "sandbox")
	settings := types.DefaultSettings()
	ctxNilEnabled := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{},
		},
	}
	item := types.WidgetItem{Type: "sandbox"}

	titleNilEnabled, outNilEnabled, err := w.Render(item, ctxNilEnabled, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleNilEnabled != "" || outNilEnabled != "" {
		t.Errorf("Expected empty title/body when sandbox.enabled is nil, got title '%s' and body '%s'", titleNilEnabled, outNilEnabled)
	}
}

func TestSandboxWidget_EnabledTrue_Normal(t *testing.T) {
	w := initTestWidget(t, "sandbox")
	settings := types.DefaultSettings()
	trueVal := true
	ctxTrue := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{
				Enabled: &trueVal,
			},
		},
	}
	item := types.WidgetItem{Type: "sandbox"}

	titleTrue, outTrue, err := w.Render(item, ctxTrue, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleTrue != "sandbox" || outTrue != "on" {
		t.Errorf("Expected title 'sandbox' and body 'on', got title '%s' and body '%s'", titleTrue, outTrue)
	}
}

func TestSandboxWidget_EnabledTrue_RawValue(t *testing.T) {
	w := initTestWidget(t, "sandbox")
	settings := types.DefaultSettings()
	trueVal := true
	ctxTrue := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{
				Enabled: &trueVal,
			},
		},
	}
	itemRaw := types.WidgetItem{Type: "sandbox", RawValue: &trueVal}

	titleTrueRaw, outTrueRaw, err := w.Render(itemRaw, ctxTrue, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleTrueRaw != "" || outTrueRaw != "on" {
		t.Errorf("Expected title '' and body 'on', got title '%s' and body '%s'", titleTrueRaw, outTrueRaw)
	}
}

func TestSandboxWidget_EnabledFalse_Normal(t *testing.T) {
	w := initTestWidget(t, "sandbox")
	settings := types.DefaultSettings()
	falseVal := false
	ctxFalse := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{
				Enabled: &falseVal,
			},
		},
	}
	item := types.WidgetItem{Type: "sandbox"}

	titleFalse, outFalse, err := w.Render(item, ctxFalse, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleFalse != "sandbox" || outFalse != "off" {
		t.Errorf("Expected title 'sandbox' and body 'off', got title '%s' and body '%s'", titleFalse, outFalse)
	}
}

func TestSandboxWidget_EnabledFalse_RawValue(t *testing.T) {
	w := initTestWidget(t, "sandbox")
	settings := types.DefaultSettings()
	falseVal := false
	trueVal := true
	ctxFalse := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{
				Enabled: &falseVal,
			},
		},
	}
	itemRaw := types.WidgetItem{Type: "sandbox", RawValue: &trueVal}

	titleFalseRaw, outFalseRaw, err := w.Render(itemRaw, ctxFalse, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleFalseRaw != "" || outFalseRaw != "off" {
		t.Errorf("Expected title '' and body 'off', got title '%s' and body '%s'", titleFalseRaw, outFalseRaw)
	}
}

func TestSandboxWidget_EdgeCase_NilSandbox(t *testing.T) {
	w := initTestWidget(t, "sandbox")
	settings := types.DefaultSettings()
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "sandbox"}

	_, bodyNil, err := w.Render(item, ctxNil, settings)
	if err != nil || bodyNil != "" {
		t.Errorf("Expected empty body when sandbox is nil, got %q, err=%v", bodyNil, err)
	}
}

func TestSandboxWidget_EdgeCase_SandboxOff(t *testing.T) {
	w := initTestWidget(t, "sandbox")
	settings := types.DefaultSettings()
	falseVal := false
	ctxOff := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{Enabled: &falseVal},
		},
	}
	item := types.WidgetItem{Type: "sandbox"}

	titleOff, bodyOff, err := w.Render(item, ctxOff, settings)
	if err != nil || titleOff != "sandbox" || bodyOff != "off" {
		t.Errorf("Expected 'sandbox' and 'off', got %q, %q, err=%v", titleOff, bodyOff, err)
	}
	if w.GetBodyColor(item, ctxOff) != "brightBlack" {
		t.Errorf("Expected brightBlack for sandbox off")
	}
}

func TestSandboxWidget_EdgeCase_SandboxOff_PreserveColors(t *testing.T) {
	w := initTestWidget(t, "sandbox")
	settings := types.DefaultSettings()
	falseVal := false
	ctxOff := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{Enabled: &falseVal},
		},
	}
	trueVal := true
	itemPreserve := types.WidgetItem{Type: "sandbox", PreserveColors: &trueVal}

	_, bodyPreserve, err := w.Render(itemPreserve, ctxOff, settings)
	if err != nil || !strings.Contains(bodyPreserve, "sandbox off") {
		t.Errorf("Expected preserve colors text to contain 'sandbox off', got %q, err=%v", bodyPreserve, err)
	}
}

func TestSandboxWidget_Interface(t *testing.T) {
	w := initTestWidget(t, "sandbox")
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "sandbox"}

	if nameStr := w.GetDisplayName(); nameStr == "" {
		t.Errorf("GetDisplayName() returned empty for sandbox")
	}
	if defaultColor := w.GetDefaultColor(); defaultColor == "" {
		t.Errorf("GetDefaultColor() returned empty for sandbox")
	}
	if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
		t.Errorf("GetBodyColor() returned empty for sandbox")
	}
}
