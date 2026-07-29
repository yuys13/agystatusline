package widgets

import (
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestArtifactsWidget_Normal(t *testing.T) {
	w := initTestWidget(t, "artifacts")
	settings := types.DefaultSettings()
	count := 5
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			ArtifactCount: &count,
		},
	}
	item := types.WidgetItem{Type: "artifacts"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "artifacts" || output != "5" {
		t.Errorf("Expected title 'artifacts' and body '5', got title '%s' and body '%s'", title, output)
	}
}

func TestArtifactsWidget_Interface(t *testing.T) {
	w := initTestWidget(t, "artifacts")
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "artifacts"}

	if nameStr := w.GetDisplayName(); nameStr == "" {
		t.Errorf("GetDisplayName() returned empty for artifacts")
	}
	if defaultColor := w.GetDefaultColor(); defaultColor == "" {
		t.Errorf("GetDefaultColor() returned empty for artifacts")
	}
	if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
		t.Errorf("GetBodyColor() returned empty for artifacts")
	}
}
