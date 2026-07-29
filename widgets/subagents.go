package widgets

import (
	"strconv"

	"github.com/yuys13/agystatusline/types"
)

// SubagentsWidget displays count of subagents.
type SubagentsWidget struct{}

func (s *SubagentsWidget) GetDefaultColor() string { return "brightWhite" }
func (s *SubagentsWidget) GetDisplayName() string  { return "Subagents" }
func (s *SubagentsWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightWhite"
}

func (s *SubagentsWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	count := 0
	if ctx.Data.Subagents != nil {
		if list, ok := ctx.Data.Subagents.([]any); ok {
			count = len(list)
		} else if num, ok := ctx.Data.Subagents.(float64); ok {
			count = int(num)
		}
	}

	countStr := strconv.Itoa(count)
	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		titleStr := "\x1b[90msubagents\x1b[39m"
		bodyStr := "\x1b[97m\x1b[1m" + countStr + "\x1b[22m\x1b[39m"
		return titleStr, bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", countStr, nil
	}
	return "subagents", countStr, nil
}
