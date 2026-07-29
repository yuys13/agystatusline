package widgets

import (
	"fmt"
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
