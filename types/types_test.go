package types

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
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
					t.Errorf("Expected SessionID %q, got %q", "test-session", status.SessionID)
				}
				if status.Model.ID != "Claude 3.5 Sonnet" || status.Model.DisplayName != "Claude 3.5 Sonnet" {
					t.Errorf("Expected Model string to be parsed into ID/DisplayName, got ID=%q, DisplayName=%q", status.Model.ID, status.Model.DisplayName)
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
					t.Errorf("Expected Model.ID %q, got %q", "claude-3-5-sonnet-20241022", status.Model.ID)
				}
				if status.Model.DisplayName != "Claude 3.5 Sonnet" {
					t.Errorf("Expected Model.DisplayName %q, got %q", "Claude 3.5 Sonnet", status.Model.DisplayName)
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
					t.Errorf("Expected gemini-5h ResetTime %q, got %q", "2026-06-20T11:27:27Z", g5h.ResetTime)
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
		{
			name: "vcs and workspace parsing",
			input: `{
				"cwd": "/home/user/project",
				"workspace": {
					"current_dir": "/home/user/project",
					"project_dir": "/home/user/project"
				},
				"vcs": {
					"type": "git",
					"branch": "main",
					"dirty": false
				},
				"agent_state": "THINKING",
				"artifact_count": 3,
				"task_count": 2
			}`,
			verify: func(t *testing.T, status StatusJSON) {
				if status.CWD != "/home/user/project" {
					t.Errorf("Expected CWD /home/user/project, got %q", status.CWD)
				}
				if status.Workspace == nil || status.Workspace.ProjectDir != "/home/user/project" {
					t.Errorf("Expected Workspace.ProjectDir /home/user/project, got %v", status.Workspace)
				}
				if status.VCS == nil || status.VCS.Branch != "main" {
					t.Errorf("Expected VCS.Branch main, got %v", status.VCS)
				}
				if status.VCS.Dirty == nil || *status.VCS.Dirty {
					t.Errorf("Expected VCS.Dirty false, got %v", status.VCS.Dirty)
				}
				if status.AgentState != "THINKING" {
					t.Errorf("Expected AgentState THINKING, got %q", status.AgentState)
				}
				if status.ArtifactCount == nil || *status.ArtifactCount != 3 {
					t.Errorf("Expected ArtifactCount 3, got %v", status.ArtifactCount)
				}
				if status.TaskCount == nil || *status.TaskCount != 2 {
					t.Errorf("Expected TaskCount 2, got %v", status.TaskCount)
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
				t.Errorf("Expected error for invalid ModelInfo input %q, got nil", tc.input)
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
					t.Errorf("Expected error for invalid ContextUsage input %q, got nil", tc.input)
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
				t.Errorf("Expected GetCwd() %q, got %q", tc.expectedCwd, tc.ctx.GetCwd())
			}
			if tc.ctx.GetWorkspaceCurrentDir() != tc.expectedWorkCur {
				t.Errorf("Expected GetWorkspaceCurrentDir() %q, got %q", tc.expectedWorkCur, tc.ctx.GetWorkspaceCurrentDir())
			}
			if tc.ctx.GetWorkspaceProjectDir() != tc.expectedWorkProj {
				t.Errorf("Expected GetWorkspaceProjectDir() %q, got %q", tc.expectedWorkProj, tc.ctx.GetWorkspaceProjectDir())
			}
		})
	}
}

