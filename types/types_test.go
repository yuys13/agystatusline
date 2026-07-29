package types

import (
	"encoding/json"
	"testing"
)

func TestParseStatusJSON(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		verify func(t *testing.T, status StatusJSON)
	}{
		{
			name: "model string parsing",
			input: `{
				"model": "Claude 3.5 Sonnet",
				"session_id": "test-session",
				"cwd": "/path/to/project"
			}`,
			verify: func(t *testing.T, status StatusJSON) {
				if status.SessionID != "test-session" {
					t.Errorf("Expected SessionID 'test-session', got '%s'", status.SessionID)
				}
				if status.Model.ID != "Claude 3.5 Sonnet" || status.Model.DisplayName != "Claude 3.5 Sonnet" {
					t.Errorf("Expected Model string to be parsed into ID/DisplayName, got ID='%s', DisplayName='%s'", status.Model.ID, status.Model.DisplayName)
				}
			},
		},
		{
			name: "model object parsing",
			input: `{
				"model": {
					"id": "claude-3-5-sonnet-20241022",
					"display_name": "Claude 3.5 Sonnet"
				}
			}`,
			verify: func(t *testing.T, status StatusJSON) {
				if status.Model.ID != "claude-3-5-sonnet-20241022" {
					t.Errorf("Expected Model.ID 'claude-3-5-sonnet-20241022', got '%s'", status.Model.ID)
				}
				if status.Model.DisplayName != "Claude 3.5 Sonnet" {
					t.Errorf("Expected Model.DisplayName 'Claude 3.5 Sonnet', got '%s'", status.Model.DisplayName)
				}
			},
		},
		{
			name: "quota map parsing",
			input: `{
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
			}`,
			verify: func(t *testing.T, status StatusJSON) {
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
			},
		},
		{
			name: "sandbox info parsing",
			input: `{
				"sandbox": {
					"enabled": true
				}
			}`,
			verify: func(t *testing.T, status StatusJSON) {
				if status.Sandbox == nil {
					t.Fatalf("Expected Sandbox info to be parsed, got nil")
				}
				if status.Sandbox.Enabled == nil || !*status.Sandbox.Enabled {
					t.Errorf("Expected Sandbox.Enabled to be true, got %v", status.Sandbox.Enabled)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var status StatusJSON
			err := json.Unmarshal([]byte(tc.input), &status)
			if err != nil {
				t.Fatalf("Failed to unmarshal StatusJSON: %v", err)
			}
			tc.verify(t, status)
		})
	}
}

func TestDefaultSettings(t *testing.T) {
	settings := DefaultSettings()

	t.Run("version and line count", func(t *testing.T) {
		if settings.Version != 3 {
			t.Errorf("Expected default settings version 3, got %d", settings.Version)
		}

		if len(settings.Lines) != 3 {
			t.Errorf("Expected 3 default lines, got %d", len(settings.Lines))
		}
	})

	t.Run("default widgets on line 0", func(t *testing.T) {
		line0 := settings.Lines[0]
		if len(line0) != 7 {
			t.Fatalf("Expected 7 widgets on line 0, got %d", len(line0))
		}

		widgetTests := []struct {
			name          string
			index         int
			expectedType  string
			expectedColor string
		}{
			{"first widget agent-state", 0, "agent-state", "brightGreen"},
			{"second widget model", 1, "model", "brightMagenta"},
			{"seventh widget sandbox", 6, "sandbox", "yellow"},
		}

		for _, tc := range widgetTests {
			t.Run(tc.name, func(t *testing.T) {
				w := line0[tc.index]
				if w.Type != tc.expectedType || w.Color != tc.expectedColor {
					t.Errorf("Expected widget at index %d to be %s (%s), got Type='%s', Color='%s'",
						tc.index, tc.expectedType, tc.expectedColor, w.Type, w.Color)
				}
			})
		}
	})

	t.Run("powerline theme", func(t *testing.T) {
		if settings.Powerline.Theme != "nord-aurora" {
			t.Errorf("Expected default powerline theme 'nord-aurora', got '%s'", settings.Powerline.Theme)
		}
	})
}

func TestModelInfo_UnmarshalJSON_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"array input", `["invalid", "array"]`},
		{"numeric input", `12345`},
		{"boolean input", `true`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var m ModelInfo
			err := m.UnmarshalJSON([]byte(tc.input))
			if err == nil {
				t.Errorf("Expected error for invalid ModelInfo input '%s', got nil", tc.input)
			}
		})
	}
}

func TestContextUsage_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedInput  float64
		expectedOutput float64
		expectErr      bool
	}{
		{
			name:          "valid numeric context usage",
			input:         `1234.5`,
			expectedInput: 1234.5,
			expectErr:     false,
		},
		{
			name:           "valid object context usage",
			input:          `{"input_tokens": 100, "output_tokens": 50}`,
			expectedInput:  100,
			expectedOutput: 50,
			expectErr:      false,
		},
		{
			name:      "invalid array input",
			input:     `["invalid"]`,
			expectErr: true,
		},
		{
			name:      "invalid string input",
			input:     `"string"`,
			expectErr: true,
		},
		{
			name:      "invalid boolean input",
			input:     `true`,
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var c ContextUsage
			err := c.UnmarshalJSON([]byte(tc.input))
			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error for invalid ContextUsage input '%s', got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error unmarshaling ContextUsage: %v", err)
			}
			if c.InputTokens != tc.expectedInput || c.OutputTokens != tc.expectedOutput {
				t.Errorf("Got InputTokens=%f, OutputTokens=%f; expected %f, %f", c.InputTokens, c.OutputTokens, tc.expectedInput, tc.expectedOutput)
			}
		})
	}
}

func TestRenderContext_Getters(t *testing.T) {
	tests := []struct {
		name             string
		ctx              RenderContext
		expectedCwd      string
		expectedWorkCur  string
		expectedWorkProj string
	}{
		{
			name: "with workspace and cwd",
			ctx: RenderContext{
				Data: StatusJSON{
					CWD: "/cwd/path",
					Workspace: &WorkspaceInfo{
						CurrentDir: "/workspace/current",
						ProjectDir: "/workspace/project",
					},
				},
			},
			expectedCwd:      "/cwd/path",
			expectedWorkCur:  "/workspace/current",
			expectedWorkProj: "/workspace/project",
		},
		{
			name: "nil workspace",
			ctx: RenderContext{
				Data: StatusJSON{
					CWD: "/cwd/path",
				},
			},
			expectedCwd:      "/cwd/path",
			expectedWorkCur:  "",
			expectedWorkProj: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.ctx.GetCwd() != tc.expectedCwd {
				t.Errorf("Expected GetCwd() '%s', got '%s'", tc.expectedCwd, tc.ctx.GetCwd())
			}
			if tc.ctx.GetWorkspaceCurrentDir() != tc.expectedWorkCur {
				t.Errorf("Expected GetWorkspaceCurrentDir() '%s', got '%s'", tc.expectedWorkCur, tc.ctx.GetWorkspaceCurrentDir())
			}
			if tc.ctx.GetWorkspaceProjectDir() != tc.expectedWorkProj {
				t.Errorf("Expected GetWorkspaceProjectDir() '%s', got '%s'", tc.expectedWorkProj, tc.ctx.GetWorkspaceProjectDir())
			}
		})
	}
}
