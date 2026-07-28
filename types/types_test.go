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
	if len(line0) != 7 {
		t.Fatalf("Expected 7 widgets on line 0, got %d", len(line0))
	}

	if line0[0].Type != "agent-state" || line0[0].Color != "brightGreen" {
		t.Errorf("Expected first widget on line 0 to be agent-state (brightGreen), got Type='%s', Color='%s'", line0[0].Type, line0[0].Color)
	}

	if line0[1].Type != "model" || line0[1].Color != "brightMagenta" {
		t.Errorf("Expected second widget on line 0 to be model (brightMagenta), got Type='%s', Color='%s'", line0[1].Type, line0[1].Color)
	}

	if line0[6].Type != "sandbox" || line0[6].Color != "yellow" {
		t.Errorf("Expected seventh widget on line 0 to be sandbox (yellow), got Type='%s', Color='%s'", line0[6].Type, line0[6].Color)
	}

	if settings.Powerline.Theme != "nord-aurora" {
		t.Errorf("Expected default powerline theme 'nord-aurora', got '%s'", settings.Powerline.Theme)
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

func TestModelInfo_UnmarshalJSON_Invalid(t *testing.T) {
	invalidInputs := []string{
		`["invalid", "array"]`,
		`12345`,
		`true`,
	}

	for _, input := range invalidInputs {
		var m ModelInfo
		err := m.UnmarshalJSON([]byte(input))
		if err == nil {
			t.Errorf("Expected error for invalid ModelInfo input '%s', got nil", input)
		}
	}
}

func TestContextUsage_UnmarshalJSON(t *testing.T) {
	// Valid numeric context usage
	var c1 ContextUsage
	if err := c1.UnmarshalJSON([]byte(`1234.5`)); err != nil {
		t.Fatalf("Failed to unmarshal numeric ContextUsage: %v", err)
	}
	if c1.InputTokens != 1234.5 {
		t.Errorf("Expected InputTokens 1234.5, got %f", c1.InputTokens)
	}

	// Valid object context usage
	var c2 ContextUsage
	if err := c2.UnmarshalJSON([]byte(`{"input_tokens": 100, "output_tokens": 50}`)); err != nil {
		t.Fatalf("Failed to unmarshal object ContextUsage: %v", err)
	}
	if c2.InputTokens != 100 || c2.OutputTokens != 50 {
		t.Errorf("Expected InputTokens 100 and OutputTokens 50, got InputTokens=%f OutputTokens=%f", c2.InputTokens, c2.OutputTokens)
	}

	// Invalid inputs
	invalidInputs := []string{
		`["invalid"]`,
		`"string"`,
		`true`,
	}

	for _, input := range invalidInputs {
		var c ContextUsage
		err := c.UnmarshalJSON([]byte(input))
		if err == nil {
			t.Errorf("Expected error for invalid ContextUsage input '%s', got nil", input)
		}
	}
}

func TestRenderContext_Getters(t *testing.T) {
	// 1. With Workspace and CWD
	c1 := RenderContext{
		Data: StatusJSON{
			CWD: "/cwd/path",
			Workspace: &WorkspaceInfo{
				CurrentDir: "/workspace/current",
				ProjectDir: "/workspace/project",
			},
		},
	}
	if c1.GetCwd() != "/cwd/path" {
		t.Errorf("Expected GetCwd() '/cwd/path', got '%s'", c1.GetCwd())
	}
	if c1.GetWorkspaceCurrentDir() != "/workspace/current" {
		t.Errorf("Expected GetWorkspaceCurrentDir() '/workspace/current', got '%s'", c1.GetWorkspaceCurrentDir())
	}
	if c1.GetWorkspaceProjectDir() != "/workspace/project" {
		t.Errorf("Expected GetWorkspaceProjectDir() '/workspace/project', got '%s'", c1.GetWorkspaceProjectDir())
	}

	// 2. Nil Workspace
	c2 := RenderContext{
		Data: StatusJSON{
			CWD: "/cwd/path",
		},
	}
	if c2.GetCwd() != "/cwd/path" {
		t.Errorf("Expected GetCwd() '/cwd/path', got '%s'", c2.GetCwd())
	}
	if c2.GetWorkspaceCurrentDir() != "" {
		t.Errorf("Expected empty GetWorkspaceCurrentDir(), got '%s'", c2.GetWorkspaceCurrentDir())
	}
	if c2.GetWorkspaceProjectDir() != "" {
		t.Errorf("Expected empty GetWorkspaceProjectDir(), got '%s'", c2.GetWorkspaceProjectDir())
	}
}