func TestDefaultSettings(t *testing.T) {
	settings := DefaultSettings()

	t.Run("line count", func(t *testing.T) {
		if len(settings.Lines) != 1 {
			t.Fatalf("Expected 1 default line, got %d", len(settings.Lines))
		}
	})

	t.Run("default widgets on line 0", func(t *testing.T) {
		line0 := settings.Lines[0]
		expectedTypes := []string{
			"agent-state",
			"model",
			"context-bar",
			"artifacts",
			"subagents",
			"tasks",
			"sandbox",
		}

		if len(line0) != len(expectedTypes) {
			t.Fatalf("Expected %d widgets on line 0, got %d", len(expectedTypes), len(line0))
		}

		for i, expectedType := range expectedTypes {
			if line0[i].Type != expectedType {
				t.Errorf("Expected widget at index %d to be %q, got %q", i, expectedType, line0[i].Type)
			}
		}
	})

	t.Run("powerline defaults", func(t *testing.T) {
		if settings.Powerline.Enabled != false {
			t.Errorf("Expected Powerline.Enabled false, got %v", settings.Powerline.Enabled)
		}
		if settings.Powerline.Theme != "nord-aurora" {
			t.Errorf("Expected default powerline theme %q, got %q", "nord-aurora", settings.Powerline.Theme)
		}
		if settings.Powerline.Separator != "\uE0B0" {
			t.Errorf("Expected default powerline separator %q, got %q", "\uE0B0", settings.Powerline.Separator)
		}
		if settings.Powerline.StartCaps != "" {
			t.Errorf("Expected default powerline start_caps %q, got %q", "", settings.Powerline.StartCaps)
		}
		if settings.Powerline.EndCaps != "" {
			t.Errorf("Expected default powerline end_caps %q, got %q", "", settings.Powerline.EndCaps)
		}
	})

	t.Run("general defaults", func(t *testing.T) {
		if settings.General.ColorLevel != 1 {
			t.Errorf("Expected General.ColorLevel 1, got %d", settings.General.ColorLevel)
		}
		if settings.General.GitCacheTTL != 5 {
			t.Errorf("Expected General.GitCacheTTL 5, got %d", settings.General.GitCacheTTL)
		}
		if settings.General.Separator != " · " {
			t.Errorf("Expected General.Separator %q, got %q", " · ", settings.General.Separator)
		}
		if settings.General.Padding != "" {
			t.Errorf("Expected General.Padding %q, got %q", "", settings.General.Padding)
		}
		if settings.General.Minimalist != false {
			t.Errorf("Expected General.Minimalist false, got %v", settings.General.Minimalist)
		}
	})
}

func TestWidgetItem_UnmarshalTOML(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected WidgetItem
		hasErr   bool
	}{
		{
			name:     "plain string widget",
			input:    "model",
			expected: WidgetItem{Type: "model"},
			hasErr:   false,
		},
		{
			name:     "quota shorthand",
			input:    "quota:gemini-5h",
			expected: WidgetItem{Type: "quota", Key: "gemini-5h"},
			hasErr:   false,
		},
		{
			name:     "quota-bar shorthand",
			input:    "quota-bar:gemini-weekly",
			expected: WidgetItem{Type: "quota-bar", Key: "gemini-weekly"},
			hasErr:   false,
		},
		{
			name:     "custom-text shorthand",
			input:    "custom-text:PROD",
			expected: WidgetItem{Type: "custom-text", Text: "PROD"},
			hasErr:   false,
		},
		{
			name:     "custom alias shorthand",
			input:    "custom:DEV",
			expected: WidgetItem{Type: "custom-text", Text: "DEV"},
			hasErr:   false,
		},
		{
			name:     "git-branch symbol shorthand with trailing space",
			input:    "git-branch: ",
			expected: WidgetItem{Type: "git-branch", Symbol: " "},
			hasErr:   false,
		},
		{
			name:     "git-branch custom glyph shorthand",
			input:    "git-branch: ",
			expected: WidgetItem{Type: "git-branch", Symbol: " "},
			hasErr:   false,
		},
		{
			name:     "generic unknown shorthand with colon",
			input:    "other-widget:my-param",
			expected: WidgetItem{Type: "other-widget", Key: "my-param"},
			hasErr:   false,
		},
		{
			name:     "plain string byte slice",
			input:    []byte("artifacts"),
			expected: WidgetItem{Type: "artifacts"},
			hasErr:   false,
		},
		{
			name: "map with type and key",
			input: map[string]any{
				"type": "quota-bar",
				"key":  "gemini-weekly",
			},
			expected: WidgetItem{Type: "quota-bar", Key: "gemini-weekly"},
			hasErr:   false,
		},
		{
			name: "map with all fields",
			input: map[string]any{
				"type":   "custom-text",
				"key":    "k1",
				"text":   "MY_TEXT",
				"symbol": "sym",
				"color":  "brightRed",
				"raw":    true,
			},
			expected: WidgetItem{
				Type:   "custom-text",
				Key:    "k1",
				Text:   "MY_TEXT",
				Symbol: "sym",
				Color:  "brightRed",
				Raw:    true,
			},
			hasErr: false,
		},
		{
			name: "map with custom alias",
			input: map[string]any{
				"type": "custom",
				"text": "ALIASED",
			},
			expected: WidgetItem{
				Type: "custom-text",
				Text: "ALIASED",
			},
			hasErr: false,
		},
		{
			name:     "invalid integer input",
			input:    42,
			expected: WidgetItem{},
			hasErr:   true,
		},
		{
			name:     "invalid boolean input",
			input:    true,
			expected: WidgetItem{},
			hasErr:   true,
		},
		{
			name:     "invalid float input",
			input:    3.14,
			expected: WidgetItem{},
			hasErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var w WidgetItem
			err := w.UnmarshalTOML(tc.input)
			if tc.hasErr {
				if err == nil {
					t.Errorf("Expected error for input %v, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !reflect.DeepEqual(w, tc.expected) {
				t.Errorf("Expected %+v, got %+v", tc.expected, w)
			}
		})
	}
}

