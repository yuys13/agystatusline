package widgets

import (
	"github.com/yuys13/agystatusline/types"
	"strings"
	"testing"
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
