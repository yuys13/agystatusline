package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfigArg(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantPath      string
		wantRemaining []string
	}{
		{
			name:          "Config flag present with value",
			args:          []string{"./agystatusline", "--config", "/tmp/custom.toml"},
			wantPath:      "/tmp/custom.toml",
			wantRemaining: []string{"./agystatusline"},
		},
		{
			name:          "No config flag present",
			args:          []string{"./agystatusline", "--other"},
			wantPath:      "",
			wantRemaining: []string{"./agystatusline", "--other"},
		},
		{
			name:          "Config flag without value",
			args:          []string{"./agystatusline", "--config"},
			wantPath:      "",
			wantRemaining: []string{"./agystatusline", "--config"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settingsPath = "default_settings.toml"
			path, remaining := parseConfigArg(tt.args)
			if path != tt.wantPath {
				t.Errorf("parseConfigArg() path = %q, want %q", path, tt.wantPath)
			}
			if len(remaining) != len(tt.wantRemaining) {
				t.Fatalf("parseConfigArg() remaining length = %d, want %d", len(remaining), len(tt.wantRemaining))
			}
			for i := range remaining {
				if remaining[i] != tt.wantRemaining[i] {
					t.Errorf("parseConfigArg() remaining[%d] = %q, want %q", i, remaining[i], tt.wantRemaining[i])
				}
			}
		})
	}
}

func TestLoadSettings(t *testing.T) {
	t.Run("valid settings", func(t *testing.T) {
		t.Run("Return default settings without creating file when file does not exist", func(t *testing.T) {
			tempDir := t.TempDir()
			customPath := filepath.Join(tempDir, "settings.toml")
			initConfigPath(customPath)
			lastLoadError = ""

			settings, err := loadSettings()
			if err != nil {
				t.Fatalf("loadSettings() unexpected error: %v", err)
			}
			if len(settings.Lines) != 1 {
				t.Errorf("loadSettings() lines count = %d, want 1", len(settings.Lines))
			}
			if _, statErr := os.Stat(customPath); !os.IsNotExist(statErr) {
				t.Errorf("Settings file was expected NOT to exist at %q, but exists", customPath)
			}
			if lastLoadError != "" {
				t.Errorf("Expected empty lastLoadError, got %q", lastLoadError)
			}
		})

		tests := []struct {
			name           string
			tomlContent    string
			wantLinesCount int
		}{
			{
				name:           "Omitted lines in TOML preserves default lines",
				tomlContent:    "[general]\ncolor_level = 1\n",
				wantLinesCount: 1,
			},
			{
				name:           "Explicit empty lines in TOML sets zero lines",
				tomlContent:    "lines = []\n",
				wantLinesCount: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tempDir := t.TempDir()
				customPath := filepath.Join(tempDir, "settings.toml")
				initConfigPath(customPath)
				lastLoadError = ""

				_ = os.WriteFile(customPath, []byte(tt.tomlContent), 0644)

				settings, err := loadSettings()
				if err != nil {
					t.Fatalf("loadSettings() unexpected error: %v", err)
				}
				if len(settings.Lines) != tt.wantLinesCount {
					t.Errorf("loadSettings() lines count = %d, want %d", len(settings.Lines), tt.wantLinesCount)
				}
				if _, statErr := os.Stat(customPath); os.IsNotExist(statErr) {
					t.Errorf("Settings file was expected to exist at %q, but does not", customPath)
				}
				if lastLoadError != "" {
					t.Errorf("Expected empty lastLoadError, got %q", lastLoadError)
				}
			})
		}
	})

	t.Run("error handling", func(t *testing.T) {
		t.Run("Fallback to default settings on invalid TOML", func(t *testing.T) {
			tempDir := t.TempDir()
			customPath := filepath.Join(tempDir, "settings.toml")
			initConfigPath(customPath)
			lastLoadError = ""

			_ = os.WriteFile(customPath, []byte("invalid = toml [ broken"), 0644)

			settings, err := loadSettings()
			if err != nil {
				t.Fatalf("loadSettings() unexpected error: %v", err)
			}
			if len(settings.Lines) != 1 {
				t.Errorf("loadSettings() lines count = %d, want 1", len(settings.Lines))
			}
			if _, statErr := os.Stat(customPath); os.IsNotExist(statErr) {
				t.Errorf("Settings file was expected to exist at %q, but does not", customPath)
			}
			if lastLoadError == "" {
				t.Errorf("Expected lastLoadError to be recorded, but got empty")
			}
		})
	})
}

