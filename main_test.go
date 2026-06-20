package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestParseConfigArg(t *testing.T) {
	// Dummy initial path
	settingsPath = "default_settings.json"

	args := []string{"./agystatusline", "--config", "/tmp/custom.json"}
	path, remaining := parseConfigArg(args)

	if path != "/tmp/custom.json" {
		t.Errorf("Expected custom path '/tmp/custom.json', got '%s'", path)
	}

	if len(remaining) != 1 || remaining[0] != "./agystatusline" {
		t.Errorf("Expected remaining args to contain only program name, got %v", remaining)
	}
}

func TestLoadSettings_CreateDefault(t *testing.T) {
	tempDir := t.TempDir()
	customPath := filepath.Join(tempDir, "settings.json")
	initConfigPath(customPath)

	// File should not exist initially
	if _, err := os.Stat(customPath); !os.IsNotExist(err) {
		t.Fatalf("Settings file should not exist yet")
	}

	settings, err := loadSettings()
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	if settings.Version != 3 {
		t.Errorf("Expected loaded settings version 3, got %d", settings.Version)
	}

	// File should now be written to disk
	if _, err := os.Stat(customPath); os.IsNotExist(err) {
		t.Errorf("Settings file was not created on disk")
	}
}

func TestLoadSettings_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	customPath := filepath.Join(tempDir, "settings.json")
	initConfigPath(customPath)

	// Write invalid JSON
	err := os.WriteFile(customPath, []byte("{invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid settings: %v", err)
	}

	settings, err := loadSettings()
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	// Should fallback to default settings
	if len(settings.Lines) != 3 {
		t.Errorf("Expected 3 fallback lines, got %d", len(settings.Lines))
	}

	// Error should be stored
	if lastLoadError == "" {
		t.Errorf("Expected load error to be recorded, got empty")
	}
}

func TestUpgradeLegacyWidgetTypes(t *testing.T) {
	lines := [][]types.WidgetItem{
		{
			{Type: "git-pr"},
		},
	}
	upgraded := upgradeLegacyWidgetTypes(lines)
	if upgraded[0][0].Type != "git-review" {
		t.Errorf("Expected 'git-pr' to be upgraded to 'git-review', got '%s'", upgraded[0][0].Type)
	}
}
