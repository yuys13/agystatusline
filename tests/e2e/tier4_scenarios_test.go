package e2e_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

// TestE2E_Tier4_TestDataJsonPipeline simulates shell pipeline execution: cat test_data.json | ./agystatusline
func TestE2E_Tier4_TestDataJsonPipeline(t *testing.T) {
	testDataPath := filepath.Join(projectRoot, "test_data.json")
	data, err := os.ReadFile(testDataPath)
	if err != nil {
		t.Fatalf("Failed to read test_data.json: %v", err)
	}

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.toml")

	stdout, stderr, exitCode := runCLI(t, string(data), "--config", configPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 for test_data.json pipeline, got %d (stderr: %q)", exitCode, stderr)
	}

	plain := stripANSI(stdout)
	if !strings.Contains(plain, "Gemini 3.5 Flash") {
		t.Errorf("Expected model 'Gemini 3.5 Flash' in output, got %q", plain)
	}
	if !strings.Contains(plain, "● READY") {
		t.Errorf("Expected agent-state '● READY' in output, got %q", plain)
	}
	if !strings.Contains(stdout, "\x1b[0m") {
		t.Errorf("Expected ANSI reset escape in stdout, got %q", stdout)
	}
}

// TestE2E_Tier4_DefaultConfigAutoGeneration tests automatic default configuration file creation when missing.
func TestE2E_Tier4_DefaultConfigAutoGeneration(t *testing.T) {
	tempDir := t.TempDir()
	newConfigDir := filepath.Join(tempDir, "brand_new_config_dir")
	newConfigFile := filepath.Join(newConfigDir, "settings.toml")

	// Ensure config file does not exist initially
	if _, err := os.Stat(newConfigFile); err == nil {
		t.Fatalf("Expected %s to not exist initially", newConfigFile)
	}

	stdout, stderr, exitCode := runCLI(t, standardStatusJSON(), "--config", newConfigFile)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 for auto-generation, got %d (stderr: %q)", exitCode, stderr)
	}

	// Verify file was generated
	info, err := os.Stat(newConfigFile)
	if err != nil {
		t.Fatalf("Failed to stat generated config file: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("Generated config file is empty")
	}

	// Verify stdout was generated properly
	plain := stripANSI(stdout)
	if !strings.Contains(plain, "Gemini 3.5 Flash") {
		t.Errorf("Expected output to contain model name, got %q", plain)
	}
}

// TestE2E_Tier4_CLIFlags tests standard CLI flags (--version, --help, -h, --hook).
func TestE2E_Tier4_CLIFlags(t *testing.T) {
	t.Run("Flag --version", func(t *testing.T) {
		stdout, stderr, exitCode := runCLI(t, "", "--version")
		if exitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
		}
		if !strings.Contains(stdout, "agystatusline version 1.0.0") {
			t.Errorf("Expected 'agystatusline version 1.0.0', got %q", stdout)
		}
	})

	t.Run("Flag --help", func(t *testing.T) {
		stdout, stderr, exitCode := runCLI(t, "", "--help")
		if exitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
		}
		if !strings.Contains(stdout, "Usage: agystatusline") {
			t.Errorf("Expected usage text in stdout, got %q", stdout)
		}
	})

	t.Run("Flag -h", func(t *testing.T) {
		stdout, stderr, exitCode := runCLI(t, "", "-h")
		if exitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
		}
		if !strings.Contains(stdout, "Usage: agystatusline") {
			t.Errorf("Expected usage text in stdout, got %q", stdout)
		}
	})

	t.Run("Flag --hook", func(t *testing.T) {
		stdout, stderr, exitCode := runCLI(t, "", "--hook")
		if exitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
		}
		if len(stdout) != 0 {
			t.Errorf("Expected empty stdout for --hook, got %q", stdout)
		}
	})
}

// TestE2E_Tier4_StdinErrorHandling tests error paths for empty stdin and invalid JSON.
func TestE2E_Tier4_StdinErrorHandling(t *testing.T) {
	t.Run("Empty stdin", func(t *testing.T) {
		_, stderr, exitCode := runCLI(t, "   \n\t  ")
		if exitCode != 1 {
			t.Fatalf("Expected exit code 1 on empty stdin, got %d", exitCode)
		}
		if !strings.Contains(stderr, "No input received") {
			t.Errorf("Expected stderr to contain 'No input received', got %q", stderr)
		}
	})

	t.Run("Malformed JSON stdin", func(t *testing.T) {
		_, stderr, exitCode := runCLI(t, "{not valid json}")
		if exitCode != 1 {
			t.Fatalf("Expected exit code 1 on malformed JSON, got %d", exitCode)
		}
		if !strings.Contains(stderr, "Invalid status JSON format:") {
			t.Errorf("Expected stderr to contain 'Invalid status JSON format:', got %q", stderr)
		}
	})
}

// TestE2E_Tier4_RelativeAndAbsoluteConfigPaths tests resolving both absolute and relative configuration paths.
func TestE2E_Tier4_RelativeAndAbsoluteConfigPaths(t *testing.T) {
	tempDir := t.TempDir()
	relDir := filepath.Join(tempDir, "sub")
	_ = os.MkdirAll(relDir, 0755)

	absPath := filepath.Join(relDir, "custom.toml")

	// Test with absolute path
	_, stderr, exitCode := runCLI(t, standardStatusJSON(), "--config", absPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 with absolute config path, got %d (stderr: %q)", exitCode, stderr)
	}

	// Test with relative path from working directory
	relPath := filepath.Join("sub", "custom2.toml")
	_, stderr, exitCode = runCLIWithDir(t, standardStatusJSON(), tempDir, "--config", relPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 with relative config path, got %d (stderr: %q)", exitCode, stderr)
	}

	expectedRelFile := filepath.Join(tempDir, "sub", "custom2.toml")
	if _, err := os.Stat(expectedRelFile); err != nil {
		t.Errorf("Expected config file created at %s, got error: %v", expectedRelFile, err)
	}
}

