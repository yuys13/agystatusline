package renderer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ColorEntry struct {
	Name         string
	DisplayName  string
	IsBackground bool
	Ansi16Code   string
	Ansi256Index int
	TruecolorHex string // e.g. "cc0000" (without #)
}

var colorMap = []ColorEntry{
	// Regular colors
	{Name: "black", DisplayName: "Black", IsBackground: false, Ansi16Code: "30", Ansi256Index: 16, TruecolorHex: "000000"},
	{Name: "red", DisplayName: "Red", IsBackground: false, Ansi16Code: "31", Ansi256Index: 160, TruecolorHex: "cc0000"},
	{Name: "green", DisplayName: "Green", IsBackground: false, Ansi16Code: "32", Ansi256Index: 70, TruecolorHex: "4e9a06"},
	{Name: "yellow", DisplayName: "Yellow", IsBackground: false, Ansi16Code: "33", Ansi256Index: 178, TruecolorHex: "c4a000"},
	{Name: "blue", DisplayName: "Blue", IsBackground: false, Ansi16Code: "34", Ansi256Index: 26, TruecolorHex: "3465a4"},
	{Name: "magenta", DisplayName: "Magenta", IsBackground: false, Ansi16Code: "35", Ansi256Index: 96, TruecolorHex: "75507b"},
	{Name: "cyan", DisplayName: "Cyan", IsBackground: false, Ansi16Code: "36", Ansi256Index: 30, TruecolorHex: "06989a"},
	{Name: "white", DisplayName: "White", IsBackground: false, Ansi16Code: "37", Ansi256Index: 188, TruecolorHex: "d3d7cf"},

	// Bright colors
	{Name: "brightBlack", DisplayName: "Bright Black", IsBackground: false, Ansi16Code: "90", Ansi256Index: 59, TruecolorHex: "555753"},
	{Name: "brightRed", DisplayName: "Bright Red", IsBackground: false, Ansi16Code: "91", Ansi256Index: 203, TruecolorHex: "ef2929"},
	{Name: "brightGreen", DisplayName: "Bright Green", IsBackground: false, Ansi16Code: "92", Ansi256Index: 155, TruecolorHex: "8ae234"},
	{Name: "brightYellow", DisplayName: "Bright Yellow", IsBackground: false, Ansi16Code: "93", Ansi256Index: 227, TruecolorHex: "fce94f"},
	{Name: "brightBlue", DisplayName: "Bright Blue", IsBackground: false, Ansi16Code: "94", Ansi256Index: 111, TruecolorHex: "729fcf"},
	{Name: "brightMagenta", DisplayName: "Bright Magenta", IsBackground: false, Ansi16Code: "95", Ansi256Index: 140, TruecolorHex: "ad7fa8"},
	{Name: "brightCyan", DisplayName: "Bright Cyan", IsBackground: false, Ansi16Code: "96", Ansi256Index: 80, TruecolorHex: "34e2e2"},
	{Name: "brightWhite", DisplayName: "Bright White", IsBackground: false, Ansi16Code: "97", Ansi256Index: 231, TruecolorHex: "eeeeec"},

	// Background colors
	{Name: "bgBlack", DisplayName: "Black", IsBackground: true, Ansi16Code: "40", Ansi256Index: 16, TruecolorHex: "000000"},
	{Name: "bgRed", DisplayName: "Red", IsBackground: true, Ansi16Code: "41", Ansi256Index: 160, TruecolorHex: "cc0000"},
	{Name: "bgGreen", DisplayName: "Green", IsBackground: true, Ansi16Code: "42", Ansi256Index: 70, TruecolorHex: "4e9a06"},
	{Name: "bgYellow", DisplayName: "Yellow", IsBackground: true, Ansi16Code: "43", Ansi256Index: 178, TruecolorHex: "c4a000"},
	{Name: "bgBlue", DisplayName: "Blue", IsBackground: true, Ansi16Code: "44", Ansi256Index: 26, TruecolorHex: "3465a4"},
	{Name: "bgMagenta", DisplayName: "Magenta", IsBackground: true, Ansi16Code: "45", Ansi256Index: 96, TruecolorHex: "75507b"},
	{Name: "bgCyan", DisplayName: "Cyan", IsBackground: true, Ansi16Code: "46", Ansi256Index: 30, TruecolorHex: "06989a"},
	{Name: "bgWhite", DisplayName: "White", IsBackground: true, Ansi16Code: "47", Ansi256Index: 188, TruecolorHex: "d3d7cf"},

	// Bright background colors
	{Name: "bgBrightBlack", DisplayName: "Bright Black", IsBackground: true, Ansi16Code: "100", Ansi256Index: 59, TruecolorHex: "555753"},
	{Name: "bgBrightRed", DisplayName: "Bright Red", IsBackground: true, Ansi16Code: "101", Ansi256Index: 203, TruecolorHex: "ef2929"},
	{Name: "bgBrightGreen", DisplayName: "Bright Green", IsBackground: true, Ansi16Code: "102", Ansi256Index: 155, TruecolorHex: "8ae234"},
	{Name: "bgBrightYellow", DisplayName: "Bright Yellow", IsBackground: true, Ansi16Code: "103", Ansi256Index: 227, TruecolorHex: "fce94f"},
	{Name: "bgBrightBlue", DisplayName: "Bright Blue", IsBackground: true, Ansi16Code: "104", Ansi256Index: 111, TruecolorHex: "729fcf"},
	{Name: "bgBrightMagenta", DisplayName: "Bright Magenta", IsBackground: true, Ansi16Code: "105", Ansi256Index: 140, TruecolorHex: "ad7fa8"},
	{Name: "bgBrightCyan", DisplayName: "Bright Cyan", IsBackground: true, Ansi16Code: "106", Ansi256Index: 80, TruecolorHex: "34e2e2"},
	{Name: "bgBrightWhite", DisplayName: "Bright White", IsBackground: true, Ansi16Code: "107", Ansi256Index: 231, TruecolorHex: "eeeeec"},
}

