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

	modelName := strings.TrimSpace(displayName)
	if modelName == "" {
		modelName = "no-model"
	}

	return "", modelName, nil
}
