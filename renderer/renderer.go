package renderer

import (
	"strings"

	"github.com/yuys13/agystatusline/types"
	"github.com/yuys13/agystatusline/widgets"
)

type PreRenderedWidget struct {
	Content     string
	PlainLength int
	Item        types.WidgetItem
}

// RenderStatusLines renders all statusline lines according to settings and context.
func RenderStatusLines(settings types.Settings, ctx types.RenderContext) []string {
	var results []string

	padding := settings.General.Padding
	separator := settings.General.Separator
	if separator == "" {
		separator = " · "
	} else {
		separator = formatSeparator(separator)
	}

	var colorLevelStr string
	switch settings.General.ColorLevel {
	case 2:
		colorLevelStr = "ansi256"
	case 3:
		colorLevelStr = "truecolor"
	default:
		colorLevelStr = "ansi16"
	}

	for _, lineItems := range settings.Lines {
		if len(lineItems) == 0 {
			results = append(results, "")
			continue
		}

		var rendered []PreRenderedWidget
		for _, item := range lineItems {
			if item.Type == "separator" || item.Type == "flex-separator" {
				rendered = append(rendered, PreRenderedWidget{Item: item})
				continue
			}

			w := widgets.GetWidget(item.Type)
			if w == nil {
				rendered = append(rendered, PreRenderedWidget{Item: item})
				continue
			}

			effectiveItem := item
			if ctx.Minimalist || settings.General.Minimalist {
				effectiveItem.Raw = true
			}

			title, body, err := w.Render(effectiveItem, ctx, settings)
			if err != nil || (title == "" && body == "") {
				rendered = append(rendered, PreRenderedWidget{Item: item})
				continue
			}

			if effectiveItem.Raw || ctx.Minimalist || settings.General.Minimalist {
				title = ""
			}

			visibleText := body
			if title != "" && body != "" {
				visibleText = title + " " + body
			} else if title != "" {
				visibleText = title
			}

			var colored string
			if settings.Powerline.Enabled {
				colored = visibleText
			} else {
				titleColored := ""
				if title != "" {
					titleColored = ApplyColors(title, "brightBlack", "", nil, colorLevelStr, nil)
				}

				bodyColor := item.Color
				if bodyColor == "" {
					bodyColor = w.GetBodyColor(effectiveItem, ctx)
				}
				if bodyColor == "" {
					bodyColor = w.GetDefaultColor()
				}

				bodyColored := ""
				if body != "" {
					bodyColored = ApplyColors(body, bodyColor, "", nil, colorLevelStr, nil)
				}

				if titleColored != "" && bodyColored != "" {
					colored = titleColored + " " + bodyColored
				} else if titleColored != "" {
					colored = titleColored
				} else {
					colored = bodyColored
				}
			}

			rendered = append(rendered, PreRenderedWidget{
				Content:     colored,
				PlainLength: GetVisibleWidth(visibleText),
				Item:        item,
			})
		}

		var lineStr string
		if settings.Powerline.Enabled {
			lineStr = renderPowerline(rendered, settings, ctx)
		} else {
			var parts []string

			var activeRendered []PreRenderedWidget
			for _, r := range rendered {
				if r.Item.Type == "separator" || r.Item.Type == "flex-separator" {
					continue
				}
				if r.Content != "" {
					activeRendered = append(activeRendered, r)
				}
			}

			coloredSeparator := ApplyColors(separator, "brightBlack", "", nil, colorLevelStr, nil)

			for i, r := range activeRendered {
				parts = append(parts, padding+r.Content+padding)
				if i < len(activeRendered)-1 {
					parts = append(parts, coloredSeparator)
				}
			}

			lineStr = strings.Join(parts, "")
		}

		if ctx.TerminalWidth != nil && *ctx.TerminalWidth > 0 {
			termWidth := *ctx.TerminalWidth
			if ctx.IsPreview {
				termWidth = termWidth - 6
			}
			if termWidth < 0 {
				termWidth = 0
			}
			visibleWidth := GetVisibleWidth(lineStr)
			if visibleWidth > termWidth {
				lineStr = TruncateStyledText(lineStr, termWidth)
			}
		}

		results = append(results, lineStr)
	}

	return results
}

func formatSeparator(sep string) string {
	switch sep {
	case "|":
		return " | "
	case " ":
		return " "
	case ",":
		return ", "
	case "-":
		return " - "
	}
	return sep
}

