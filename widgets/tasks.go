package widgets

import (
	"strconv"

	"github.com/yuys13/agystatusline/types"
)

// TasksWidget displays count of background tasks.
type TasksWidget struct{}

func (t *TasksWidget) GetDefaultColor() string { return "brightWhite" }
func (t *TasksWidget) GetDisplayName() string  { return "Tasks" }
func (t *TasksWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightWhite"
}

func (t *TasksWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	count := 0
	if ctx.Data.TaskCount != nil {
		count = *ctx.Data.TaskCount
	}

	countStr := strconv.Itoa(count)
	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		titleStr := "\x1b[90mtasks\x1b[39m"
		bodyStr := "\x1b[97m\x1b[1m" + countStr + "\x1b[22m\x1b[39m"
		return titleStr, bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", countStr, nil
	}
	return "tasks", countStr, nil
}
