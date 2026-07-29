package widgets

import (
	"fmt"
	"strings"

	"github.com/yuys13/agystatusline/types"
)

// ContextBarWidget displays a progress bar representing context window usage.
type ContextBarWidget struct{}

func (c *ContextBarWidget) GetDefaultColor() string { return "brightWhite" }
func (c *ContextBarWidget) GetDisplayName() string  { return "Context Bar" }

func (c *ContextBarWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	var pct float64
	if ctx.Data.ContextWindow != nil && ctx.Data.ContextWindow.UsedPercentage != nil {
		pct = *ctx.Data.ContextWindow.UsedPercentage
	}
	if pct >= 90 {
		return "brightRed"
	} else if pct >= 60 {
		return "brightYellow"
	}
	return "brightWhite"
}

func (c *ContextBarWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	var pct float64
	if ctx.Data.ContextWindow != nil && ctx.Data.ContextWindow.UsedPercentage != nil {
		pct = *ctx.Data.ContextWindow.UsedPercentage
	} else {
		return "ctx", "", nil
	}

	pctInt := int(pct)
	barLen := 15
	filled := pctInt * barLen / 100
	remainder := (pctInt * barLen) % 100

	var barBuilder strings.Builder
	for i := range barLen {
		if i < filled {
			barBuilder.WriteString("█")
		} else if i == filled {
			if remainder >= 75 {
				barBuilder.WriteString("▓")
			} else if remainder >= 50 {
				barBuilder.WriteString("▒")
			} else if remainder >= 25 {
				barBuilder.WriteString("░")
			} else {
				barBuilder.WriteString("·")
			}
		} else {
			barBuilder.WriteString("·")
		}
	}
	bar := barBuilder.String()

	pctFmt := fmt.Sprintf("%.1f%%", pct)

	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		var barColor string
		if pctInt >= 90 {
			barColor = "\x1b[91m"
		} else if pctInt >= 60 {
			barColor = "\x1b[93m"
		} else {
			barColor = "\x1b[97m"
		}
		titleStr := "\x1b[90mctx\x1b[39m"
		bodyStr := barColor + bar + "\x1b[39m \x1b[97m\x1b[1m" + pctFmt + "\x1b[22m\x1b[39m"
		return titleStr, bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", bar + " " + pctFmt, nil
	}
	return "ctx", bar + " " + pctFmt, nil
}
