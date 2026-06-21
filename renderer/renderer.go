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

	padding := settings.DefaultPadding
	separator := " | "
	if settings.DefaultSeparator != "" {
		separator = formatSeparator(settings.DefaultSeparator)
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
			if ctx.Minimalist {
				raw := true
				effectiveItem.RawValue = &raw
			}

			title, body, err := w.Render(effectiveItem, ctx, settings)
			if err != nil || (title == "" && body == "") {
				rendered = append(rendered, PreRenderedWidget{Item: item})
				continue
			}

			visibleText := body
			if title != "" {
				visibleText = title + " " + body
			}

			colored := visibleText
			if settings.Powerline.Enabled {
				colored = visibleText
			} else {
				preserveColors := item.PreserveColors != nil && *item.PreserveColors
				if preserveColors {
					colored = visibleText
				} else {
					bold := false
					if settings.GlobalBold || (item.Bold != nil && *item.Bold) {
						bold = true
					}
					colorLevelStr := "ansi16"
					if settings.ColorLevel == 2 {
						colorLevelStr = "ansi256"
					} else if settings.ColorLevel == 3 {
						colorLevelStr = "truecolor"
					}

					titleColored := ""
					if title != "" && !(item.RawValue != nil && *item.RawValue) && !settings.MinimalistMode {
						titleColored = ApplyColors(title, "brightBlack", "", nil, colorLevelStr, nil)
					}

					bodyColor := item.Color
					if bodyColor == "" {
						bodyColor = w.GetBodyColor(effectiveItem, ctx)
					}
					if bodyColor == "" {
						bodyColor = w.GetDefaultColor()
					}

					bodyColored := ApplyColors(body, bodyColor, item.BackgroundColor, &bold, colorLevelStr, item.Dim)

					if titleColored != "" {
						colored = titleColored + " " + bodyColored
					} else {
						colored = bodyColored
					}
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

			for i, r := range activeRendered {
				parts = append(parts, padding+r.Content+padding)
				if i < len(activeRendered)-1 {
					parts = append(parts, separator)
				}
			}

			lineStr = strings.Join(parts, "")
		}

		if ctx.TerminalWidth != nil && *ctx.TerminalWidth > 0 {
			termWidth := *ctx.TerminalWidth
			if ctx.IsPreview {
				termWidth = termWidth - 6
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
	content        string
	bgColor        string
	fgColor        string
	mergesWithNext bool
	widgetType     string
}

func renderPowerline(rendered []PreRenderedWidget, settings types.Settings, ctx types.RenderContext) string {
	var themeColors *PowerlineThemeColors
	themeName := settings.Powerline.Theme
	if themeName != "" && themeName != "custom" {
		theme := GetPowerlineTheme(themeName)
		if theme != nil {
			if settings.ColorLevel == 2 && theme.Colors256 != nil {
				themeColors = theme.Colors256
			} else if settings.ColorLevel == 3 && theme.Truecolor != nil {
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
		if fgColor == "" {
			w := widgets.GetWidget(r.Item.Type)
			if w != nil {
				fgColor = w.GetDefaultColor()
			}
		}
		bgColor := r.Item.BackgroundColor

		if themeColors != nil && len(themeColors.Bg) > 0 {
			fgColor = themeColors.Fg[themeColorIndex%len(themeColors.Fg)]
			bgColor = themeColors.Bg[themeColorIndex%len(themeColors.Bg)]
		}

		mergesWithNext := false
		if r.Item.Merge == true || r.Item.Merge == "no-padding" {
			mergesWithNext = true
		}

		// Advancing color index only if it does not merge
		if themeColors != nil && len(themeColors.Bg) > 0 && !mergesWithNext {
			themeColorIndex++
		}

		padding := settings.DefaultPadding
		paddedContent := padding + r.Content + padding

		elements = append(elements, powerlineElement{
			content:        paddedContent,
			bgColor:        bgColor,
			fgColor:        fgColor,
			mergesWithNext: mergesWithNext,
			widgetType:     r.Item.Type,
		})
	}

	if len(elements) == 0 {
		return ""
	}

	var builder strings.Builder
	separators := settings.Powerline.Separators
	if len(separators) == 0 {
		separators = []string{"\uE0B0"}
	}
	sep := separators[0]

	colorLevel := "ansi16"
	if settings.ColorLevel == 2 {
		colorLevel = "ansi256"
	} else if settings.ColorLevel == 3 {
		colorLevel = "truecolor"
	}

	// Prepend StartCap if configured
	startCaps := settings.Powerline.StartCaps
	if len(startCaps) > 0 && startCaps[0] != "" {
		startCap := startCaps[0]
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
		bold := false
		if settings.GlobalBold {
			bold = true
		}

		fgCode := GetColorAnsiCode(el.fgColor, colorLevel, false)
		bgCode := GetColorAnsiCode(el.bgColor, colorLevel, true)

		if bold {
			builder.WriteString("\x1b[1m")
		}
		builder.WriteString(fgCode)
		builder.WriteString(bgCode)
		builder.WriteString(el.content)
		builder.WriteString("\x1b[49m\x1b[39m")
		if bold {
			builder.WriteString("\x1b[22m")
		}

		if i < len(elements)-1 && !el.mergesWithNext {
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
	endCaps := settings.Powerline.EndCaps
	if len(endCaps) > 0 && endCaps[0] != "" {
		endCap := endCaps[0]
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
