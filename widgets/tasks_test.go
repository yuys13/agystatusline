package widgets

import (
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestTasksWidget_Normal(t *testing.T) {
	w := initTestWidget(t, "tasks")
	settings := types.DefaultSettings()
	count := 2
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			TaskCount: &count,
		},
	}
	item := types.WidgetItem{Type: "tasks"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "tasks" || output != "2" {
		t.Errorf("Expected title 'tasks' and body '2', got title '%s' and body '%s'", title, output)
	}
}

func TestTasksWidget_EdgeCase_NilTaskCount(t *testing.T) {
	w := initTestWidget(t, "tasks")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "tasks"}

	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	title, body, err := w.Render(item, ctxNil, settings)
	if err != nil || title != "tasks" || body != "0" {
		t.Errorf("Expected 'tasks' and '0' for nil TaskCount, got %q, %q, err=%v", title, body, err)
	}
}

func TestTasksWidget_ValidTaskCount(t *testing.T) {
	w := initTestWidget(t, "tasks")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "tasks"}

	taskCount := 3
	ctxVal := types.RenderContext{Data: types.StatusJSON{TaskCount: &taskCount}}
	_, bodyVal, err := w.Render(item, ctxVal, settings)
	if err != nil || bodyVal != "3" {
		t.Errorf("Expected '3' for TaskCount=3, got %q, err=%v", bodyVal, err)
	}
}

func TestTasksWidget_Interface(t *testing.T) {
	w := initTestWidget(t, "tasks")
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "tasks"}

	if nameStr := w.GetDisplayName(); nameStr == "" {
		t.Errorf("GetDisplayName() returned empty for tasks")
	}
	if defaultColor := w.GetDefaultColor(); defaultColor == "" {
		t.Errorf("GetDefaultColor() returned empty for tasks")
	}
	if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
		t.Errorf("GetBodyColor() returned empty for tasks")
	}
}