// GetColorAnsiCode returns the raw ANSI escape code for a color.
func GetColorAnsiCode(colorName string, colorLevel string, isBg bool) string {
	if colorName == "" {
		return ""
	}

	// Handle gradient specifier (collapse to the first stop as solid for powerline boundaries)
	if strings.HasPrefix(colorName, "gradient:") {
		// e.g. "gradient:hex:FF0000,hex:0000FF" or "gradient:red,blue"
		stops := parseGradientStops(colorName)
		if len(stops) > 0 {
			first := stops[0]
			if colorLevel == "ansi16" {
				return ""
			}
			if colorLevel == "ansi256" {
				code := rgbToAnsi256(first.R, first.G, first.B)
				if isBg {
					return fmt.Sprintf("\x1b[48;5;%dm", code)
				}
				return fmt.Sprintf("\x1b[38;5;%dm", code)
			}
			// truecolor
			if isBg {
				return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", first.R, first.G, first.B)
			}
			return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", first.R, first.G, first.B)
		}
		return ""
	}

	// Handle ansi256:X
	if strings.HasPrefix(colorName, "ansi256:") {
		codeStr := colorName[8:]
		code, err := strconv.Atoi(codeStr)
		if err == nil && code >= 0 && code <= 255 {
			if isBg {
				return fmt.Sprintf("\x1b[48;5;%dm", code)
			}
			return fmt.Sprintf("\x1b[38;5;%dm", code)
		}
		return ""
	}

	// Handle hex:XXXXXX
	if strings.HasPrefix(colorName, "hex:") {
		hexStr := colorName[4:]
		if len(hexStr) == 6 {
			var r, g, b int
			_, err := fmt.Sscanf(hexStr, "%02x%02x%02x", &r, &g, &b)
			if err == nil {
				if isBg {
					return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
				}
				return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
			}
		}
		return ""
	}

	// Lookup named color
	for _, entry := range colorMap {
		if entry.Name == colorName {
			switch colorLevel {
			case "ansi256":
				if isBg {
					return fmt.Sprintf("\x1b[48;5;%dm", entry.Ansi256Index)
				}
				return fmt.Sprintf("\x1b[38;5;%dm", entry.Ansi256Index)
			case "truecolor":
				var r, g, b int
				fmt.Sscanf(entry.TruecolorHex, "%02x%02x%02x", &r, &g, &b)
				if isBg {
					return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
				}
				return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
			default: // ansi16
				return fmt.Sprintf("\x1b[%sm", entry.Ansi16Code)
			}
		}
	}

	return ""
}

