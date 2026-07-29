package widgets

import (
	"strconv"

	"github.com/yuys13/agystatusline/types"
)

// ArtifactsWidget displays count of artifacts.
type ArtifactsWidget struct{}

func (a *ArtifactsWidget) GetDefaultColor() string { return "brightWhite" }
func (a *ArtifactsWidget) GetDisplayName() string  { return "Artifacts" }
func (a *ArtifactsWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightWhite"
}

func (a *ArtifactsWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	count := 0
	if ctx.Data.ArtifactCount != nil {
		count = *ctx.Data.ArtifactCount
	}

	countStr := strconv.Itoa(count)
	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		titleStr := "\x1b[90martifacts\x1b[39m"
		bodyStr := "\x1b[97m\x1b[1m" + countStr + "\x1b[22m\x1b[39m"
		return titleStr, bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", countStr, nil
	}
	return "artifacts", countStr, nil
}
