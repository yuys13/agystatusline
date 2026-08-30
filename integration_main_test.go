package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/yuys13/agystatusline/renderer"
	"github.com/yuys13/agystatusline/types"
)

type failingReader struct{}

func (f *failingReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func standardTestJSON() string {
	trueVal := true
	sizeVal := 2000000.0
	inTokensVal := 14200.0
	outTokensVal := 850.0
	usedVal := 20.0
	remVal := 80.0
	remFracVal := 0.5019
	resetSecVal := 8891.0

	payload := types.StatusJSON{
		HookEventName: "status",
		SessionID:     "test-session-12345",
		CWD:           "/tmp/test-repo",
		Model: types.ModelInfo{
			ID:          "gemini-3.5-flash-medium",
			DisplayName: "Gemini 3.5 Flash (Medium)",
		},
		Workspace: &types.WorkspaceInfo{
			CurrentDir: "/tmp/test-repo",
			ProjectDir: "/tmp/test-repo",
		},
		Version: "1.0.0",
		ContextWindow: &types.ContextWindowInfo{
			ContextWindowSize:   &sizeVal,
			TotalInputTokens:    &inTokensVal,
			TotalOutputTokens:   &outTokensVal,
			UsedPercentage:      &usedVal,
			RemainingPercentage: &remVal,
		},
		Quota: map[string]types.QuotaInfo{
			"gemini-5h": {
				RemainingFraction: &remFracVal,
				ResetInSeconds:    &resetSecVal,
			},
		},
		Sandbox: &types.SandboxInfo{
			Enabled: &trueVal,
		},
		TerminalWidth: func(w int) *int { return &w }(120),
		AgentState:    "idle",
	}
	data, _ := json.Marshal(payload)
	return string(data)
}

func notTerminal(fd uintptr) bool {
	return false
}

// IT-01: --version flag
func TestRunMain_VersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader("")

	code := runMain([]string{"agystatusline", "--version"}, stdin, &stdout, &stderr, notTerminal)
	if code != 0 {
		t.Fatalf("Expected exit code 0 for --version, got %d", code)
	}

	out := stdout.String()
	if !strings.Contains(out, "version 1.0.0") {
		t.Errorf("Expected version output containing 'version 1.0.0', got %q", out)
	}
}

// IT-02: --help flag
func TestRunMain_HelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader("")

	code := runMain([]string{"agystatusline", "--help"}, stdin, &stdout, &stderr, notTerminal)
	if code != 0 {
		t.Fatalf("Expected exit code 0 for --help, got %d", code)
	}

	out := stdout.String()
	if !strings.Contains(out, "Usage: agystatusline") {
		t.Errorf("Expected usage output containing 'Usage: agystatusline', got %q", out)
	}
}

// IT-03: --hook flag
func TestRunMain_HookFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader("")

	code := runMain([]string{"agystatusline", "--hook"}, stdin, &stdout, &stderr, notTerminal)
	if code != 0 {
		t.Fatalf("Expected exit code 0 for --hook, got %d", code)
	}

	if stdout.Len() != 0 {
		t.Errorf("Expected empty stdout for --hook, got %q", stdout.String())
	}
}

// IT-04: Valid Stdin JSON pipeline
func TestRunMain_ValidStdinJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.toml")

	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader(standardTestJSON())

	code := runMain([]string{"agystatusline", "--config", configPath}, stdin, &stdout, &stderr, notTerminal)
	if code != 0 {
		t.Fatalf("Expected exit code 0 for valid JSON, got %d (stderr: %q)", code, stderr.String())
	}

	out := stdout.String()
	plainOut := renderer.StripAnsi(out)
	if !strings.Contains(plainOut, "Gemini 3.5 Flash") {
		t.Errorf("Expected rendered output containing model name 'Gemini 3.5 Flash', got %q", plainOut)
	}
	if !strings.Contains(out, "\x1b[0m") {
		t.Errorf("Expected ANSI reset escape sequence in stdout, got %q", out)
	}
}

// IT-05: Empty Stdin handling
func TestRunMain_EmptyStdin(t *testing.T) {
	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader("   \n  ")

	code := runMain([]string{"agystatusline"}, stdin, &stdout, &stderr, notTerminal)
	if code != 1 {
		t.Fatalf("Expected exit code 1 for empty stdin, got %d", code)
	}

	errOut := stderr.String()
	if !strings.Contains(errOut, "No input received") {
		t.Errorf("Expected stderr containing 'No input received', got %q", errOut)
	}
}

// IT-06: Invalid JSON format
func TestRunMain_InvalidJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader("{malformed_json: true,}")

	code := runMain([]string{"agystatusline"}, stdin, &stdout, &stderr, notTerminal)
	if code != 1 {
		t.Fatalf("Expected exit code 1 for invalid JSON, got %d", code)
	}

	errOut := stderr.String()
	if !strings.Contains(errOut, "Invalid status JSON format:") {
		t.Errorf("Expected stderr containing 'Invalid status JSON format:', got %q", errOut)
	}
}

