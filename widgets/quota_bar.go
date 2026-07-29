package widgets

import (
	"fmt"
	"strings"

	"github.com/yuys13/agystatusline/types"
)

// QuotaBarWidget displays a progress bar representing remaining quota.
type QuotaBarWidget struct{}

func (q *QuotaBarWidget) GetDefaultColor() string { return "brightGreen" }
func (q *QuotaBarWidget) GetDisplayName() string  { return "Quota Bar" }

func (q *QuotaBarWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	if ctx.Data.Quota == nil {
		return "brightGreen"
	}
	key := item.Metadata["key"]
	if key == "" {
		return "brightGreen"
	}
	quota, ok := ctx.Data.Quota[key]
	if !ok || quota.RemainingFraction == nil {
		return "brightGreen"
	}
	pct := *quota.RemainingFraction * 100.0
	if pct >= 50 {
		return "brightGreen"
	} else if pct >= 10 {
		return "brightYellow"
	}
	return "brightRed"
}

func (q *QuotaBarWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	if ctx.Data.Quota == nil {
		return "", "", nil
	}
	key := item.Metadata["key"]
	if key == "" {
		return "", "", nil
	}
	quota, ok := ctx.Data.Quota[key]
	if !ok || quota.RemainingFraction == nil {
		return "", "", nil
	}

	pct := *quota.RemainingFraction * 100.0
	pctInt := int(pct)
	barLen := 10
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

	label := item.CustomText
	if label == "" {
		switch key {
		case "gemini-5h":
			label = "5h"
		case "gemini-weekly":
			label = "7d"
		case "3p-weekly":
			label = "3p-7d"
		default:
			label = key
		}
	}

	resetStr := formatResetInSeconds(quota.ResetInSeconds)
	bodyVal := bar + " " + pctFmt
	if resetStr != "" {
		bodyVal = bodyVal + " (" + resetStr + ")"
	}

	if item.RawValue != nil && *item.RawValue {
		return "", bodyVal, nil
	}
	return label, bodyVal, nil
}
