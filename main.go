package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pelletier/go-toml/v2"
	"github.com/yuys13/agystatusline/renderer"
	"github.com/yuys13/agystatusline/tui"
	"github.com/yuys13/agystatusline/types"
	"github.com/yuys13/agystatusline/widgets"
)

var (
	settingsPath  = ""
	lastLoadError = ""
)

func init() {
	home, _ := os.UserHomeDir()
	settingsPath = filepath.Join(home, ".config", "agystatusline", "settings.toml")
}

func initConfigPath(filePath string) {
	if filePath != "" {
		if filepath.IsAbs(filePath) {
			settingsPath = filePath
		} else {
			abs, err := filepath.Abs(filePath)
			if err == nil {
				settingsPath = abs
			}
		}
	}
}

func parseConfigArg(args []string) (string, []string) {
	var path string
	var remaining []string

	for i := 0; i < len(args); i++ {
		if args[i] == "--config" && i+1 < len(args) {
			path = args[i+1]
			i++ // skip next arg
		} else {
			remaining = append(remaining, args[i])
		}
	}
	return path, remaining
}

func loadSettings() (types.Settings, error) {
	lastLoadError = ""

	// Read file
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Write default settings
			defaults := types.DefaultSettings()
			dir := filepath.Dir(settingsPath)
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return defaults, err
			}

			bytes, err := toml.Marshal(defaults)
			if err != nil {
				return defaults, err
			}

			err = os.WriteFile(settingsPath, bytes, 0644)
			if err != nil {
				return defaults, err
			}
			return defaults, nil
		}
		lastLoadError = "settings.toml could not be read"
		return types.DefaultSettings(), err
	}

	var settings types.Settings
	err = toml.Unmarshal(data, &settings)
	if err != nil {
		lastLoadError = "settings.toml is not valid TOML"
		return types.DefaultSettings(), nil
	}

	return settings, nil
}

func main() {
	os.Exit(runMain(os.Args, os.Stdin, os.Stdout, os.Stderr, isatty.IsTerminal))
}

func runMain(args []string, stdin io.Reader, stdout, stderr io.Writer, isTerminal func(fd uintptr) bool) int {
	home, _ := os.UserHomeDir()
	settingsPath = filepath.Join(home, ".config", "agystatusline", "settings.toml")

	if contains(args, "--version") {
		_, _ = fmt.Fprintln(stdout, "agystatusline version 1.0.0")
		return 0
	}

	if contains(args, "--help") || contains(args, "-h") {
		_, _ = fmt.Fprintln(stdout, "Usage: agystatusline [--version] [--help] [--config <path>] [--hook]")
		return 0
	}

	configPath, args := parseConfigArg(args)
	if configPath != "" {
		initConfigPath(configPath)
	}

	if contains(args, "--hook") {
		// Hook stub
		return 0
	}

	// Register all widgets
	widgets.RegisterAll()

	var stdinFd uintptr
	if f, ok := stdin.(*os.File); ok {
		stdinFd = f.Fd()
	}

	// Check if stdin is a TTY
	if isTerminal(stdinFd) {
		// Interactive TUI mode (will launch Bubble Tea TUI)
		settings, err := loadSettings()
		if err != nil {
			_, _ = fmt.Fprintln(stderr, "Failed to load settings:", err)
			return 1
		}

		err = tui.RunTUI(settings, settingsPath)
		if err != nil {
			_, _ = fmt.Fprintln(stderr, "TUI error:", err)
			return 1
		}
		return 0
	}

	// Piped non-TTY mode
	bytes, err := io.ReadAll(stdin)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "Error reading stdin:", err)
		return 1
	}

	if len(strings.TrimSpace(string(bytes))) == 0 {
		_, _ = fmt.Fprintln(stderr, "No input received")
		return 1
	}

	var status types.StatusJSON
	err = json.Unmarshal(bytes, &status)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "Invalid status JSON format:", err)
		return 1
	}

	settings, err := loadSettings()
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "Failed to load settings:", err)
		return 1
	}

	// Build render context
	termWidth := 80 // fallback
	if status.TerminalWidth != nil {
		termWidth = *status.TerminalWidth
	}
	ctx := types.RenderContext{
		Data:               status,
		TerminalWidth:      &termWidth,
		IsPreview:          false,
		Minimalist:         settings.General.Minimalist,
		GitCacheTTLSeconds: settings.General.GitCacheTTL,
	}

	lines := renderer.RenderStatusLines(settings, ctx)
	for _, line := range lines {
		// Output statusline with ANSI reset prefix
		_, _ = fmt.Fprintln(stdout, "\x1b[0m"+line)
	}
	return 0
}

func contains(slice []string, val string) bool {
	return slices.Contains(slice, val)
}
