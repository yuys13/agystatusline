package e2e_test

import (
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

// TestE2E_Tier3_MultiLineRendering verifies multi-line configuration renders exact separate lines.
func TestE2E_Tier3_MultiLineRendering(t *testing.T) {
	tempDir := t.TempDir()
	cfg := types.DefaultSettings()
	cfg.Lines = [][]types.WidgetItem{
		{
			{Type: "model"},
			{Type: "agent-state"},
		},
		{
			{Type: "context-bar"},
			{Type: "artifacts"},
		},
		{
			{Type: "git-branch"},
			{Type: "sandbox"},
		},
	}
	cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

	stdout, stderr, exitCode := runCLI(t, standardStatusJSON(), "--config", cfgPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 for multi-line rendering, got %d (stderr: %q)", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected exactly 3 rendered lines, got %d (raw: %q)", len(lines), stdout)
	}

	// Line 1 should have model and agent-state
	p0 := stripANSI(lines[0])
	if !strings.Contains(p0, "Gemini 3.5 Flash") || !strings.Contains(p0, "● READY") {
		t.Errorf("Line 1 missing expected widgets: %q", p0)
	}

	// Line 2 should have context and artifacts
	p1 := stripANSI(lines[1])
	if !strings.Contains(p1, "ctx") || !strings.Contains(p1, "artifacts 3") {
		t.Errorf("Line 2 missing expected widgets: %q", p1)
	}

	// Line 3 should have git-branch and sandbox
	p2 := stripANSI(lines[2])
	if !strings.Contains(p2, "⎇ feature/e2e-testing") || !strings.Contains(p2, "sandbox on") {
		t.Errorf("Line 3 missing expected widgets: %q", p2)
	}
}

// TestE2E_Tier3_PowerlineThemes verifies all 10 built-in Powerline themes with Powerline enabled.
func TestE2E_Tier3_PowerlineThemes(t *testing.T) {
	themes := []string{
		"nord",
		"nord-aurora",
		"monokai",
		"solarized",
		"minimal",
		"dracula",
		"catppuccin",
		"gruvbox",
		"onedark",
		"tokyonight",
	}

	for _, themeName := range themes {
		t.Run("Theme_"+themeName, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfg.Powerline.Enabled = true
			cfg.Powerline.Theme = themeName
			cfg.Lines = [][]types.WidgetItem{
				{
					{Type: "model"},
					{Type: "agent-state"},
					{Type: "context-bar"},
				},
			}
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			stdout, stderr, exitCode := runCLI(t, standardStatusJSON(), "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0 for theme %s, got %d (stderr: %q)", themeName, exitCode, stderr)
			}

			plain := stripANSI(stdout)
			if !strings.Contains(plain, "Gemini 3.5 Flash") {
				t.Errorf("Expected model in theme %s output: %q", themeName, plain)
			}
			if !strings.Contains(plain, "● READY") {
				t.Errorf("Expected agent-state in theme %s output: %q", themeName, plain)
			}
		})
	}
}

// TestE2E_Tier3_ColorLevels verifies ANSI 16, ANSI 256, and Truecolor formatting codes.
func TestE2E_Tier3_ColorLevels(t *testing.T) {
	tests := []struct {
		name       string
		colorLevel int
		powerline  bool
		wantEscape string
	}{
		{
			name:       "Non-powerline ColorLevel 2 (ANSI 256)",
			colorLevel: 2,
			powerline:  false,
			wantEscape: "\x1b[38;5;",
		},
		{
			name:       "Powerline ColorLevel 2 (ANSI 256)",
			colorLevel: 2,
			powerline:  true,
			wantEscape: "\x1b[48;5;",
		},
		{
			name:       "Non-powerline ColorLevel 3 (Truecolor)",
			colorLevel: 3,
			powerline:  false,
			wantEscape: "\x1b[38;2;",
		},
		{
			name:       "Powerline ColorLevel 3 (Truecolor)",
			colorLevel: 3,
			powerline:  true,
			wantEscape: "\x1b[48;2;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfg.General.ColorLevel = tt.colorLevel
			cfg.Powerline.Enabled = tt.powerline
			cfg.Powerline.Theme = "nord"
			cfg.Lines = [][]types.WidgetItem{
				{
					{Type: "model", Color: "brightMagenta"},
					{Type: "agent-state"},
				},
			}
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			stdout, stderr, exitCode := runCLI(t, standardStatusJSON(), "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0 for color level %d, got %d (stderr: %q)", tt.colorLevel, exitCode, stderr)
			}

			if !strings.Contains(stdout, tt.wantEscape) {
				t.Errorf("Expected output to contain color escape %q, got raw: %q", tt.wantEscape, stdout)
			}
		})
	}
}

