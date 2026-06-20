package main

import (
	"encoding/json"
	"testing"
)

func TestParseStatusJSON_ModelString(t *testing.T) {
	input := `{
		"model": "Claude 3.5 Sonnet",
		"session_id": "test-session",
		"cwd": "/path/to/project"
	}`

	var status StatusJSON
	err := json.Unmarshal([]byte(input), &status)
	if err != nil {
		t.Fatalf("Failed to unmarshal StatusJSON: %v", err)
	}

	if status.SessionID != "test-session" {
		t.Errorf("Expected SessionID 'test-session', got '%s'", status.SessionID)
	}

	if status.Model.ID != "Claude 3.5 Sonnet" || status.Model.DisplayName != "Claude 3.5 Sonnet" {
		t.Errorf("Expected Model string to be parsed into ID/DisplayName, got ID='%s', DisplayName='%s'", status.Model.ID, status.Model.DisplayName)
	}
}

func TestParseStatusJSON_ModelObject(t *testing.T) {
	input := `{
		"model": {
			"id": "claude-3-5-sonnet-20241022",
			"display_name": "Claude 3.5 Sonnet"
		}
	}`

	var status StatusJSON
	err := json.Unmarshal([]byte(input), &status)
	if err != nil {
		t.Fatalf("Failed to unmarshal StatusJSON: %v", err)
	}

	if status.Model.ID != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected Model.ID 'claude-3-5-sonnet-20241022', got '%s'", status.Model.ID)
	}
	if status.Model.DisplayName != "Claude 3.5 Sonnet" {
		t.Errorf("Expected Model.DisplayName 'Claude 3.5 Sonnet', got '%s'", status.Model.DisplayName)
	}
}

func TestDefaultSettings(t *testing.T) {
	settings := DefaultSettings()

	if settings.Version != 3 {
		t.Errorf("Expected default settings version 3, got %d", settings.Version)
	}

	if len(settings.Lines) != 3 {
		t.Errorf("Expected 3 default lines, got %d", len(settings.Lines))
	}

	// Verify default widgets on line 0
	line0 := settings.Lines[0]
	if len(line0) == 0 {
		t.Fatalf("Expected widgets on line 0, got none")
	}

	if line0[0].Type != "model" || line0[0].Color != "cyan" {
		t.Errorf("Expected first widget on line 0 to be model (cyan), got Type='%s', Color='%s'", line0[0].Type, line0[0].Color)
	}
}