func TestWidgetItem_UnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected WidgetItem
	}{
		{
			name:     "model widget",
			input:    "model",
			expected: WidgetItem{Type: "model"},
		},
		{
			name:     "quota shorthand",
			input:    "quota:3p-5h",
			expected: WidgetItem{Type: "quota", Key: "3p-5h"},
		},
		{
			name:     "custom text shorthand",
			input:    "custom-text:STAGING",
			expected: WidgetItem{Type: "custom-text", Text: "STAGING"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var w WidgetItem
			err := w.UnmarshalText([]byte(tc.input))
			if err != nil {
				t.Fatalf("Unexpected error unmarshaling text: %v", err)
			}
			if !reflect.DeepEqual(w, tc.expected) {
				t.Errorf("Expected %+v, got %+v", tc.expected, w)
			}
		})
	}
}

func TestWidgetItem_MarshalTOML(t *testing.T) {
	tests := []struct {
		name     string
		widget   WidgetItem
		expected any
	}{
		{
			name:     "plain widget",
			widget:   WidgetItem{Type: "model"},
			expected: "model",
		},
		{
			name:     "quota shorthand",
			widget:   WidgetItem{Type: "quota", Key: "gemini-5h"},
			expected: "quota:gemini-5h",
		},
		{
			name:     "quota-bar shorthand",
			widget:   WidgetItem{Type: "quota-bar", Key: "gemini-weekly"},
			expected: "quota-bar:gemini-weekly",
		},
		{
			name:     "custom-text shorthand",
			widget:   WidgetItem{Type: "custom-text", Text: "PROD"},
			expected: "custom-text:PROD",
		},
		{
			name:     "git-branch shorthand",
			widget:   WidgetItem{Type: "git-branch", Symbol: " "},
			expected: "git-branch: ",
		},
		{
			name: "custom colored widget marshaled as map",
			widget: WidgetItem{
				Type:  "custom-text",
				Text:  "PROD",
				Color: "brightRed",
			},
			expected: map[string]any{
				"type":  "custom-text",
				"text":  "PROD",
				"color": "brightRed",
			},
		},
		{
			name: "raw widget marshaled as map",
			widget: WidgetItem{
				Type: "model",
				Raw:  true,
			},
			expected: map[string]any{
				"type": "model",
				"raw":  true,
			},
		},
		{
			name: "widget with multiple properties marshaled as map",
			widget: WidgetItem{
				Type:   "custom-text",
				Key:    "k1",
				Text:   "txt",
				Symbol: "sym",
			},
			expected: map[string]any{
				"type":   "custom-text",
				"key":    "k1",
				"text":   "txt",
				"symbol": "sym",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val, err := tc.widget.MarshalTOML()
			if err != nil {
				t.Fatalf("MarshalTOML returned error: %v", err)
			}
			if !reflect.DeepEqual(val, tc.expected) {
				t.Errorf("Expected marshal result %v (%T), got %v (%T)", tc.expected, tc.expected, val, val)
			}
		})
	}
}

