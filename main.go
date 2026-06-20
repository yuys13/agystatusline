package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
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
	settingsPath = filepath.Join(home, ".config", "agystatusline", "settings.json")
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
			
			bytes, err := json.MarshalIndent(defaults, "", "  ")
			if err != nil {
				return defaults, err
			}
			
			err = os.WriteFile(settingsPath, bytes, 0644)
			if err != nil {
				return defaults, err
			}
			return defaults, nil
		}
		lastLoadError = "settings.json could not be read"
		return types.DefaultSettings(), err
	}

	var settings types.Settings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		lastLoadError = "settings.json is not valid JSON"
		return types.DefaultSettings(), nil
	}

	settings.Lines = upgradeLegacyWidgetTypes(settings.Lines)
	return settings, nil
}

func upgradeLegacyWidgetTypes(lines [][]types.WidgetItem) [][]types.WidgetItem {
	for i, line := range lines {
		for j, item := range line {
			if item.Type == "git-pr" {
				lines[i][j].Type = "git-review"
			}
		}
	}
	return lines
}

func main() {
	// Parse CLI options
	args := os.Args
	if contains(args, "--version") {
		fmt.Println("agystatusline version 1.0.0")
		os.Exit(0)
	}

	configPath, args := parseConfigArg(args)
	if configPath != "" {
		initConfigPath(configPath)
	}

	if contains(args, "--hook") {
		// Hook stub
		os.Exit(0)
	}

	// Register all widgets
	widgets.RegisterAll()

	// Check if stdin is a TTY
	if isatty.IsTerminal(os.Stdin.Fd()) {
		// Interactive TUI mode (will launch Bubble Tea TUI)
		settings, err := loadSettings()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to load settings:", err)
			os.Exit(1)
		}
		
		err = tui.RunTUI(settings, settingsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "TUI error:", err)
			os.Exit(1)
		}
	} else {
		// Piped non-TTY mode
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading stdin:", err)
			os.Exit(1)
		}

		if len(strings.TrimSpace(string(bytes))) == 0 {
			fmt.Fprintln(os.Stderr, "No input received")
			os.Exit(1)
		}

		var status types.StatusJSON
		err = json.Unmarshal(bytes, &status)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid status JSON format:", err)
			os.Exit(1)
		}

		settings, err := loadSettings()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to load settings:", err)
			os.Exit(1)
		}

		// Build render context
		// Terminal width detection (can use os.Stdout term size or telemetry data)
		termWidth := 80 // fallback
		if status.TerminalWidth != nil {
			termWidth = *status.TerminalWidth
		}
		ctx := types.RenderContext{
			Data:               status,
			TerminalWidth:      &termWidth,
			IsPreview:          false,
			Minimalist:         settings.MinimalistMode,
			GitCacheTTLSeconds: settings.GitCacheTTLSeconds,
		}

		lines := renderer.RenderStatusLines(settings, ctx)
		for _, line := range lines {
			// Claude Code's reset handling
			fmt.Println("\x1b[0m" + line)
		}
	}
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
