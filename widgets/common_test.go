package widgets

import (
	"github.com/yuys13/agystatusline/types"
	"testing"
)

func TestWidgetInterfaces(t *testing.T) {
	RegisterAll()
	ctx := types.RenderContext{
		Data: types.StatusJSON{},
	}
	item := types.WidgetItem{
		Type: "model",
	}

	for _, name := range []string{
		"model", "git-branch", "git-changes", "quota", "custom-text",
		"sandbox", "agent-state", "context-bar", "quota-bar", "artifacts",
		"subagents", "tasks",
	} {
		w := GetWidget(name)
		if w == nil {
			t.Fatalf("Widget %s not registered", name)
		}
		if nameStr := w.GetDisplayName(); nameStr == "" {
			t.Errorf("GetDisplayName() returned empty for %s", name)
		}
		if defaultColor := w.GetDefaultColor(); defaultColor == "" {
			t.Errorf("GetDefaultColor() returned empty for %s", name)
		}
		if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
			t.Errorf("GetBodyColor() returned empty for %s", name)
		}
	}
}