func TestSettings_TOML_Integration(t *testing.T) {
	t.Run("full TOML document unmarshaling", func(t *testing.T) {
		tomlData := `
lines = [
  ["agent-state", "model", "context-bar", "quota:gemini-5h", "git-branch: "],
  ["artifacts", "subagents", "tasks", "sandbox", { type = "custom-text", text = "PROD", color = "brightRed", raw = true }]
]

[powerline]
enabled = true
theme = "tokyonight"
separator = ""
start_caps = ""
end_caps = ""

[general]
color_level = 3
git_cache_ttl = 10
separator = " | "
padding = " "
minimalist = true
`
		var s Settings
		err := toml.Unmarshal([]byte(tomlData), &s)
		if err != nil {
			t.Fatalf("Failed to unmarshal settings TOML: %v", err)
		}

		if len(s.Lines) != 2 {
			t.Fatalf("Expected 2 lines, got %d", len(s.Lines))
		}

		// Check Line 0
		if len(s.Lines[0]) != 5 {
			t.Fatalf("Expected 5 widgets in line 0, got %d", len(s.Lines[0]))
		}
		if s.Lines[0][0].Type != "agent-state" {
			t.Errorf("Expected line 0 widget 0 agent-state, got %q", s.Lines[0][0].Type)
		}
		if s.Lines[0][3].Type != "quota" || s.Lines[0][3].Key != "gemini-5h" {
			t.Errorf("Expected line 0 widget 3 quota:gemini-5h, got %+v", s.Lines[0][3])
		}
		if s.Lines[0][4].Type != "git-branch" || s.Lines[0][4].Symbol != " " {
			t.Errorf("Expected line 0 widget 4 git-branch with space symbol, got %+v", s.Lines[0][4])
		}

		// Check Line 1
		if len(s.Lines[1]) != 5 {
			t.Fatalf("Expected 5 widgets in line 1, got %d", len(s.Lines[1]))
		}
		customWidget := s.Lines[1][4]
		if customWidget.Type != "custom-text" || customWidget.Text != "PROD" || customWidget.Color != "brightRed" || !customWidget.Raw {
			t.Errorf("Expected line 1 widget 4 custom table, got %+v", customWidget)
		}

		// Check Powerline
		if !s.Powerline.Enabled {
			t.Errorf("Expected Powerline.Enabled true")
		}
		if s.Powerline.Theme != "tokyonight" {
			t.Errorf("Expected Powerline.Theme tokyonight, got %q", s.Powerline.Theme)
		}
		if s.Powerline.Separator != "" {
			t.Errorf("Expected Powerline.Separator , got %q", s.Powerline.Separator)
		}
		if s.Powerline.StartCaps != "" || s.Powerline.EndCaps != "" {
			t.Errorf("Expected StartCaps , EndCaps , got %q, %q", s.Powerline.StartCaps, s.Powerline.EndCaps)
		}

		// Check General
		if s.General.ColorLevel != 3 {
			t.Errorf("Expected ColorLevel 3, got %d", s.General.ColorLevel)
		}
		if s.General.GitCacheTTL != 10 {
			t.Errorf("Expected GitCacheTTL 10, got %d", s.General.GitCacheTTL)
		}
		if s.General.Separator != " | " {
			t.Errorf("Expected Separator ' | ', got %q", s.General.Separator)
		}
		if s.General.Padding != " " {
			t.Errorf("Expected Padding ' ', got %q", s.General.Padding)
		}
		if !s.General.Minimalist {
			t.Errorf("Expected Minimalist true")
		}
	})

	t.Run("default settings marshaling and unmarshaling", func(t *testing.T) {
		defaults := DefaultSettings()
		bytes, err := toml.Marshal(defaults)
		if err != nil {
			t.Fatalf("Failed to marshal DefaultSettings: %v", err)
		}

		var restored Settings
		err = toml.Unmarshal(bytes, &restored)
		if err != nil {
			t.Fatalf("Failed to unmarshal marshaled DefaultSettings: %v", err)
		}

		if len(restored.Lines) != 1 || len(restored.Lines[0]) != 7 {
			t.Fatalf("Expected restored settings to have 1 line with 7 widgets, got %+v", restored.Lines)
		}
		if restored.Powerline.Theme != "nord-aurora" {
			t.Errorf("Expected restored Powerline.Theme nord-aurora, got %q", restored.Powerline.Theme)
		}
		if restored.General.ColorLevel != 1 {
			t.Errorf("Expected restored General.ColorLevel 1, got %d", restored.General.ColorLevel)
		}
	})
}

