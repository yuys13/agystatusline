package e2e_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

// TestE2E_Tier2_EmptyTelemetry verifies system stability when receiving empty JSON object.
func TestE2E_Tier2_EmptyTelemetry(t *testing.T) {
	tempDir := t.TempDir()
	cfg := types.DefaultSettings()
	cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

	stdout, stderr, exitCode := runCLI(t, `{}`, "--config", cfgPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 on empty telemetry, got %d (stderr: %q)", exitCode, stderr)
	}

	plain := stripANSI(stdout)
	// Default settings include agent-state which defaults to ● READY
	if !strings.Contains(plain, "● READY") {
		t.Errorf("Expected fallback agent state '● READY' on empty telemetry, got %q", plain)
	}
}

// TestE2E_Tier2_NilFieldsTelemetry verifies robust handling when all optional pointer fields are explicitly null/nil.
func TestE2E_Tier2_NilFieldsTelemetry(t *testing.T) {
	nullJSON := `{
		"hook_event_name": "status",
		"session_id": "null-test",
		"transcript_path": "",
		"cwd": "/tmp",
		"model": {"id": "", "display_name": ""},
		"workspace": null,
		"version": "1.0.0",
		"output_style": null,
		"effort": null,
		"cost": null,
		"context_window": null,
		"vim": null,
		"worktree": null,
		"rate_limits": null,
		"quota": null,
		"sandbox": null,
		"terminal_width": 100,
		"agent_state": "",
		"artifact_count": null,
		"subagents": null,
		"task_count": null,
		"vcs": null
	}`

	tempDir := t.TempDir()
	cfg := types.DefaultSettings()
	cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

	stdout, stderr, exitCode := runCLI(t, nullJSON, "--config", cfgPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 on null fields telemetry, got %d (stderr: %q)", exitCode, stderr)
	}

	if len(stdout) == 0 {
		t.Errorf("Expected non-empty stdout output, got empty")
	}
}

// TestE2E_Tier2_TokenAndPercentageBoundaries tests boundary percentage and token values (0%, 100%, 0 tokens, large tokens).
func TestE2E_Tier2_TokenAndPercentageBoundaries(t *testing.T) {
	tests := []struct {
		name       string
		pct        float64
		wantTokens string
		wantBar    string
	}{
		{
			name:       "Zero percent usage",
			pct:        0.0,
			wantTokens: "0.0%",
			wantBar:    "···············",
		},
		{
			name:       "Boundary 59.9% usage",
			pct:        59.9,
			wantTokens: "59.9%",
			wantBar:    "████████▓······",
		},
		{
			name:       "Boundary 60.0% usage",
			pct:        60.0,
			wantTokens: "60.0%",
			wantBar:    "█████████······",
		},
		{
			name:       "Boundary 89.9% usage",
			pct:        89.9,
			wantTokens: "89.9%",
			wantBar:    "█████████████░·",
		},
		{
			name:       "Boundary 90.0% usage",
			pct:        90.0,
			wantTokens: "90.0%",
			wantBar:    "█████████████▒·",
		},
		{
			name:       "Complete 100.0% usage",
			pct:        100.0,
			wantTokens: "100.0%",
			wantBar:    "███████████████",
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

			statusJSON := fmt.Sprintf(`{
				"context_window": {
					"used_percentage": %.1f,
					"total_input_tokens": 1000000.0
				}
			}`, tt.pct)

			stdout, stderr, exitCode := runCLI(t, statusJSON, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
			}

			plain := stripANSI(stdout)
			if !strings.Contains(plain, tt.wantTokens) {
				t.Errorf("Expected token string %q, got %q", tt.wantTokens, plain)
			}
			if !strings.Contains(plain, tt.wantBar) {
				t.Errorf("Expected progress bar %q, got %q", tt.wantBar, plain)
			}
		})
	}
}

