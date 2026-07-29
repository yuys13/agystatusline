package widgets

import (
	"fmt"

	"github.com/yuys13/agystatusline/types"
)

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