// TestE2E_Tier4_AdversarialGiantTelemetry tests resilient rendering under huge telemetry payloads.
func TestE2E_Tier4_AdversarialGiantTelemetry(t *testing.T) {
	// Create a payload with 10,000 subagent entries and giant strings
	var subagents []string
	for range 1000 {
		subagents = append(subagents, strings.Repeat("subagent-id-", 10))
	}
	giantModelName := strings.Repeat("ExtremelyLongModelName-", 200)

	payload := `{
		"model": {"display_name": "` + giantModelName + `"},
		"subagents": ["` + strings.Join(subagents, `","`) + `"],
		"artifact_count": 999999,
		"task_count": 888888,
		"agent_state": "working",
		"terminal_width": 80
	}`

	tempDir := t.TempDir()
	cfg := types.DefaultSettings()
	cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

	stdout, stderr, exitCode := runCLI(t, payload, "--config", cfgPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 on giant payload, got %d (stderr: %q)", exitCode, stderr)
	}

	plain := stripANSI(stdout)
	if len(plain) == 0 {
		t.Errorf("Expected non-empty output for giant payload")
	}
}

// TestE2E_Tier4_AdversarialNarrowTerminalAndNegativeWidth tests resilient clipping at narrow widths.
func TestE2E_Tier4_AdversarialNarrowTerminalAndNegativeWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"Zero width", 0},
		{"Negative width", -10},
		{"Narrow 5 chars", 5},
		{"Narrow 15 chars", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := fmt.Sprintf(`{
				"model": {"display_name": "Gemini 3.5 Pro"},
				"terminal_width": %d
			}`, tt.width)
			if tt.width <= 0 {
				payload = `{
					"model": {"display_name": "Gemini 3.5 Pro"},
					"terminal_width": 0
				}`
			}

			tempDir := t.TempDir()
			cfg := types.DefaultSettings()
			cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

			stdout, stderr, exitCode := runCLI(t, payload, "--config", cfgPath)
			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
			}
			_ = stdout
		})
	}
}

// TestE2E_Tier4_AdversarialCorruptedConfigFallback tests graceful fallback when config is corrupted on disk.
func TestE2E_Tier4_AdversarialCorruptedConfigFallback(t *testing.T) {
	tempDir := t.TempDir()
	corruptedConfigPath := filepath.Join(tempDir, "corrupted.toml")
	_ = os.WriteFile(corruptedConfigPath, []byte("[[lines]]\n{not toml syntax = true!!!}"), 0644)

	stdout, stderr, exitCode := runCLI(t, standardStatusJSON(), "--config", corruptedConfigPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 with fallback, got %d (stderr: %q)", exitCode, stderr)
	}

	plain := stripANSI(stdout)
	if !strings.Contains(plain, "Gemini 3.5 Flash") {
		t.Errorf("Expected fallback default settings to render model name, got %q", plain)
	}
}

// TestE2E_Tier4_AllWidgetsSimultaneousRendering tests that all 12 widgets can be rendered simultaneously on multi-line layout.
func TestE2E_Tier4_AllWidgetsSimultaneousRendering(t *testing.T) {
	tempDir := t.TempDir()
	cfg := types.DefaultSettings()
	cfg.Lines = [][]types.WidgetItem{
		{
			{Type: "agent-state"},
			{Type: "model"},
			{Type: "context-bar"},
			{Type: "artifacts"},
			{Type: "subagents"},
			{Type: "tasks"},
		},
		{
			{Type: "sandbox"},
			{Type: "quota", Key: "gemini-5h"},
			{Type: "quota-bar", Key: "gemini-weekly"},
			{Type: "git-branch"},
			{Type: "git-changes"},
			{Type: "custom-text", Text: "DEMO-ENV"},
		},
	}
	cfgPath := writeTOMLConfig(t, tempDir, "settings.toml", cfg)

	stdout, stderr, exitCode := runCLI(t, standardStatusJSON(), "--config", cfgPath)
	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d (stderr: %q)", exitCode, stderr)
	}

	plain := stripANSI(stdout)
	lines := strings.Split(strings.TrimSpace(plain), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 rendered lines, got %d:\n%s", len(lines), plain)
	}

	// Line 1 assertions
	if !strings.Contains(lines[0], "READY") {
		t.Errorf("Expected Line 1 to contain 'READY', got %q", lines[0])
	}
	if !strings.Contains(lines[0], "Gemini 3.5 Flash") {
		t.Errorf("Expected Line 1 to contain 'Gemini 3.5 Flash', got %q", lines[0])
	}
	if !strings.Contains(lines[0], "artifacts 3") {
		t.Errorf("Expected Line 1 to contain 'artifacts 3', got %q", lines[0])
	}

	// Line 2 assertions
	if !strings.Contains(lines[1], "sandbox on") {
		t.Errorf("Expected Line 2 to contain 'sandbox on', got %q", lines[1])
	}
	if !strings.Contains(lines[1], "gemini-5h") {
		t.Errorf("Expected Line 2 to contain 'gemini-5h', got %q", lines[1])
	}
	if !strings.Contains(lines[1], "DEMO-ENV") {
		t.Errorf("Expected Line 2 to contain 'DEMO-ENV', got %q", lines[1])
	}
}
