package renderer

import (
	"regexp"

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
