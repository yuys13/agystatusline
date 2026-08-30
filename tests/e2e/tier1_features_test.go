package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

// TestE2E_Tier1_ModelWidget tests the 'model' widget with various inputs and fallbacks.
func TestE2E_Tier1_ModelWidget(t *testing.T) {
	tests := []struct {
		name       string
		statusJSON string
		wantPlain  string
	}{
		{
			name: "Model with DisplayName",
			statusJSON: `{
				"model": {"id": "gemini-3.5-flash-medium", "display_name": "Gemini 3.5 Flash (Medium)"}
			}`,
			wantPlain: "Gemini 3.5 Flash (Medium)",
		},
		{
			name: "Model with ID only (DisplayName empty)",
			statusJSON: `{
				"model": {"id": "claude-3-7-sonnet", "display_name": ""}
			}`,
			wantPlain: "claude-3-7-sonnet",
		},
		{
			name: "Model with plain string unmarshaling",
			statusJSON: `{
				"model": "gpt-4o-mini"
			}`,
			wantPlain: "gpt-4o-mini",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfg.Lines = [][]types.WidgetItem{
				{
					{Type: "model"},
				},
			}
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			stdout, stderr, exitCode := runCLI(t, tt.statusJSON, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
			}

			plain := stripANSI(stdout)
			if !strings.Contains(plain, tt.wantPlain) {
				t.Errorf("Expected output to contain %q, got %q (raw: %q)", tt.wantPlain, plain, stdout)
			}
		})
	}
}

// TestE2E_Tier1_ContextBarWidget tests the 'context-bar' widget rendering progress bar and percentages.
func TestE2E_Tier1_ContextBarWidget(t *testing.T) {
	tests := []struct {
		name       string
		statusJSON string
		wantTokens string
		wantRune   string
	}{
		{
			name: "Normal usage 20%",
			statusJSON: `{
				"context_window": {
					"used_percentage": 20.0,
					"total_input_tokens": 14200.0
				}
			}`,
			wantTokens: "20.0%",
			wantRune:   "█",
		},
		{
			name: "High usage 92.5%",
			statusJSON: `{
				"context_window": {
					"used_percentage": 92.5,
					"total_input_tokens": 185000.0
				}
			}`,
			wantTokens: "92.5%",
			wantRune:   "█",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfg.Lines = [][]types.WidgetItem{
				{
					{Type: "context-bar"},
				},
			}
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			stdout, stderr, exitCode := runCLI(t, tt.statusJSON, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
			}

			plain := stripANSI(stdout)
			if !strings.Contains(plain, "ctx") {
				t.Errorf("Expected output to contain title 'ctx', got %q", plain)
			}
			if !strings.Contains(plain, tt.wantTokens) {
				t.Errorf("Expected output to contain token pct %q, got %q", tt.wantTokens, plain)
			}
			if !strings.Contains(plain, tt.wantRune) {
				t.Errorf("Expected output to contain progress rune %q, got %q", tt.wantRune, plain)
			}
		})
	}
}

// TestE2E_Tier1_AgentStateWidget tests all states for the 'agent-state' widget.
func TestE2E_Tier1_AgentStateWidget(t *testing.T) {
	tests := []struct {
		state     string
		wantPlain string
	}{
		{state: "idle", wantPlain: "● READY"},
		{state: "thinking", wantPlain: "◆ THINKING"},
		{state: "working", wantPlain: "⚙ WORKING"},
		{state: "tool_use", wantPlain: "🔧 TOOL"},
		{state: "waiting", wantPlain: "⏳ WAITING"},
		{state: "", wantPlain: "● READY"}, // default fallback
	}

	for _, tt := range tests {
		t.Run("State_"+tt.state, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfg.Lines = [][]types.WidgetItem{
				{
					{Type: "agent-state"},
				},
			}
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			statusJSON := fmt.Sprintf(`{"agent_state": "%s"}`, tt.state)
			stdout, stderr, exitCode := runCLI(t, statusJSON, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
			}

			plain := stripANSI(stdout)
			if !strings.Contains(plain, tt.wantPlain) {
				t.Errorf("Expected state symbol %q, got %q", tt.wantPlain, plain)
			}
		})
	}
}

// TestE2E_Tier1_ArtifactsWidget tests artifact count formatting.
func TestE2E_Tier1_ArtifactsWidget(t *testing.T) {
	tests := []struct {
		name       string
		statusJSON string
		wantPlain  string
	}{
		{
			name:       "Positive artifact count",
			statusJSON: `{"artifact_count": 5}`,
			wantPlain:  "artifacts 5",
		},
		{
			name:       "Zero artifact count",
			statusJSON: `{"artifact_count": 0}`,
			wantPlain:  "artifacts 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfg.Lines = [][]types.WidgetItem{
				{
					{Type: "artifacts"},
				},
			}
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			stdout, stderr, exitCode := runCLI(t, tt.statusJSON, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
			}

			plain := stripANSI(stdout)
			if !strings.Contains(plain, tt.wantPlain) {
				t.Errorf("Expected %q in output, got %q", tt.wantPlain, plain)
			}
		})
	}
}

