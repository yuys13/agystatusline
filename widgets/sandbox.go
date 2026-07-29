package widgets

import (
	"github.com/yuys13/agystatusline/types"
)

// SandboxWidget displays the sandbox enabled status.
type SandboxWidget struct{}

func (s *SandboxWidget) GetDefaultColor() string { return "yellow" }
func (s *SandboxWidget) GetDisplayName() string  { return "Sandbox" }
func (s *SandboxWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	if ctx.Data.Sandbox != nil && ctx.Data.Sandbox.Enabled != nil && *ctx.Data.Sandbox.Enabled {
		return "brightGreen"
	}
	return "brightBlack"
}

func (s *SandboxWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	if ctx.Data.Sandbox == nil || ctx.Data.Sandbox.Enabled == nil {
		return "", "", nil
	}

	enabled := *ctx.Data.Sandbox.Enabled
	valStr := "off"
	if enabled {
		valStr = "on"
	}

	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		// statusline.sh style coloring:
		// sandbox is gray (ansi 90), ON is green and bold (ansi 92), off is gray (ansi 90)
		var bodyStr string
		if enabled {
			bodyStr = "\x1b[90msandbox\x1b[39m \x1b[92m\x1b[1mON\x1b[22m\x1b[39m"
		} else {
			bodyStr = "\x1b[90msandbox off\x1b[39m"
		}
		return "", bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", valStr, nil
	}

	if enabled {
		return "sandbox", "on", nil
	}
	return "sandbox", "off", nil
}
