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
	tests := []struct {
		name           string
		setup          func(path string)
		wantLinesCount int
		expectCreated  bool
		expectErr      bool
	}{
		{
			name: "Create default settings when file does not exist",
			setup: func(path string) {
				// Ensure file does not exist initially
			},
			wantLinesCount: 1,
			expectCreated:  true,
			expectErr:      false,
		},
		{
			name: "Fallback to default settings on invalid TOML",
			setup: func(path string) {
				_ = os.WriteFile(path, []byte("invalid = toml [ broken"), 0644)
			},
			wantLinesCount: 1,
			expectCreated:  true,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			customPath := filepath.Join(tempDir, "settings.toml")
			initConfigPath(customPath)
			lastLoadError = "" // reset error state

			tt.setup(customPath)

			settings, err := loadSettings()
			if err != nil {
				t.Fatalf("loadSettings() unexpected error: %v", err)
			}
			if len(settings.Lines) != tt.wantLinesCount {
				t.Errorf("loadSettings() lines count = %d, want %d", len(settings.Lines), tt.wantLinesCount)
			}
			if tt.expectCreated {
				if _, err := os.Stat(customPath); os.IsNotExist(err) {
					t.Errorf("Settings file was not created on disk at %q", customPath)
				}
			}
			if tt.expectErr && lastLoadError == "" {
				t.Errorf("Expected lastLoadError to be recorded, but got empty")
			}
			if !tt.expectErr && lastLoadError != "" {
				t.Errorf("Expected empty lastLoadError, got %q", lastLoadError)
			}
		})
	}
}