// TestE2E_Tier1_SubagentsWidget tests subagent count representations (array & number).
func TestE2E_Tier1_SubagentsWidget(t *testing.T) {
	tests := []struct {
		name       string
		statusJSON string
		wantPlain  string
	}{
		{
			name:       "Subagents array",
			statusJSON: `{"subagents": ["subagent-1", "subagent-2", "subagent-3"]}`,
			wantPlain:  "subagents 3",
		},
		{
			name:       "Subagents number",
			statusJSON: `{"subagents": 4.0}`,
			wantPlain:  "subagents 4",
		},
		{
			name:       "Subagents empty array",
			statusJSON: `{"subagents": []}`,
			wantPlain:  "subagents 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfg.Lines = [][]types.WidgetItem{
				{
					{Type: "subagents"},
				},
			}
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			stdout, stderr, exitCode := runCLI(t, tt.statusJSON, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
			}

			plain := stripANSI(stdout)
			if !strings.Contains(plain, tt.wantPlain) {
				t.Errorf("Expected %q in output, got %q", tt.wantPlain, plain)
			}
		})
	}
}

// TestE2E_Tier1_TasksWidget tests task count formatting.
func TestE2E_Tier1_TasksWidget(t *testing.T) {
	tests := []struct {
		name       string
		statusJSON string
		wantPlain  string
	}{
		{
			name:       "Positive task count",
			statusJSON: `{"task_count": 2}`,
			wantPlain:  "tasks 2",
		},
		{
			name:       "Zero task count",
			statusJSON: `{"task_count": 0}`,
			wantPlain:  "tasks 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfg.Lines = [][]types.WidgetItem{
				{
					{Type: "tasks"},
				},
			}
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			stdout, stderr, exitCode := runCLI(t, tt.statusJSON, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
			}

			plain := stripANSI(stdout)
			if !strings.Contains(plain, tt.wantPlain) {
				t.Errorf("Expected %q in output, got %q", tt.wantPlain, plain)
			}
		})
	}
}

// TestE2E_Tier1_SandboxWidget tests sandbox state display.
func TestE2E_Tier1_SandboxWidget(t *testing.T) {
	tests := []struct {
		name       string
		statusJSON string
		wantPlain  string
	}{
		{
			name:       "Sandbox enabled true",
			statusJSON: `{"sandbox": {"enabled": true}}`,
			wantPlain:  "sandbox on",
		},
		{
			name:       "Sandbox enabled false",
			statusJSON: `{"sandbox": {"enabled": false}}`,
			wantPlain:  "sandbox off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfg.Lines = [][]types.WidgetItem{
				{
					{Type: "sandbox"},
				},
			}
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			stdout, stderr, exitCode := runCLI(t, tt.statusJSON, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
			}

			plain := stripANSI(stdout)
			if !strings.Contains(plain, tt.wantPlain) {
				t.Errorf("Expected %q in output, got %q", tt.wantPlain, plain)
			}
		})
	}
}

// TestE2E_Tier1_QuotaWidget tests quota display and calculations.
func TestE2E_Tier1_QuotaWidget(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		customText string
		wantPlain  string
	}{
		{
			name:      "Default quota display with key",
			key:       "gemini-5h",
			wantPlain: "gemini-5h 50.19% (2h 28m)",
		},
		{
			name:       "Quota with custom text label",
			key:        "gemini-5h",
			customText: "RateLimit-5h",
			wantPlain:  "RateLimit-5h 50.19% (2h 28m)",
		},
	}

	statusJSON := `{
		"quota": {
			"gemini-5h": {
				"remaining_fraction": 0.5019,
				"reset_in_seconds": 8891.0
			}
		}
	}`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfg.Lines = [][]types.WidgetItem{
				{
					{
						Type: "quota",
						Key:  tt.key,
						Text: tt.customText,
					},
				},
			}
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			stdout, stderr, exitCode := runCLI(t, statusJSON, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
			}

			plain := stripANSI(stdout)
			if !strings.Contains(plain, tt.wantPlain) {
				t.Errorf("Expected %q in output, got %q", tt.wantPlain, plain)
			}
		})
	}
}

// TestE2E_Tier1_QuotaBarWidget tests the quota-bar progress bar visualization.
func TestE2E_Tier1_QuotaBarWidget(t *testing.T) {
	statusJSON := `{
		"quota": {
			"gemini-5h": {
				"remaining_fraction": 0.75,
				"reset_in_seconds": 3600.0
			}
		}
	}`

	tempDir := t.TempDir()
	cfg := types.DefaultSettings()
	cfg.Lines = [][]types.WidgetItem{
		{
			{
				Type: "quota-bar",
				Key:  "gemini-5h",
			},
		},
	}
	cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

	stdout, stderr, exitCode := runCLI(t, statusJSON, "--config", cfgPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
	}

	plain := stripANSI(stdout)
	if !strings.Contains(plain, "5h") {
		t.Errorf("Expected default label '5h' in output, got %q", plain)
	}
	if !strings.Contains(plain, "75.0%") {
		t.Errorf("Expected percentage '75.0%%' in output, got %q", plain)
	}
	if !strings.Contains(plain, "(1h)") {
		t.Errorf("Expected reset countdown '(1h)' in output, got %q", plain)
	}
	if !strings.Contains(plain, "█") {
		t.Errorf("Expected progress bar block character '█' in output, got %q", plain)
	}
}

