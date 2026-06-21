package types

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

	if line0[0].Type != "model" || line0[0].Color != "brightMagenta" {
		t.Errorf("Expected first widget on line 0 to be model (brightMagenta), got Type='%s', Color='%s'", line0[0].Type, line0[0].Color)
	}
}

func TestParseStatusJSON_Quota(t *testing.T) {
	input := `{
		"quota": {
			"gemini-5h": {
				"remaining_fraction": 0.5019274,
				"reset_time": "2026-06-20T11:27:27Z",
				"reset_in_seconds": 8891
			},
			"3p-weekly": {
				"remaining_fraction": 1.0,
				"reset_time": "2026-06-27T08:58:32Z",
				"reset_in_seconds": 604756
			}
		}
	}`

	var status StatusJSON
	err := json.Unmarshal([]byte(input), &status)
	if err != nil {
		t.Fatalf("Failed to unmarshal StatusJSON with quota: %v", err)
	}

	if status.Quota == nil {
		t.Fatalf("Expected Quota map to be parsed, got nil")
	}

	g5h, ok := status.Quota["gemini-5h"]
	if !ok {
		t.Fatalf("Expected 'gemini-5h' key in Quota map")
	}
	if g5h.RemainingFraction == nil || *g5h.RemainingFraction != 0.5019274 {
		t.Errorf("Expected gemini-5h RemainingFraction 0.5019274, got %v", g5h.RemainingFraction)
	}
	if g5h.ResetTime != "2026-06-20T11:27:27Z" {
		t.Errorf("Expected gemini-5h ResetTime '2026-06-20T11:27:27Z', got '%s'", g5h.ResetTime)
	}
	if g5h.ResetInSeconds == nil || *g5h.ResetInSeconds != 8891 {
		t.Errorf("Expected gemini-5h ResetInSeconds 8891, got %v", g5h.ResetInSeconds)
	}

	p3w, ok := status.Quota["3p-weekly"]
	if !ok {
		t.Fatalf("Expected '3p-weekly' key in Quota map")
	}
	if p3w.RemainingFraction == nil || *p3w.RemainingFraction != 1.0 {
		t.Errorf("Expected 3p-weekly RemainingFraction 1.0, got %v", p3w.RemainingFraction)
	}
}

func TestParseStatusJSON_Sandbox(t *testing.T) {
	input := `{
		"sandbox": {
			"enabled": true
		}
	}`

	var status StatusJSON
	err := json.Unmarshal([]byte(input), &status)
	if err != nil {
		t.Fatalf("Failed to unmarshal StatusJSON with sandbox: %v", err)
	}

	if status.Sandbox == nil {
		t.Fatalf("Expected Sandbox info to be parsed, got nil")
	}

	if status.Sandbox.Enabled == nil || !*status.Sandbox.Enabled {
		t.Errorf("Expected Sandbox.Enabled to be true, got %v", status.Sandbox.Enabled)
	}
}