// BgToFg converts a background color name (e.g. bgRed) to its foreground equivalent (e.g. red).
func BgToFg(colorName string) string {
	if colorName == "" {
		return ""
	}

	if strings.HasPrefix(colorName, "ansi256:") || strings.HasPrefix(colorName, "hex:") || strings.HasPrefix(colorName, "gradient:") {
		return colorName
	}

	if strings.HasPrefix(colorName, "bgBright") {
		// bgBrightRed -> brightRed
		base := colorName[8:]
		if len(base) > 0 {
			return "bright" + strings.ToUpper(base[:1]) + strings.ToLower(base[1:])
		}
	} else if strings.HasPrefix(colorName, "bg") {
		// bgRed -> red
		base := colorName[2:]
		if len(base) > 0 {
			return strings.ToLower(base[:1]) + base[1:]
		}
	}

	return colorName
}

// applyParensDim dims each (...) span in the text.
func applyParensDim(text string, bold bool) string {
	intensityReset := "\x1b[22m"
	if bold {
		intensityReset = "\x1b[22;1m"
	}
	re := regexp.MustCompile(`\([^()]*\)`)
	return re.ReplaceAllStringFunc(text, func(span string) string {
		return "\x1b[2m" + span + intensityReset
	})
}

// ApplyColors applies foreground, background, bold, and dim styling to text.
func ApplyColors(text string, fgColor, bgColor string, bold *bool, colorLevel string, dim interface{}) string {
	isBold := bold != nil && *bold
	isDim := dim == true

	styledText := text
	if dim == "parens" {
		styledText = applyParensDim(text, isBold)
	}

	if fgColor == "" && bgColor == "" && !isBold && !isDim {
		return styledText
	}

	var prefix, suffix string

	if isBold {
		prefix += "\x1b[1m"
	}
	if isDim {
		prefix += "\x1b[2m"
	}
	if isBold || isDim {
		suffix = "\x1b[22m" + suffix
	}

	if bgColor != "" {
		bgCode := GetColorAnsiCode(bgColor, colorLevel, true)
		if bgCode != "" {
			prefix += bgCode
			suffix = "\x1b[49m" + suffix
		}
	}

	if fgColor != "" {
		// Check if it is a gradient and we have truecolor/ansi256 support
		if strings.HasPrefix(fgColor, "gradient:") && colorLevel != "ansi16" {
			stops := parseGradientStops(fgColor)
			if len(stops) > 0 {
				return prefix + applyGradientToText(styledText, stops, colorLevel) + "\x1b[39m" + suffix
			}
		}

		fgCode := GetColorAnsiCode(fgColor, colorLevel, false)
		if fgCode != "" {
			prefix += fgCode
			suffix = "\x1b[39m" + suffix
		}
	}

	return prefix + styledText + suffix
}

type RGB struct {
	R, G, B int
}

// parseGradientStops parses a gradient specifier like "gradient:hex:FF0000,hex:0000FF"
func parseGradientStops(spec string) []RGB {
	if !strings.HasPrefix(spec, "gradient:") {
		return nil
	}
	content := spec[9:]
	parts := strings.Split(content, ",")
	var stops []RGB

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "hex:") {
			hexStr := p[4:]
			if len(hexStr) == 6 {
				var r, g, b int
				if _, err := fmt.Sscanf(hexStr, "%02x%02x%02x", &r, &g, &b); err == nil {
					stops = append(stops, RGB{R: r, G: g, B: b})
				}
			}
		} else {
			// Lookup named colors in map
			for _, entry := range colorMap {
				if entry.Name == p && !entry.IsBackground {
					var r, g, b int
					fmt.Sscanf(entry.TruecolorHex, "%02x%02x%02x", &r, &g, &b)
					stops = append(stops, RGB{R: r, G: g, B: b})
					break
				}
			}
		}
	}
	return stops
}