func TestSettings_TOML_RoundTrip_ComplexCombinations(t *testing.T) {
	tests := []struct {
		name     string
		settings Settings
	}{
		{
			name:     "default settings",
			settings: DefaultSettings(),
		},
		{
			name: "multi-line with mixed shorthands, tables, and unicode",
			settings: Settings{
				Lines: [][]WidgetItem{
					{
						{Type: "agent-state"},
						{Type: "model"},
						{Type: "context-bar"},
						{Type: "artifacts"},
						{Type: "subagents"},
						{Type: "tasks"},
						{Type: "sandbox"},
					},
					{
						{Type: "quota", Key: "gemini-5h"},
						{Type: "quota-bar", Key: "3p-weekly"},
						{Type: "git-branch", Symbol: " "},
						{Type: "git-changes"},
						{Type: "custom-text", Text: "PROD"},
					},
					{
						{Type: "custom-text", Text: "http://example.com:8080/api", Color: "yellow", Raw: true},
						{Type: "quota", Key: "custom:key:with:colons", Color: "#ff00ff"},
						{Type: "git-branch", Symbol: "🌿 ", Color: "green"},
						{Type: "model", Raw: true},
					},
				},
				Powerline: PowerlineConfig{
					Enabled:   true,
					Theme:     "tokyonight",
					Separator: "",
					StartCaps: "",
					EndCaps:   "",
				},
				General: GeneralConfig{
					ColorLevel:  3,
					GitCacheTTL: 60,
					Separator:   " // ",
					Padding:     "  ",
					Minimalist:  true,
				},
			},
		},
		{
			name: "all 12 widget types plain",
			settings: Settings{
				Lines: [][]WidgetItem{
					{
						{Type: "agent-state"},
						{Type: "model"},
						{Type: "context-bar"},
						{Type: "artifacts"},
						{Type: "subagents"},
						{Type: "tasks"},
						{Type: "sandbox"},
						{Type: "quota"},
						{Type: "quota-bar"},
						{Type: "git-branch"},
						{Type: "git-changes"},
						{Type: "custom-text"},
					},
				},
				Powerline: PowerlineConfig{
					Enabled:   false,
					Theme:     "default",
					Separator: "",
				},
				General: GeneralConfig{
					ColorLevel:  0,
					GitCacheTTL: 0,
					Separator:   "",
					Padding:     "",
					Minimalist:  false,
				},
			},
		},
		{
			name: "empty lines",
			settings: Settings{
				Lines: [][]WidgetItem{},
			},
		},
		{
			name: "single empty line",
			settings: Settings{
				Lines: [][]WidgetItem{{}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			marshaled, err := toml.Marshal(tc.settings)
			if err != nil {
				t.Fatalf("Failed to marshal settings: %v", err)
			}

			var restored Settings
			err = toml.Unmarshal(marshaled, &restored)
			if err != nil {
				t.Fatalf("Failed to unmarshal settings: %v\nTOML:\n%s", err, string(marshaled))
			}

			// Marshal again to ensure idempotency and canonical stability
			remarshaled, err := toml.Marshal(restored)
			if err != nil {
				t.Fatalf("Failed to remarshal restored settings: %v", err)
			}

			var restoredAgain Settings
			err = toml.Unmarshal(remarshaled, &restoredAgain)
			if err != nil {
				t.Fatalf("Failed to unmarshal remarshaled settings: %v\nTOML:\n%s", err, string(remarshaled))
			}

			if !reflect.DeepEqual(restored, restoredAgain) {
				t.Errorf("Round-trip serialization not idempotent!\nFirst restored: %+v\nSecond restored: %+v", restored, restoredAgain)
			}

			// Validate deep equality with original when lines are non-empty
			if len(tc.settings.Lines) > 0 && len(tc.settings.Lines[0]) > 0 {
				if !reflect.DeepEqual(tc.settings, restored) {
					t.Errorf("Restored settings did not match original!\nExpected: %+v\nGot:      %+v\nTOML:\n%s", tc.settings, restored, string(marshaled))
				}
			}
		})
	}
}

func TestStatusJSON_AntigravityAndClaudeCode_Payloads(t *testing.T) {
	tests := []struct {
		name     string
		jsonText string
		verify   func(t *testing.T, s StatusJSON)
	}{
		{
			name: "antigravity full telemetry with all fields",
			jsonText: `{
				"hook_event_name": "status",
				"session_id": "antigravity-session-001",
				"transcript_path": "/path/to/transcript.jsonl",
				"cwd": "/workspace/app",
				"model": {
					"id": "gemini-1.5-pro",
					"display_name": "Gemini 1.5 Pro"
				},
				"workspace": {
					"current_dir": "/workspace/app",
					"project_dir": "/workspace"
				},
				"version": "1.2.0",
				"output_style": {
					"name": "full"
				},
				"effort": {
					"level": "medium"
				},
				"cost": {
					"total_cost_usd": 0.1234,
					"total_duration_ms": 45000.0,
					"total_api_duration_ms": 32000.0,
					"total_lines_added": 250.0,
					"total_lines_removed": 45.0
				},
				"context_window": {
					"context_window_size": 1000000.0,
					"total_input_tokens": 50000.0,
					"total_output_tokens": 1200.0,
					"current_usage": {
						"input_tokens": 50000.0,
						"output_tokens": 1200.0,
						"cache_creation_input_tokens": 5000.0,
						"cache_read_input_tokens": 10000.0
					},
					"used_percentage": 5.0,
					"remaining_percentage": 95.0
				},
				"vim": {
					"mode": "INSERT"
				},
				"worktree": {
					"name": "wt1",
					"path": "/workspace/worktree",
					"branch": "feature/branch",
					"original_cwd": "/workspace",
					"original_branch": "main"
				},
				"rate_limits": {
					"five_hour": {
						"used_percentage": 15.0,
						"resets_at": 1788090000.0
					},
					"seven_day": {
						"used_percentage": 1.0,
						"resets_at": 1788690000.0
					},
					"seven_day_sonnet": {
						"used_percentage": 2.5,
						"resets_at": 1788690000.0
					},
					"seven_day_opus": {
						"used_percentage": 0.0,
						"resets_at": 1788690000.0
					}
				},
				"quota": {
					"gemini-5h": {
						"remaining_fraction": 0.85,
						"reset_time": "2026-08-30T16:00:00Z",
						"reset_in_seconds": 3600.0
					},
					"3p-weekly": {
						"remaining_fraction": 0.99,
						"reset_time": "2026-09-06T00:00:00Z",
						"reset_in_seconds": 604800.0
					}
				},
				"sandbox": {
					"enabled": true
				},
				"terminal_width": 140,
				"agent_state": "RUNNING",
				"artifact_count": 5,
				"subagents": ["subagent-1", "subagent-2"],
				"task_count": 3,
				"vcs": {
					"type": "git",
					"branch": "feature/branch",
					"dirty": true
				}
			}`,
			verify: func(t *testing.T, s StatusJSON) {
				if s.HookEventName != "status" || s.SessionID != "antigravity-session-001" {
					t.Errorf("Unexpected basic metadata: %+v", s)
				}
				if s.Model.ID != "gemini-1.5-pro" || s.Model.DisplayName != "Gemini 1.5 Pro" {
					t.Errorf("Unexpected Model: %+v", s.Model)
				}
				if s.Workspace == nil || s.Workspace.CurrentDir != "/workspace/app" {
					t.Errorf("Unexpected Workspace: %+v", s.Workspace)
				}
				if s.Cost == nil || *s.Cost.TotalCostUSD != 0.1234 {
					t.Errorf("Unexpected Cost: %+v", s.Cost)
				}
				if s.ContextWindow == nil || *s.ContextWindow.UsedPercentage != 5.0 {
					t.Errorf("Unexpected ContextWindow: %+v", s.ContextWindow)
				}
				if s.ContextWindow.CurrentUsage == nil || s.ContextWindow.CurrentUsage.CacheCreationInputTokens != 5000.0 {
					t.Errorf("Unexpected CurrentUsage: %+v", s.ContextWindow.CurrentUsage)
				}
				if s.RateLimits == nil || s.RateLimits.FiveHour == nil || *s.RateLimits.FiveHour.UsedPercentage != 15.0 {
					t.Errorf("Unexpected RateLimits: %+v", s.RateLimits)
				}
				if len(s.Quota) != 2 || *s.Quota["gemini-5h"].RemainingFraction != 0.85 {
					t.Errorf("Unexpected Quota: %+v", s.Quota)
				}
				if s.Sandbox == nil || !*s.Sandbox.Enabled {
					t.Errorf("Unexpected Sandbox: %+v", s.Sandbox)
				}
				if s.AgentState != "RUNNING" || *s.ArtifactCount != 5 || *s.TaskCount != 3 {
					t.Errorf("Unexpected Agent/Artifact/Task: state=%s, artifacts=%v, tasks=%v", s.AgentState, s.ArtifactCount, s.TaskCount)
				}
				if s.VCS == nil || s.VCS.Branch != "feature/branch" || !*s.VCS.Dirty {
					t.Errorf("Unexpected VCS: %+v", s.VCS)
				}
			},
		},
		{
			name: "claude code minimal with numeric context and string model",
			jsonText: `{
				"model": "claude-3-7-sonnet",
				"cwd": "/home/user/project",
				"context_window": {
					"current_usage": 42000.0
				},
				"subagents": 4,
				"terminal_width": 80
			}`,
			verify: func(t *testing.T, s StatusJSON) {
				if s.Model.ID != "claude-3-7-sonnet" || s.Model.DisplayName != "claude-3-7-sonnet" {
					t.Errorf("Unexpected Model string unmarshaling: %+v", s.Model)
				}
				if s.ContextWindow == nil || s.ContextWindow.CurrentUsage == nil || s.ContextWindow.CurrentUsage.InputTokens != 42000.0 {
					t.Errorf("Unexpected numeric current_usage unmarshaling: %+v", s.ContextWindow)
				}
				if s.CWD != "/home/user/project" {
					t.Errorf("Unexpected CWD: %s", s.CWD)
				}
			},
		},
		{
			name: "payload with null fields and empty values",
			jsonText: `{
				"hook_event_name": "",
				"session_id": "",
				"transcript_path": "",
				"cwd": "",
				"model": "",
				"workspace": null,
				"version": "",
				"output_style": null,
				"effort": null,
				"cost": null,
				"context_window": null,
				"vim": null,
				"worktree": null,
				"rate_limits": null,
				"quota": null,
				"sandbox": null,
				"terminal_width": null,
				"agent_state": "",
				"artifact_count": null,
				"subagents": null,
				"task_count": null,
				"vcs": null
			}`,
			verify: func(t *testing.T, s StatusJSON) {
				if s.Model.ID != "" || s.Model.DisplayName != "" {
					t.Errorf("Expected empty Model, got %+v", s.Model)
				}
				if s.Workspace != nil || s.Cost != nil || s.ContextWindow != nil || s.VCS != nil {
					t.Errorf("Expected nil pointers for null fields, got non-nil")
				}
			},
		},
		{
			name:     "payload with completely omitted fields",
			jsonText: `{}`,
			verify: func(t *testing.T, s StatusJSON) {
				if s.Model.ID != "" || s.Workspace != nil || s.ContextWindow != nil {
					t.Errorf("Expected zero values for omitted fields, got %+v", s)
				}
			},
		},
		{
			name: "payload with extra unknown fields",
			jsonText: `{
				"model": "gemini-2.5-flash",
				"unknown_future_field_object": {"nested": "value", "arr": [1, 2, 3]},
				"unknown_number": 999999,
				"unknown_bool": true
			}`,
			verify: func(t *testing.T, s StatusJSON) {
				if s.Model.ID != "gemini-2.5-flash" {
					t.Errorf("Expected Model ID gemini-2.5-flash, got %s", s.Model.ID)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var s StatusJSON
			err := json.Unmarshal([]byte(tc.jsonText), &s)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON payload: %v", err)
			}
			tc.verify(t, s)
		})
	}
}

func TestWidgetItem_EdgeCases_Adversarial(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected WidgetItem
	}{
		{
			name:     "custom text with multiple colons and URL",
			input:    "custom-text:http://example.com:8080/path",
			expected: WidgetItem{Type: "custom-text", Text: "http://example.com:8080/path"},
		},
		{
			name:     "quota with multi-part colon key",
			input:    "quota:cloud:region:tenant:5h",
			expected: WidgetItem{Type: "quota", Key: "cloud:region:tenant:5h"},
		},
		{
			name:     "git-branch with colon inside symbol",
			input:    "git-branch:git: branch: ",
			expected: WidgetItem{Type: "git-branch", Symbol: "git: branch: "},
		},
		{
			name:     "custom text with leading and trailing spaces in text",
			input:    "custom-text:  padded text  ",
			expected: WidgetItem{Type: "custom-text", Text: "  padded text  "},
		},
		{
			name:     "git-branch with space symbol",
			input:    "git-branch:   ",
			expected: WidgetItem{Type: "git-branch", Symbol: "   "},
		},
		{
			name:     "japanese unicode text in custom-text",
			input:    "custom-text:本番環境 (東京)",
			expected: WidgetItem{Type: "custom-text", Text: "本番環境 (東京)"},
		},
		{
			name:     "nerd font glyph in branch symbol",
			input:    "git-branch: ",
			expected: WidgetItem{Type: "git-branch", Symbol: " "},
		},
		{
			name:     "empty shorthand parameter custom-text:",
			input:    "custom-text:",
			expected: WidgetItem{Type: "custom-text", Text: ""},
		},
		{
			name:     "empty shorthand parameter quota:",
			input:    "quota:",
			expected: WidgetItem{Type: "quota", Key: ""},
		},
		{
			name:     "empty string",
			input:    "",
			expected: WidgetItem{Type: ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var w WidgetItem
			err := w.UnmarshalText([]byte(tc.input))
			if err != nil {
				t.Fatalf("Unexpected error unmarshaling text %q: %v", tc.input, err)
			}
			if !reflect.DeepEqual(w, tc.expected) {
				t.Errorf("For input %q: expected %+v, got %+v", tc.input, tc.expected, w)
			}
		})
	}
}