// TestE2E_Tier3_PowerlineCapsAndSeparators tests custom start/end caps and separators.
func TestE2E_Tier3_PowerlineCapsAndSeparators(t *testing.T) {
	tempDir := t.TempDir()
	cfg := types.DefaultSettings()
	cfg.Powerline.Enabled = true
	cfg.Powerline.Theme = "nord"
	cfg.Powerline.Separator = "\uE0B0"
	cfg.Powerline.StartCaps = "\uE0B6"
	cfg.Powerline.EndCaps = "\uE0B4"
	cfg.Lines = [][]types.WidgetItem{
		{
			{Type: "model"},
			{Type: "agent-state"},
		},
	}
	cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

	stdout, stderr, exitCode := runCLI(t, standardStatusJSON(), "--config", cfgPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 for caps test, got %d (stderr: %q)", exitCode, stderr)
	}

	plain := stripANSI(stdout)
	if !strings.Contains(plain, "\uE0B6") {
		t.Errorf("Expected StartCap \uE0B6 in output, got %q", plain)
	}
	if !strings.Contains(plain, "\uE0B4") {
		t.Errorf("Expected EndCap \uE0B4 in output, got %q", plain)
	}
	if !strings.Contains(plain, "\uE0B0") {
		t.Errorf("Expected Separator \uE0B0 in output, got %q", plain)
	}
}

// TestE2E_Tier3_MinimalistMode verifies title prefixes are omitted in minimalist mode.
func TestE2E_Tier3_MinimalistMode(t *testing.T) {
	tempDir := t.TempDir()
	cfg := types.DefaultSettings()
	cfg.General.Minimalist = true
	cfg.Lines = [][]types.WidgetItem{
		{
			{Type: "context-bar"},
			{Type: "artifacts"},
			{Type: "tasks"},
		},
	}
	cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

	stdout, stderr, exitCode := runCLI(t, standardStatusJSON(), "--config", cfgPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 for minimalist mode, got %d (stderr: %q)", exitCode, stderr)
	}

	plain := stripANSI(stdout)
	// In minimalist mode: titles 'ctx', 'artifacts', 'tasks' should not precede raw values
	if strings.Contains(plain, "ctx ") {
		t.Errorf("Expected 'ctx ' title to be omitted in minimalist mode, got %q", plain)
	}
	if strings.Contains(plain, "artifacts ") {
		t.Errorf("Expected 'artifacts ' title to be omitted in minimalist mode, got %q", plain)
	}
	if strings.Contains(plain, "tasks ") {
		t.Errorf("Expected 'tasks ' title to be omitted in minimalist mode, got %q", plain)
	}
	// Raw numbers should still appear
	if !strings.Contains(plain, "3") || !strings.Contains(plain, "1") {
		t.Errorf("Expected raw values '3' and '1' in output, got %q", plain)
	}
}

// TestE2E_Tier3_All8IndividualQuotaWidgetsSimultaneousRendering verifies all 8 quota widgets render simultaneously across multi-line layouts.
func TestE2E_Tier3_All8IndividualQuotaWidgetsSimultaneousRendering(t *testing.T) {
	tempDir := t.TempDir()
	cfg := types.DefaultSettings()
	cfg.Lines = [][]types.WidgetItem{
		{
			{Type: "quota-5h"},
			{Type: "quota-7d"},
			{Type: "quota-3p-5h"},
			{Type: "quota-3p-7d"},
		},
		{
			{Type: "quota-bar-5h"},
			{Type: "quota-bar-7d"},
			{Type: "quota-bar-3p-5h"},
			{Type: "quota-bar-3p-7d"},
		},
	}
	cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

	statusJSON := `{
		"quota": {
			"gemini-5h": {
				"remaining_fraction": 0.5019,
				"reset_in_seconds": 8891.0
			},
			"gemini-weekly": {
				"remaining_fraction": 0.90,
				"reset_in_seconds": 604756.0
			},
			"3p-5h": {
				"remaining_fraction": 0.25,
				"reset_in_seconds": 1800.0
			},
			"3p-weekly": {
				"remaining_fraction": 0.75,
				"reset_in_seconds": 86400.0
			}
		},
		"terminal_width": 200
	}`

	stdout, stderr, exitCode := runCLI(t, statusJSON, "--config", cfgPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 for 8 quota widgets rendering, got %d (stderr: %q)", exitCode, stderr)
	}

	plain := stripANSI(stdout)
	lines := strings.Split(strings.TrimSpace(plain), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 rendered lines, got %d:\n%s", len(lines), plain)
	}

	// Line 1 assertions: quota-5h, quota-7d, quota-3p-5h, quota-3p-7d
	expectedLine1 := []string{
		"5h 50.19% (2h 28m)",
		"7d 90.00% (6d 23h)",
		"3p-5h 25.00% (30m)",
		"3p-7d 75.00% (1d)",
	}
	for _, exp := range expectedLine1 {
		if !strings.Contains(lines[0], exp) {
			t.Errorf("Expected Line 1 to contain %q, got %q", exp, lines[0])
		}
	}

	// Line 2 assertions: quota-bar-5h, quota-bar-7d, quota-bar-3p-5h, quota-bar-3p-7d
	expectedLine2 := []string{
		"5h █████····· 50.2% (2h 28m)",
		"7d █████████· 90.0% (6d 23h)",
		"3p-5h ██▒······· 25.0% (30m)",
		"3p-7d ███████▒·· 75.0% (1d)",
	}
	for _, exp := range expectedLine2 {
		if !strings.Contains(lines[1], exp) {
			t.Errorf("Expected Line 2 to contain %q, got %q", exp, lines[1])
		}
	}
}