// TestE2E_Tier2_UnicodeAndCJKBranchNames tests multibyte UTF-8, Japanese CJK, and emoji branch names.
func TestE2E_Tier2_UnicodeAndCJKBranchNames(t *testing.T) {
	tests := []struct {
		name       string
		branch     string
		wantBranch string
	}{
		{
			name:       "Japanese Kanji and Kana branch",
			branch:     "feature/日本語ブランチ名",
			wantBranch: "⎇ feature/日本語ブランチ名",
		},
		{
			name:       "Emoji in branch name",
			branch:     "fix/⚡️-speedup-🚀",
			wantBranch: "⎇ fix/⚡️-speedup-🚀",
		},
		{
			name:       "Greek and Cyrillic symbols",
			branch:     "release/v1.0-αβγ-тест",
			wantBranch: "⎇ release/v1.0-αβγ-тест",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfg.Lines = [][]types.WidgetItem{
				{
					{Type: "git-branch"},
				},
			}
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			statusJSON := fmt.Sprintf(`{
				"vcs": {"type": "git", "branch": %q, "dirty": false}
			}`, tt.branch)

			stdout, stderr, exitCode := runCLI(t, statusJSON, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
			}

			plain := stripANSI(stdout)
			if !strings.Contains(plain, tt.wantBranch) {
				t.Errorf("Expected output to contain %q, got %q", tt.wantBranch, plain)
			}
		})
	}
}

// TestE2E_Tier2_QuotaResetTimeBoundaries verifies time unit boundary conversions (s, m, h, d).
func TestE2E_Tier2_QuotaResetTimeBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		seconds   float64
		wantReset string
	}{
		{name: "0 seconds", seconds: 0.0, wantReset: "0s"},
		{name: "59 seconds", seconds: 59.0, wantReset: "59s"},
		{name: "60 seconds (1 minute boundary)", seconds: 60.0, wantReset: "1m"},
		{name: "61 seconds (1m 1s)", seconds: 61.0, wantReset: "1m 1s"},
		{name: "3599 seconds (59m 59s)", seconds: 3599.0, wantReset: "59m 59s"},
		{name: "3600 seconds (1 hour boundary)", seconds: 3600.0, wantReset: "1h"},
		{name: "7320 seconds (2h 2m)", seconds: 7320.0, wantReset: "2h 2m"},
		{name: "86399 seconds (23h 59m)", seconds: 86399.0, wantReset: "23h 59m"},
		{name: "86400 seconds (1 day boundary)", seconds: 86400.0, wantReset: "1d"},
		{name: "90000 seconds (1d 1h)", seconds: 90000.0, wantReset: "1d 1h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfg.Lines = [][]types.WidgetItem{
				{
					{
						Type: "quota",
						Key:  "gemini-5h",
					},
				},
			}
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			statusJSON := fmt.Sprintf(`{
				"quota": {
					"gemini-5h": {
						"remaining_fraction": 0.5,
						"reset_in_seconds": %.1f
					}
				}
			}`, tt.seconds)

			stdout, stderr, exitCode := runCLI(t, statusJSON, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
			}

			plain := stripANSI(stdout)
			if !strings.Contains(plain, tt.wantReset) {
				t.Errorf("Expected reset countdown %q, got %q", tt.wantReset, plain)
			}
		})
	}
}

// TestE2E_Tier2_TerminalWidthConstraints tests ANSI-aware string truncation across diverse widths.
func TestE2E_Tier2_TerminalWidthConstraints(t *testing.T) {
	widths := []int{10, 20, 30, 60, 80, 200}

	for _, w := range widths {
		t.Run(fmt.Sprintf("Width_%d", w), func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			statusJSON := fmt.Sprintf(`{
				"model": {"display_name": "Very Long Model Name That Exceeds Narrow Terminal Width Constraints"},
				"terminal_width": %d,
				"agent_state": "working"
			}`, w)

			stdout, stderr, exitCode := runCLI(t, statusJSON, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0 for width %d, got %d (stderr: %q)", w, exitCode, stderr)
			}

			lines := strings.Split(strings.TrimSpace(stdout), "\n")
			for i, line := range lines {
				visWidth := getVisibleWidth(line)
				if visWidth > w {
					t.Errorf("Line %d visible width %d exceeds specified terminal width %d (content: %q)", i, visWidth, w, stripANSI(line))
				}
			}
		})
	}
}