func TestLoadSettings_PayloadDumpDefaults(t *testing.T) {
	tests := []struct {
		name         string
		tomlContent  string
		expectedPath string
	}{
		{
			name: "toml without payload_dump populates default path",
			tomlContent: `
lines = [["model"]]
[general]
color_level = 1
`,
			expectedPath: "~/.local/share/agystatusline/payload.json",
		},
		{
			name: "toml with payload_dump but no path populates default path",
			tomlContent: `
lines = [["model"]]
[payload_dump]
enabled = true
`,
			expectedPath: "~/.local/share/agystatusline/payload.json",
		},
		{
			name: "toml with explicit empty path populates default path",
			tomlContent: `
lines = [["model"]]
[payload_dump]
enabled = true
path = ""
`,
			expectedPath: "~/.local/share/agystatusline/payload.json",
		},
		{
			name: "toml with whitespace only path populates default path",
			tomlContent: `
lines = [["model"]]
[payload_dump]
enabled = true
path = "   "
`,
			expectedPath: "~/.local/share/agystatusline/payload.json",
		},
		{
			name: "toml with explicit custom path preserves custom path",
			tomlContent: `
lines = [["model"]]
[payload_dump]
enabled = true
path = "/custom/status.json"
`,
			expectedPath: "/custom/status.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			customPath := filepath.Join(tempDir, "settings.toml")
			initConfigPath(customPath)
			_ = os.WriteFile(customPath, []byte(tt.tomlContent), 0644)

			settings, err := loadSettings()
			if err != nil {
				t.Fatalf("loadSettings() unexpected error: %v", err)
			}

			if settings.PayloadDump.Path != tt.expectedPath {
				t.Errorf("Expected PayloadDump.Path %q, got %q", tt.expectedPath, settings.PayloadDump.Path)
			}
		})
	}
}

func TestLoadSettings_QuotaLabels(t *testing.T) {
	origPath := settingsPath
	defer func() {
		settingsPath = origPath
	}()

	tempDir := t.TempDir()
	customPath := filepath.Join(tempDir, "settings.toml")
	initConfigPath(customPath)

	tomlContent := `
lines = [["quota-all"]]

[quota]
order = ["3p-5h", "gemini-5h"]
exclude = ["3p-weekly"]

[quota.labels]
gemini-5h = "Custom G5h"
`
	if err := os.WriteFile(customPath, []byte(tomlContent), 0600); err != nil {
		t.Fatalf("Failed to write settings.toml: %v", err)
	}

	settings, err := loadSettings()
	if err != nil {
		t.Fatalf("loadSettings() unexpected error: %v", err)
	}

	if settings.Quota.Labels["gemini-5h"] != "Custom G5h" {
		t.Errorf("Expected user override 'Custom G5h' for gemini-5h, got %q", settings.Quota.Labels["gemini-5h"])
	}
	if settings.Quota.Labels["3p-5h"] != "Claude 5h" {
		t.Errorf("Expected default 'Claude 5h' for omitted 3p-5h key, got %q", settings.Quota.Labels["3p-5h"])
	}
	wantOrder := []string{"3p-5h", "gemini-5h"}
	if len(settings.Quota.Order) != len(wantOrder) || settings.Quota.Order[0] != "3p-5h" || settings.Quota.Order[1] != "gemini-5h" {
		t.Errorf("Expected Quota.Order %v, got %v", wantOrder, settings.Quota.Order)
	}
	wantExclude := []string{"3p-weekly"}
	if len(settings.Quota.Exclude) != 1 || settings.Quota.Exclude[0] != "3p-weekly" {
		t.Errorf("Expected Quota.Exclude %v, got %v", wantExclude, settings.Quota.Exclude)
	}
}
