package widgets

import (
	"strings"

	"github.com/yuys13/agystatusline/types"
)

// ModelWidget displays the active model name.
type ModelWidget struct{}

func (m *ModelWidget) GetDefaultColor() string { return "brightMagenta" }
func (m *ModelWidget) GetDisplayName() string  { return "Model" }
func (m *ModelWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightMagenta"
}

func (m *ModelWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	displayName := ctx.Data.Model.DisplayName
	if displayName == "" {
		displayName = ctx.Data.Model.ID
	}

	if displayName == "" {
		return "", "", nil
	}

	modelName := strings.TrimSpace(displayName)

	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		// Under statusline.sh spec, model name is italic magenta
		return "", "\x1b[3m\x1b[95m" + modelName + "\x1b[23m\x1b[39m", nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", modelName, nil
	}
	return "", modelName, nil
}
