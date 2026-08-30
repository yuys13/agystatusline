package widgets

import (
	"github.com/yuys13/agystatusline/types"
)

// CustomTextWidget displays custom user-defined text.
type CustomTextWidget struct{}

func (c *CustomTextWidget) GetDefaultColor() string { return "white" }
func (c *CustomTextWidget) GetDisplayName() string  { return "Custom Text" }
func (c *CustomTextWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "white"
}

func (c *CustomTextWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	return "", item.Text, nil
}
