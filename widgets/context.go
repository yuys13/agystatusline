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
	if ctx.Data.ContextWindow == nil || ctx.Data.ContextWindow.UsedPercentage == nil {
		if item.Raw {
			return "", "", nil
		}
		return "ctx", "", nil
	}

	pct := *ctx.Data.ContextWindow.UsedPercentage
	pctInt := int(pct)
	barLen := 15
	filled := min(barLen, max(0, pctInt*barLen/100))
	remainder := (pctInt * barLen) % 100

	var barBuilder strings.Builder
	for i := range barLen {
		if i < filled {
			barBuilder.WriteString("█")
		} else if i == filled && remainder > 0 {
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
	body := bar + " " + pctFmt

	if item.Raw {
		return "", body, nil
	}
	return "ctx", body, nil
}