// IT-07: Error reading stdin
func TestRunMain_ReadStdinError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	stdin := &failingReader{}

	code := runMain([]string{"agystatusline"}, stdin, &stdout, &stderr, notTerminal)
	if code != 1 {
		t.Fatalf("Expected exit code 1 for read stdin error, got %d", code)
	}

	errOut := stderr.String()
	if !strings.Contains(errOut, "Error reading stdin:") {
		t.Errorf("Expected stderr containing 'Error reading stdin:', got %q", errOut)
	}
}

// IT-08: Absolute Config path flag
func TestRunMain_ConfigFlagAbs(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "custom_settings.toml")

	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader(standardTestJSON())

	code := runMain([]string{"agystatusline", "--config", configPath}, stdin, &stdout, &stderr, notTerminal)
	if code != 0 {
		t.Fatalf("Expected exit code 0 for --config abs, got %d (stderr: %q)", code, stderr.String())
	}

	if settingsPath != configPath {
		t.Errorf("Expected settingsPath = %q, got %q", configPath, settingsPath)
	}
}

// IT-09: Relative Config path flag
func TestRunMain_ConfigFlagRel(t *testing.T) {
	tempDir := t.TempDir()
	relConfig := filepath.Join(tempDir, "rel_settings.toml")
	absConfig, _ := filepath.Abs(relConfig)

	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader(standardTestJSON())

	code := runMain([]string{"agystatusline", "--config", relConfig}, stdin, &stdout, &stderr, notTerminal)
	if code != 0 {
		t.Fatalf("Expected exit code 0 for --config rel, got %d (stderr: %q)", code, stderr.String())
	}

	if settingsPath != absConfig {
		t.Errorf("Expected settingsPath = %q, got %q", absConfig, settingsPath)
	}
}

// IT-10: Terminal width constraints handling
func TestRunMain_TerminalWidthHandling(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.toml")

	payload := types.StatusJSON{
		Model: types.ModelInfo{
			DisplayName: "Very Very Long Model Name That Exceeds Narrow Terminal Width",
		},
		TerminalWidth: func(w int) *int { return &w }(30),
	}
	data, _ := json.Marshal(payload)

	var stdout, stderr bytes.Buffer
	stdin := bytes.NewReader(data)

	code := runMain([]string{"agystatusline", "--config", configPath}, stdin, &stdout, &stderr, notTerminal)
	if code != 0 {
		t.Fatalf("Expected exit code 0, got %d (stderr: %q)", code, stderr.String())
	}

	out := stdout.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for i, l := range lines {
		plain := renderer.StripAnsi(l)
		visWidth := renderer.GetVisibleWidth(plain)
		if visWidth > 30 {
			t.Errorf("Line %d visible width %d exceeds terminal width limit 30: %q", i, visWidth, plain)
		}
	}
}

// IT-11: Minimalist mode rendering
func TestRunMain_MinimalistMode(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "minimalist_settings.toml")

	minSettings := types.DefaultSettings()
	minSettings.General.Minimalist = true
	bytesData, _ := toml.Marshal(minSettings)
	_ = os.WriteFile(configPath, bytesData, 0644)

	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader(standardTestJSON())

	code := runMain([]string{"agystatusline", "--config", configPath}, stdin, &stdout, &stderr, notTerminal)
	if code != 0 {
		t.Fatalf("Expected exit code 0 for minimalist mode, got %d (stderr: %q)", code, stderr.String())
	}

	plainOut := renderer.StripAnsi(stdout.String())
	if strings.Contains(plainOut, "Model:") {
		t.Errorf("Expected title prefix 'Model:' to be omitted in minimalist mode, got %q", plainOut)
	}
}

