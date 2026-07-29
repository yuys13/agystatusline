package widgets

import (
	"github.com/yuys13/agystatusline/types"
	"testing"
)

func TestTasksWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("tasks")
	if w == nil {
		t.Fatalf("Tasks widget not found")
	}

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

func TestTasksWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("tasks")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "tasks"}

	// Nil task count
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	title, body, _ := w.Render(item, ctxNil, settings)
	if title != "tasks" || body != "0" {
		t.Errorf("Expected 'tasks' and '0' for nil TaskCount, got %q, %q", title, body)
	}

	// Valid task count
	taskCount := 3
	ctxVal := types.RenderContext{Data: types.StatusJSON{TaskCount: &taskCount}}
	_, bodyVal, _ := w.Render(item, ctxVal, settings)
	if bodyVal != "3" {
		t.Errorf("Expected '3' for TaskCount=3, got %q", bodyVal)
	}
}