func rgbToAnsi256(r, g, b int) int {
	// Standard mapping: 16-231 is 6x6x6 color cube
	// Red, Green, Blue in [0, 5]
	qr := (r * 5) / 255
	qg := (g * 5) / 255
	qb := (b * 5) / 255
	return 16 + 36*qr + 6*qg + qb
}

// applyGradientToText interpolates colors char-by-char across the string.
func applyGradientToText(text string, stops []RGB, colorLevel string) string {
	if len(stops) == 0 {
		return text
	}
	if len(stops) == 1 {
		first := stops[0]
		return wrapSolidColor(text, first, colorLevel)
	}

	runes := []rune(text)
	length := len(runes)
	if length == 0 {
		return ""
	}

	var builder strings.Builder
	for i, r := range runes {
		// Calculate position in gradient (float between 0.0 and 1.0)
		var t float64
		if length > 1 {
			t = float64(i) / float64(length-1)
		} else {
			t = 0.0
		}

		// Find the segment between two stops
		numSegments := len(stops) - 1
		scaledT := t * float64(numSegments)
		segmentIndex := int(scaledT)
		if segmentIndex >= numSegments {
			segmentIndex = numSegments - 1
		}
		localT := scaledT - float64(segmentIndex)

		c1 := stops[segmentIndex]
		c2 := stops[segmentIndex+1]

		// Interpolate RGB
		ir := int(float64(c1.R) + float64(c2.R-c1.R)*localT)
		ig := int(float64(c1.G) + float64(c2.G-c1.G)*localT)
		ib := int(float64(c1.B) + float64(c2.B-c1.B)*localT)

		if colorLevel == "truecolor" {
			builder.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%c", ir, ig, ib, r))
		} else { // ansi256
			code := rgbToAnsi256(ir, ig, ib)
			builder.WriteString(fmt.Sprintf("\x1b[38;5;%dm%c", code, r))
		}
	}
	// Return string with reset at the end
	return builder.String()
}

func wrapSolidColor(text string, c RGB, colorLevel string) string {
	if colorLevel == "truecolor" {
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s", c.R, c.G, c.B, text)
	}
	code := rgbToAnsi256(c.R, c.G, c.B)
	return fmt.Sprintf("\x1b[38;5;%dm%s", code, text)
}

// PowerlineThemeColors defines text/bg color names for each position.
type PowerlineThemeColors struct {
	Fg []string
	Bg []string
}

type PowerlineTheme struct {
	Name       string
	Colors16   *PowerlineThemeColors
	Colors256  *PowerlineThemeColors
	Truecolor  *PowerlineThemeColors
}

var powerlineThemes = map[string]*PowerlineTheme{
	"nord": {
		Name: "Nord",
		Colors16: &PowerlineThemeColors{
			Fg: []string{"black", "brightWhite", "brightWhite", "black", "black"},
			Bg: []string{"bgBrightCyan", "bgBrightBlack", "bgBlue", "bgBrightYellow", "bgBrightGreen"},
		},
		Colors256: &PowerlineThemeColors{
			Fg: []string{"ansi256:16", "ansi256:254", "ansi256:231", "ansi256:231", "ansi256:16"},
			Bg: []string{"ansi256:73", "ansi256:239", "ansi256:25", "ansi256:96", "ansi256:152"},
		},
		Truecolor: &PowerlineThemeColors{
			Fg: []string{"hex:2e3440", "hex:d8dee9", "hex:fdf6e3", "hex:2e3440", "hex:2e3440"},
			Bg: []string{"hex:88c0d0", "hex:4c566a", "hex:5e81ac", "hex:b48ead", "hex:a3be8c"},
		},
	},
}

// GetPowerlineTheme fetches powerline color scheme mappings.
func GetPowerlineTheme(name string) *PowerlineTheme {
	return powerlineThemes[name]
}