// IT-12: loadSettings failure in TTY mode
func TestRunMain_LoadSettingsFailureInTTY(t *testing.T) {
	tempDir := t.TempDir()
	conflictFile := filepath.Join(tempDir, "conflict")
	_ = os.WriteFile(conflictFile, []byte("file"), 0644)
	invalidConfigPath := filepath.Join(conflictFile, "sub", "settings.toml")

	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader("")

	code := runMain([]string{"agystatusline", "--config", invalidConfigPath}, stdin, &stdout, &stderr, func(fd uintptr) bool { return true })
	if code != 1 {
		t.Fatalf("Expected exit code 1 when loadSettings fails in TTY mode, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Failed to load settings:") {
		t.Errorf("Expected stderr containing 'Failed to load settings:', got %q", stderr.String())
	}
}

// IT-13: loadSettings failure in Piped Non-TTY mode
func TestRunMain_LoadSettingsFailureInPipedMode(t *testing.T) {
	tempDir := t.TempDir()
	conflictFile := filepath.Join(tempDir, "conflict")
	_ = os.WriteFile(conflictFile, []byte("file"), 0644)
	invalidConfigPath := filepath.Join(conflictFile, "sub", "settings.toml")

	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader(standardTestJSON())

	code := runMain([]string{"agystatusline", "--config", invalidConfigPath}, stdin, &stdout, &stderr, notTerminal)
	if code != 1 {
		t.Fatalf("Expected exit code 1 when loadSettings fails in Piped mode, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Failed to load settings:") {
		t.Errorf("Expected stderr containing 'Failed to load settings:', got %q", stderr.String())
	}
}

// IT-14: Tier 4 test_data.json Pipeline E2E simulation via runMain
func TestRunMain_TestDataJsonPipeline(t *testing.T) {
	testDataBytes, err := os.ReadFile("test_data.json")
	if err != nil {
		t.Fatalf("Failed to read test_data.json: %v", err)
	}

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "e2e_settings.toml")

	e2eSettings := types.DefaultSettings()
	e2eSettings.Lines = [][]types.WidgetItem{
		{
			{Type: "model"},
			{Type: "git-branch"},
		},
	}
	cfgBytes, _ := toml.Marshal(e2eSettings)
	_ = os.WriteFile(configPath, cfgBytes, 0644)

	var stdout, stderr bytes.Buffer
	stdin := bytes.NewReader(testDataBytes)

	code := runMain([]string{"agystatusline", "--config", configPath}, stdin, &stdout, &stderr, notTerminal)
	if code != 0 {
		t.Fatalf("Expected exit code 0 for test_data.json pipeline, got %d (stderr: %q)", code, stderr.String())
	}

	out := stdout.String()
	plain := renderer.StripAnsi(out)

	if !strings.Contains(plain, "Gemini 3.5 Flash") {
		t.Errorf("Expected output to contain 'Gemini 3.5 Flash', got %q", plain)
	}
	if !strings.Contains(plain, "feature/e2e-test*") {
		t.Errorf("Expected output to contain 'feature/e2e-test*', got %q", plain)
	}
	if !strings.Contains(out, "\x1b[0m") {
		t.Errorf("Expected ANSI reset escape sequence, got %q", out)
	}
}

// IT-15: Tier 4 Binary Execution E2E test (cat test_data.json | ./agystatusline simulation)
func TestBinaryExec_TestDataJsonPipeline(t *testing.T) {
	testDataBytes, err := os.ReadFile("test_data.json")
	if err != nil {
		t.Fatalf("Failed to read test_data.json: %v", err)
	}

	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "agystatusline_test_bin")

	// Build binary
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build agystatusline test binary: %v (out: %s)", err, string(out))
	}

	configPath := filepath.Join(binDir, "settings.toml")
	e2eSettings := types.DefaultSettings()
	e2eSettings.Lines = [][]types.WidgetItem{
		{
			{Type: "model"},
			{Type: "git-branch"},
		},
	}
	cfgBytes, _ := toml.Marshal(e2eSettings)
	_ = os.WriteFile(configPath, cfgBytes, 0644)

	runCmd := exec.Command(binPath, "--config", configPath)
	runCmd.Stdin = bytes.NewReader(testDataBytes)
	var stdout, stderr bytes.Buffer
	runCmd.Stdout = &stdout
	runCmd.Stderr = &stderr

	if err := runCmd.Run(); err != nil {
		t.Fatalf("Binary execution failed: %v (stderr: %q)", err, stderr.String())
	}

	out := stdout.String()
	plain := renderer.StripAnsi(out)

	if !strings.Contains(plain, "Gemini 3.5 Flash") {
		t.Errorf("Expected binary stdout to contain 'Gemini 3.5 Flash', got %q", plain)
	}
	if !strings.Contains(plain, "feature/e2e-test*") {
		t.Errorf("Expected binary stdout to contain 'feature/e2e-test*', got %q", plain)
	}
}

// IT-16: Tier 3 Cross-Feature Combination: Powerline + Git Telemetry + Context % + Quota
func TestRunMain_CrossFeatureCombination_PowerlineGitContext(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cross_settings.toml")

	pwSettings := types.DefaultSettings()
	pwSettings.Powerline.Enabled = true
	pwSettings.Powerline.Theme = "slant"
	pwSettings.General.ColorLevel = 3
	pwSettings.Lines = [][]types.WidgetItem{
		{
			{Type: "model", Color: "brightMagenta"},
			{Type: "git-branch", Color: "brightBlue"},
			{Type: "context-bar", Color: "brightYellow"},
			{Type: "quota-bar", Key: "gemini-5h", Color: "brightGreen"},
		},
	}
	bytesData, _ := toml.Marshal(pwSettings)
	_ = os.WriteFile(configPath, bytesData, 0644)

	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader(standardTestJSON())

	code := runMain([]string{"agystatusline", "--config", configPath}, stdin, &stdout, &stderr, notTerminal)
	if code != 0 {
		t.Fatalf("Expected exit code 0 for cross-feature test, got %d (stderr: %q)", code, stderr.String())
	}

	out := stdout.String()
	plain := renderer.StripAnsi(out)

	if !strings.Contains(plain, "Gemini 3.5 Flash") {
		t.Errorf("Expected model name, got %q", plain)
	}
	if !strings.Contains(plain, "20.0%") && !strings.Contains(plain, "20%") {
		t.Errorf("Expected context used percentage, got %q", plain)
	}
	if !strings.Contains(plain, "50.2%") && !strings.Contains(plain, "50%") {
		t.Errorf("Expected quota percentage, got %q", plain)
	}
}