func TestWidgetItem_QuotaShorthand_AllVariants(t *testing.T) {
	tests := []struct {
		name       string
		tomlInput  string
		expected   WidgetItem
		roundtrips bool
	}{
		// Individual Quota widgets
		{"quota-5h plain", `"quota-5h"`, WidgetItem{Type: "quota-5h"}, true},
		{"quota-7d plain", `"quota-7d"`, WidgetItem{Type: "quota-7d"}, true},
		{"quota-3p-5h plain", `"quota-3p-5h"`, WidgetItem{Type: "quota-3p-5h"}, true},
		{"quota-3p-7d plain", `"quota-3p-7d"`, WidgetItem{Type: "quota-3p-7d"}, true},

		// Individual Quota-Bar widgets
		{"quota-bar-5h plain", `"quota-bar-5h"`, WidgetItem{Type: "quota-bar-5h"}, true},
		{"quota-bar-7d plain", `"quota-bar-7d"`, WidgetItem{Type: "quota-bar-7d"}, true},
		{"quota-bar-3p-5h plain", `"quota-bar-3p-5h"`, WidgetItem{Type: "quota-bar-3p-5h"}, true},
		{"quota-bar-3p-7d plain", `"quota-bar-3p-7d"`, WidgetItem{Type: "quota-bar-3p-7d"}, true},

		// Quota colon shorthands
		{"quota:gemini-5h shorthand", `"quota:gemini-5h"`, WidgetItem{Type: "quota", Key: "gemini-5h"}, true},
		{"quota:gemini-weekly shorthand", `"quota:gemini-weekly"`, WidgetItem{Type: "quota", Key: "gemini-weekly"}, true},
		{"quota:3p-5h shorthand", `"quota:3p-5h"`, WidgetItem{Type: "quota", Key: "3p-5h"}, true},
		{"quota:3p-weekly shorthand", `"quota:3p-weekly"`, WidgetItem{Type: "quota", Key: "3p-weekly"}, true},

		// Quota-Bar colon shorthands
		{"quota-bar:gemini-5h shorthand", `"quota-bar:gemini-5h"`, WidgetItem{Type: "quota-bar", Key: "gemini-5h"}, true},
		{"quota-bar:gemini-weekly shorthand", `"quota-bar:gemini-weekly"`, WidgetItem{Type: "quota-bar", Key: "gemini-weekly"}, true},
		{"quota-bar:3p-5h shorthand", `"quota-bar:3p-5h"`, WidgetItem{Type: "quota-bar", Key: "3p-5h"}, true},
		{"quota-bar:3p-weekly shorthand", `"quota-bar:3p-weekly"`, WidgetItem{Type: "quota-bar", Key: "3p-weekly"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wrapper struct {
				Widget WidgetItem `toml:"widget"`
			}
			doc := "widget = " + tt.tomlInput
			err := toml.Unmarshal([]byte(doc), &wrapper)
			if err != nil {
				t.Fatalf("Failed to unmarshal %s: %v", doc, err)
			}
			if !reflect.DeepEqual(wrapper.Widget, tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, wrapper.Widget)
			}

			// Also test UnmarshalText directly without quotes
			cleanStr := strings.Trim(tt.tomlInput, `"`)
			var textWidget WidgetItem
			err = textWidget.UnmarshalText([]byte(cleanStr))
			if err != nil {
				t.Fatalf("Failed to UnmarshalText %s: %v", cleanStr, err)
			}
			if !reflect.DeepEqual(textWidget, tt.expected) {
				t.Errorf("UnmarshalText expected %+v, got %+v", tt.expected, textWidget)
			}

			if tt.roundtrips {
				marshaled, err := toml.Marshal(wrapper)
				if err != nil {
					t.Fatalf("Failed to marshal %+v: %v", wrapper, err)
				}
				var restored struct {
					Widget WidgetItem `toml:"widget"`
				}
				err = toml.Unmarshal(marshaled, &restored)
				if err != nil {
					t.Fatalf("Failed to unmarshal remarshaled %s: %v", string(marshaled), err)
				}
				if !reflect.DeepEqual(wrapper.Widget, restored.Widget) {
					t.Errorf("Roundtrip mismatch: original %+v vs restored %+v (TOML: %s)", wrapper.Widget, restored.Widget, string(marshaled))
				}
			}
		})
	}
}

