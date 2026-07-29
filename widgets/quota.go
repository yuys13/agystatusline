package widgets

import (
	"fmt"
	"strings"

	"github.com/yuys13/agystatusline/types"
)

func formatResetInSeconds(resetInSeconds *float64) string {
	if resetInSeconds == nil {
		return ""
	}
	secs := max(int(*resetInSeconds), 0)
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	} else if secs < 3600 {
		m := secs / 60
		s := secs % 60
		if s > 0 {
			return fmt.Sprintf("%dm %ds", m, s)
		} else {
			return fmt.Sprintf("%dm", m)
		}
	} else if secs < 86400 {
		h := secs / 3600
		m := (secs % 3600) / 60
		if m > 0 {
			return fmt.Sprintf("%dh %dm", h, m)
		} else {
			return fmt.Sprintf("%dh", h)
		}
	} else {
		d := secs / 86400
		h := (secs % 86400) / 3600
		if h > 0 {
			return fmt.Sprintf("%dd %dh", d, h)
		} else {
			return fmt.Sprintf("%dd", d)
		}
	}
}

// QuotaWidget displays quota limits and usage.
type QuotaWidget struct{}

func (q *QuotaWidget) GetDefaultColor() string { return "brightWhite" }
func (q *QuotaWidget) GetDisplayName() string  { return "Quota" }
func (q *QuotaWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightWhite"
}

func (q *QuotaWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	if ctx.Data.Quota == nil {
		return "", "", nil
	}

	key := item.Metadata["key"]
	if key == "" {
		return "", "", nil
	}

	quota, ok := ctx.Data.Quota[key]
	if !ok {
		return "", "", nil
	}

	displayMode := item.Metadata["display"]
	var valueStr string

	var pctStr string
	if quota.RemainingFraction != nil {
		pct := (*quota.RemainingFraction) * 100.0
		pctStr = fmt.Sprintf("%.2f%%", pct)
	}

	resetStr := formatResetInSeconds(quota.ResetInSeconds)

	switch displayMode {
	case "reset":
		if resetStr == "" {
			return "", "", nil
		}
		valueStr = resetStr
	case "quota":
		if pctStr == "" {
			return "", "", nil
		}
		valueStr = pctStr
	default:
		// Default: quota % + reset countdown
		if pctStr != "" && resetStr != "" {
			valueStr = fmt.Sprintf("%s (%s)", pctStr, resetStr)
		} else if pctStr != "" {
			valueStr = pctStr
		} else if resetStr != "" {
			valueStr = resetStr
		} else {
			return "", "", nil
		}
	}

	if item.RawValue != nil && *item.RawValue {
		return "", valueStr, nil
	}

	label := item.CustomText
	if label == "" {
		label = key
	}

	if displayMode == "reset" {
		return label + " (reset)", valueStr, nil
	}
	return label, valueStr, nil
}

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
	} else if pct >= 20 {
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