// TestE2E_Tier1_GitBranchWidget tests git branch display via telemetry and custom symbol.
func TestE2E_Tier1_GitBranchWidget(t *testing.T) {
	tests := []struct {
		name         string
		customSymbol string
		statusJSON   string
		wantPlain    string
	}{
		{
			name: "Clean branch telemetry",
			statusJSON: `{
				"vcs": {"type": "git", "branch": "main", "dirty": false}
			}`,
			wantPlain: "⎇ main",
		},
		{
			name: "Dirty branch telemetry",
			statusJSON: `{
				"vcs": {"type": "git", "branch": "feature/refactor", "dirty": true}
			}`,
			wantPlain: "⎇ feature/refactor*",
		},
		{
			name:         "Custom branch symbol",
			customSymbol: "git:",
			statusJSON: `{
				"vcs": {"type": "git", "branch": "master", "dirty": false}
			}`,
			wantPlain: "git:master",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfg.Lines = [][]types.WidgetItem{
				{
					{
						Type:   "git-branch",
						Symbol: tt.customSymbol,
					},
				},
			}
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			stdout, stderr, exitCode := runCLI(t, tt.statusJSON, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
			}

			plain := stripANSI(stdout)
			if !strings.Contains(plain, tt.wantPlain) {
				t.Errorf("Expected %q in output, got %q", tt.wantPlain, plain)
			}
		})
	}
}

// TestE2E_Tier1_GitChangesWidget tests git-changes rendering in real repo and non-git environment.
func TestE2E_Tier1_GitChangesWidget(t *testing.T) {
	t.Run("Inside real git repository with modified files", func(t *testing.T) {
		repoDir := t.TempDir()
		initCmd := exec.Command("git", "init", "-b", "main")
		initCmd.Dir = repoDir
		if err := initCmd.Run(); err != nil {
			t.Fatalf("Failed to init test git repo: %v", err)
		}

		// Configure git user
		_ = exec.Command("git", "-C", repoDir, "config", "user.name", "Test User").Run()
		_ = exec.Command("git", "-C", repoDir, "config", "user.email", "test@example.com").Run()

		testFile := filepath.Join(repoDir, "hello.txt")
		_ = os.WriteFile(testFile, []byte("line1\nline2\n"), 0644)
		_ = exec.Command("git", "-C", repoDir, "add", "hello.txt").Run()
		_ = exec.Command("git", "-C", repoDir, "commit", "-m", "initial commit").Run()

		// Modify file (1 insertion, 0 deletions)
		_ = os.WriteFile(testFile, []byte("line1\nline2\nline3\n"), 0644)

		cfg := types.DefaultSettings()
		cfg.Lines = [][]types.WidgetItem{
			{
				{Type: "git-changes"},
			},
		}
		cfgPath := writeTOMLConfig(t, repoDir, "settings.toml", cfg)

		statusJSON := fmt.Sprintf(`{"cwd": %q}`, repoDir)
		stdout, stderr, exitCode := runCLIWithDir(t, statusJSON, repoDir, "--config", cfgPath)
		if exitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
		}

		plain := stripANSI(stdout)
		if !strings.Contains(plain, "(+1,-0)") {
			t.Errorf("Expected git changes '(+1,-0)' in output, got %q", plain)
		}
	})

	t.Run("Outside git repository", func(t *testing.T) {
		nonGitDir := t.TempDir()
		cfg := types.DefaultSettings()
		cfg.Lines = [][]types.WidgetItem{
			{
				{Type: "git-changes"},
			},
		}
		cfgPath := writeTOMLConfig(t, nonGitDir, "settings.toml", cfg)

		statusJSON := fmt.Sprintf(`{"cwd": %q}`, nonGitDir)
		stdout, stderr, exitCode := runCLIWithDir(t, statusJSON, nonGitDir, "--config", cfgPath)
		if exitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
		}

		plain := stripANSI(stdout)
		if !strings.Contains(plain, "(no git)") {
			t.Errorf("Expected '(no git)' in output, got %q", plain)
		}
	})
}

// TestE2E_Tier1_CustomTextWidget tests user-defined static custom-text widget.
func TestE2E_Tier1_CustomTextWidget(t *testing.T) {
	tempDir := t.TempDir()
	cfg := types.DefaultSettings()
	cfg.Lines = [][]types.WidgetItem{
		{
			{
				Type: "custom-text",
				Text: "AGY E2E Environment Active",
			},
		},
	}
	cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

	stdout, stderr, exitCode := runCLI(t, `{}`, "--config", cfgPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
	}

	plain := stripANSI(stdout)
	if !strings.Contains(plain, "AGY E2E Environment Active") {
		t.Errorf("Expected custom text in output, got %q", plain)
	}
}
