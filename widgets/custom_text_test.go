package widgets

import (
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestCustomTextWidget_Interface(t *testing.T) {
	w := initTestWidget(t, "custom-text")
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "custom-text"}

	if nameStr := w.GetDisplayName(); nameStr == "" {
		t.Errorf("GetDisplayName() returned empty for custom-text")
	}
	if defaultColor := w.GetDefaultColor(); defaultColor == "" {
		t.Errorf("GetDefaultColor() returned empty for custom-text")
	}
	if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
		t.Errorf("GetBodyColor() returned empty for custom-text")
	}
}

func TestCustomTextWidget_Render(t *testing.T) {
	w := initTestWidget(t, "custom-text")
	settings := types.DefaultSettings()
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{
		Type:       "custom-text",
		CustomText: "Hello World",
	}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "Hello World" {
		t.Errorf("Expected title '' and body 'Hello World', got title '%s' and body '%s'", title, output)
	}
}
