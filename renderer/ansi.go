package renderer

import (
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

// ansiRegexp matches ANSI escape sequences.
var ansiRegexp = regexp.MustCompile(`[\x1b\x9b][[()#;?]*[0-9;]*[a-zA-Z]`)

// StripAnsi removes ANSI escape sequences from a string.
func StripAnsi(str string) string {
	return ansiRegexp.ReplaceAllString(str, "")
}

// GetVisibleWidth returns the visible width of a string in terminal cells,
// ignoring ANSI codes and taking double-width characters into account.
func GetVisibleWidth(str string) int {
	clean := StripAnsi(str)
	return runewidth.StringWidth(clean)
}

// RenderOsc8Link generates an OSC 8 terminal hyperlink.
func RenderOsc8Link(url, text string) string {
	return "\x1b]8;;" + url + "\x1b\\" + text + "\x1b]8;;\x1b\\"
}

type parsedEscape struct {
	nextIndex int
	sequence  string
	isOsc8    bool
	isClose   bool
}

func parseEscape(text string, index int) (parsedEscape, bool) {
	if index >= len(text) {
		return parsedEscape{}, false
	}
	if text[index] != '\x1b' {
		return parsedEscape{}, false
	}
	if index+1 >= len(text) {
		return parsedEscape{nextIndex: len(text), sequence: text[index:]}, true
	}

	next := text[index+1]
	if next == '[' { // CSI
		for i := index + 2; i < len(text); i++ {
			b := text[i]
			if b >= 0x40 && b <= 0x7e {
				end := i + 1
				return parsedEscape{nextIndex: end, sequence: text[index:end]}, true
			}
		}
		return parsedEscape{nextIndex: len(text), sequence: text[index:]}, true
	}

	if next == ']' { // OSC
		for i := index + 2; i < len(text); i++ {
			if text[i] == '\x07' {
				end := i + 1
				seq := text[index:end]
				body := text[index+2 : i]
				isOsc8 := strings.HasPrefix(body, "8;")
				isClose := isOsc8 && (body == "8;" || strings.HasSuffix(body, ";"))
				return parsedEscape{nextIndex: end, sequence: seq, isOsc8: isOsc8, isClose: isClose}, true
			}
			if text[i] == '\x1b' && i+1 < len(text) && text[i+1] == '\\' {
				end := i + 2
				seq := text[index:end]
				body := text[index+2 : i]
				isOsc8 := strings.HasPrefix(body, "8;")
				isClose := isOsc8 && (body == "8;;" || strings.HasSuffix(body, ";"))
				return parsedEscape{nextIndex: end, sequence: seq, isOsc8: isOsc8, isClose: isClose}, true
			}
		}
		return parsedEscape{nextIndex: len(text), sequence: text[index:]}, true
	}

	return parsedEscape{nextIndex: index + 2, sequence: text[index : index+2]}, true
}

// TruncateStyledText truncates styled terminal text preserving ANSI sequences.
func TruncateStyledText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if GetVisibleWidth(text) <= maxWidth {
		return text
	}

	ellipsis := "..."
	ellipsisWidth := GetVisibleWidth(ellipsis)
	if maxWidth <= ellipsisWidth {
		return strings.Repeat(".", maxWidth)
	}

	targetWidth := maxWidth - ellipsisWidth
	var output strings.Builder
	currentWidth := 0
	didTruncate := false
	osc8Opened := false

	runes := []rune(text)
	runeIndex := 0

	for runeIndex < len(runes) {
		strIdx := len(string(runes[:runeIndex]))
		if esc, ok := parseEscape(text, strIdx); ok {
			output.WriteString(esc.sequence)
			if esc.isOsc8 {
				if esc.isClose {
					osc8Opened = false
				} else {
					osc8Opened = true
				}
			}
			runeIndex += len([]rune(esc.sequence))
			continue
		}

		r := runes[runeIndex]
		width := runewidth.RuneWidth(r)

		if currentWidth+width > targetWidth {
			didTruncate = true
			break
		}

		output.WriteRune(r)
		currentWidth += width
		runeIndex++
	}

	if !didTruncate {
		return text
	}

	if osc8Opened {
		output.WriteString("\x1b]8;;\x1b\\")
	}

	output.WriteString(ellipsis)
	return output.String()
}