type powerlineElement struct {
	content    string
	bgColor    string
	fgColor    string
	widgetType string
}

func renderPowerline(rendered []PreRenderedWidget, settings types.Settings, ctx types.RenderContext) string {
	var themeColors *PowerlineThemeColors
	themeName := settings.Powerline.Theme
	if themeName == "" {
		themeName = "nord-aurora"
	}
	if themeName != "custom" {
		theme := GetPowerlineTheme(themeName)
		if theme != nil {
			if settings.General.ColorLevel == 2 && theme.Colors256 != nil {
				themeColors = theme.Colors256
			} else if settings.General.ColorLevel == 3 && theme.Truecolor != nil {
				themeColors = theme.Truecolor
			} else if theme.Colors16 != nil {
				themeColors = theme.Colors16
			}
		}
	}

	var elements []powerlineElement
	themeColorIndex := 0

	for _, r := range rendered {
		if r.Item.Type == "separator" || r.Item.Type == "flex-separator" {
			continue
		}
		if r.Content == "" {
			continue
		}

		fgColor := r.Item.Color
		bgColor := ""

		if themeColors != nil && len(themeColors.Bg) > 0 {
			if fgColor == "" {
				fgColor = themeColors.Fg[themeColorIndex%len(themeColors.Fg)]
			}
			bgColor = themeColors.Bg[themeColorIndex%len(themeColors.Bg)]
			themeColorIndex++
		} else {
			if fgColor == "" {
				w := widgets.GetWidget(r.Item.Type)
				if w != nil {
					fgColor = w.GetDefaultColor()
				}
			}
		}

		padding := settings.General.Padding
		paddedContent := padding + r.Content + padding

		elements = append(elements, powerlineElement{
			content:    paddedContent,
			bgColor:    bgColor,
			fgColor:    fgColor,
			widgetType: r.Item.Type,
		})
	}

	if len(elements) == 0 {
		return ""
	}

	var builder strings.Builder
	sep := settings.Powerline.Separator
	if sep == "" {
		sep = "\uE0B0"
	}

	var colorLevel string
	switch settings.General.ColorLevel {
	case 2:
		colorLevel = "ansi256"
	case 3:
		colorLevel = "truecolor"
	default:
		colorLevel = "ansi16"
	}

	// Prepend StartCap if configured
	if settings.Powerline.StartCaps != "" {
		startCap := settings.Powerline.StartCaps
		firstEl := elements[0]
		if firstEl.bgColor != "" {
			capFg := BgToFg(firstEl.bgColor)
			fgCode := GetColorAnsiCode(capFg, colorLevel, false)
			builder.WriteString(fgCode + startCap + "\x1b[39m")
		} else {
			builder.WriteString(startCap)
		}
	}

	for i, el := range elements {
		fgCode := GetColorAnsiCode(el.fgColor, colorLevel, false)
		bgCode := GetColorAnsiCode(el.bgColor, colorLevel, true)

		builder.WriteString(fgCode)
		builder.WriteString(bgCode)
		builder.WriteString(el.content)
		builder.WriteString("\x1b[49m\x1b[39m")

		if i < len(elements)-1 {
			nextEl := elements[i+1]
			sepFg := BgToFg(el.bgColor)
			sepBg := nextEl.bgColor

			if el.bgColor != "" && nextEl.bgColor != "" && el.bgColor == nextEl.bgColor {
				sepFg = el.fgColor
				sepBg = el.bgColor
			}

			sepFgCode := GetColorAnsiCode(sepFg, colorLevel, false)
			sepBgCode := GetColorAnsiCode(sepBg, colorLevel, true)

			builder.WriteString(sepFgCode)
			builder.WriteString(sepBgCode)
			builder.WriteString(sep)
			builder.WriteString("\x1b[49m\x1b[39m")
		}
	}

	// Append EndCap if configured
	if settings.Powerline.EndCaps != "" {
		endCap := settings.Powerline.EndCaps
		lastEl := elements[len(elements)-1]
		if lastEl.bgColor != "" {
			capFg := BgToFg(lastEl.bgColor)
			fgCode := GetColorAnsiCode(capFg, colorLevel, false)
			builder.WriteString(fgCode + endCap + "\x1b[39m")
		} else {
			builder.WriteString(endCap)
		}
	}

	return builder.String()
}