func TestSettings_QuotaAllVariants_FullDocumentRoundTrip(t *testing.T) {
	tomlDoc := `
lines = [
  ["quota-5h", "quota-7d", "quota-3p-5h", "quota-3p-7d"],
  ["quota-bar-5h", "quota-bar-7d", "quota-bar-3p-5h", "quota-bar-3p-7d"],
  ["quota:gemini-5h", "quota:3p-weekly", "quota-bar:gemini-weekly", "quota-bar:3p-5h"]
]

[powerline]
enabled = false
theme = "nord-aurora"

[general]
color_level = 1
`

	var s Settings
	err := toml.Unmarshal([]byte(tomlDoc), &s)
	if err != nil {
		t.Fatalf("Failed to unmarshal settings with all quota variants: %v", err)
	}

	if len(s.Lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(s.Lines))
	}
	if len(s.Lines[0]) != 4 || len(s.Lines[1]) != 4 || len(s.Lines[2]) != 4 {
		t.Fatalf("Expected 4 items per line, got %d, %d, %d", len(s.Lines[0]), len(s.Lines[1]), len(s.Lines[2]))
	}

	marshaled, err := toml.Marshal(s)
	if err != nil {
		t.Fatalf("Failed to marshal settings: %v", err)
	}

	var restored Settings
	err = toml.Unmarshal(marshaled, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal remarshaled settings: %v\nTOML:\n%s", err, string(marshaled))
	}

	if !reflect.DeepEqual(s, restored) {
		t.Errorf("Roundtrip mismatch!\nExpected: %+v\nGot:      %+v\nTOML:\n%s", s, restored, string(marshaled))
	}
}
