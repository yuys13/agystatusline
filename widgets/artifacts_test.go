package widgets

import (
	"github.com/yuys13/agystatusline/types"
	"testing"
)

func TestArtifactsWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("artifacts")
	if w == nil {
		t.Fatalf("Artifacts widget not found")
	}

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
