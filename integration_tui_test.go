package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yuys13/agystatusline/renderer"
	"github.com/yuys13/agystatusline/tui"
	"github.com/yuys13/agystatusline/types"
	"github.com/yuys13/agystatusline/widgets"
)

// Helper to simulate key presses in Bubble Tea model
func sendKeys(m tui.Model, keys ...tea.KeyMsg) tui.Model {
	var curr tea.Model = m
	for _, k := range keys {
		var cmd tea.Cmd
		curr, cmd = curr.Update(k)
		_ = cmd
	}
	return curr.(tui.Model)
}

func enterKey() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyEnter}
}

func downKey() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyDown}
}

func escKey() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyEsc}
}

// TestIntegration_TUIModelInitAndSettingsPersistence tests TUI initialization, menu choices, and Save & Exit persistence.
func TestIntegration_TUIModelInitAndSettingsPersistence(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")

	initialSettings := types.DefaultSettings()
	initialSettings.Powerline.Enabled = false
	initialSettings.Powerline.Theme = "nord"
	initialSettings.ColorLevel = 2

	model := tui.NewModel(initialSettings, configPath)

	// 1. Toggle Powerline Mode: cursor 0 -> down -> enter (cursor = 1)
	model = sendKeys(model, downKey(), enterKey())

	// 2. Select Powerline Theme: cursor 1 -> down -> enter (cursor = 2, activeMenu = select_theme)
	// In select_theme menu: down -> enter (selects second theme, returns to main menu with cursor = 2)
	model = sendKeys(model, downKey(), enterKey(), downKey(), enterKey())

	// 3. Select Color Level: cursor 2 -> down (3) -> down (4) -> down (5) -> down (6) -> enter (activeMenu = select_color_level)
	// In select_color_level menu: down -> down -> enter (selects Truecolor 3, returns to main with cursor = 6)
	model = sendKeys(model, downKey(), downKey(), downKey(), downKey(), enterKey(), downKey(), downKey(), enterKey())

	// 4. Save & Exit: cursor 6 -> down (7) -> enter
	_ = sendKeys(model, downKey(), enterKey())

	// Check if file was saved to disk
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Expected saved settings file at %s, got error: %v", configPath, err)
	}

	var savedSettings types.Settings
	if err := json.Unmarshal(data, &savedSettings); err != nil {
		t.Fatalf("Failed to unmarshal saved settings: %v", err)
	}

	if !savedSettings.Powerline.Enabled {
		t.Errorf("Expected Powerline.Enabled to be toggled to true, got false")
	}
	if savedSettings.Powerline.Theme != "nord-aurora" {
		t.Errorf("Expected Powerline.Theme 'nord-aurora', got %q", savedSettings.Powerline.Theme)
	}
	if savedSettings.ColorLevel != 3 {
		t.Errorf("Expected ColorLevel to be 3 (Truecolor), got %d", savedSettings.ColorLevel)
	}
}

// TestIntegration_TUIPowerlinePreviewAndColorLevel tests TUI View() live preview rendering across menu updates.
func TestIntegration_TUIPowerlinePreviewAndColorLevel(t *testing.T) {
	widgets.RegisterAll()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "preview_settings.json")

	settings := types.DefaultSettings()
	settings.Powerline.Enabled = true
	settings.Powerline.Theme = "dracula"
	settings.ColorLevel = 3

	model := tui.NewModel(settings, configPath)

	// Get view string
	viewStr := model.View()

	if !strings.Contains(viewStr, "--- Live Preview ---") {
		t.Errorf("Expected View() to contain '--- Live Preview ---', got %q", viewStr)
	}

	// Move to Select Separator menu (cursor 0 -> down x3 -> enter)
	model = sendKeys(model, downKey(), downKey(), downKey(), enterKey())
	viewSubmenu := model.View()

	if !strings.Contains(viewSubmenu, "Select Powerline Separator") && !strings.Contains(viewSubmenu, "Separators") {
		t.Errorf("Expected submenu view to contain separator selection title, got %q", viewSubmenu)
	}
}

// TestIntegration_TUITier4TestDataSimulation tests E2E rendering with full test_data.json telemetry simulation.
func TestIntegration_TUITier4TestDataSimulation(t *testing.T) {
	widgets.RegisterAll()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sim_settings.json")

	// Construct full spec telemetry matching test_data.json
	trueVal := true
	sizeVal := 2000000.0
	inTokensVal := 14200.0
	outTokensVal := 850.0
	usedVal := 20.0
	remVal := 80.0

	q1Rem := 0.5019
	q1Reset := 8891.0
	q2Rem := 0.8500
	q2Reset := 125000.0
	q3Rem := 0.3000
	q3Reset := 3600.0
	q4Rem := 0.9000
	q4Reset := 50000.0

	telemetry := types.StatusJSON{
		HookEventName:  "status",
		SessionID:      "test-session-e2e-123",
		TranscriptPath: "/tmp/transcript.jsonl",
		CWD:            "/tmp/test-repo",
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
			"gemini-5h":     {RemainingFraction: &q1Rem, ResetInSeconds: &q1Reset},
			"gemini-weekly": {RemainingFraction: &q2Rem, ResetInSeconds: &q2Reset},
			"3p-5h":         {RemainingFraction: &q3Rem, ResetInSeconds: &q3Reset},
			"3p-weekly":     {RemainingFraction: &q4Rem, ResetInSeconds: &q4Reset},
		},
		Sandbox: &types.SandboxInfo{
			Enabled: &trueVal,
		},
		VCS: &types.VCSInfo{
			Type:   "git",
			Branch: "feature/e2e-test",
			Dirty:  &trueVal,
		},
		AgentState: "idle",
	}

	// Create custom settings with quota bar, model, git branch, sandbox
	simSettings := types.DefaultSettings()
	simSettings.Lines = [][]types.WidgetItem{
		{
			{ID: "m1", Type: "model", Color: "brightMagenta"},
			{ID: "g1", Type: "git-branch", Color: "brightBlue"},
			{ID: "qb1", Type: "quota-bar", Color: "brightGreen", Metadata: map[string]string{"key": "gemini-5h"}},
			{ID: "sb1", Type: "sandbox", Color: "yellow"},
		},
	}

	// Save settings via TUI model (cursor 0 -> down x7 -> enter)
	model := tui.NewModel(simSettings, configPath)
	_ = sendKeys(model, downKey(), downKey(), downKey(), downKey(), downKey(), downKey(), downKey(), enterKey())

	// Load settings back from file
	bytesData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved settings: %v", err)
	}

	var loadedSettings types.Settings
	_ = json.Unmarshal(bytesData, &loadedSettings)

	// Render statusline using telemetry context
	termWidth := 120
	ctx := types.RenderContext{
		Data:               telemetry,
		TerminalWidth:      &termWidth,
		IsPreview:          false,
		Minimalist:         false,
		GitCacheTTLSeconds: 5,
	}

	lines := renderer.RenderStatusLines(loadedSettings, ctx)
	if len(lines) == 0 {
		t.Fatalf("RenderStatusLines returned empty output")
	}

	plainOutput := renderer.StripAnsi(lines[0])

	if !strings.Contains(plainOutput, "Gemini 3.5 Flash") {
		t.Errorf("Expected telemetry model name 'Gemini 3.5 Flash', got %q", plainOutput)
	}
	if !strings.Contains(plainOutput, "feature/e2e-test*") {
		t.Errorf("Expected telemetry git branch 'feature/e2e-test*', got %q", plainOutput)
	}
	if !strings.Contains(plainOutput, "50.2%") && !strings.Contains(plainOutput, "50%") {
		t.Errorf("Expected quota bar percentage '50.2%%', got %q", plainOutput)
	}
	if !strings.Contains(plainOutput, "sandbox") {
		t.Errorf("Expected sandbox widget text in output, got %q", plainOutput)
	}
}

// TestIntegration_TUIWidgetLayoutEditingAndPersistence tests adding and editing widgets in TUI model.
func TestIntegration_TUIWidgetLayoutEditingAndPersistence(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "layout_settings.json")

	initialSettings := types.DefaultSettings()
	model := tui.NewModel(initialSettings, configPath)

	// 1. Enter "Edit Lines" (cursor = 0, enter) -> activeMenu = lines
	// 2. Select Line 1 (enter) -> activeMenu = items
	// 3. Add widget (press "a") -> activeMenu = add_widget
	// 4. Select first widget (enter) -> adds item, returns to items menu
	// 5. Esc back to lines, Esc back to main menu
	// 6. Navigate to Save & Exit (cursor 0 -> down x7 -> enter)
	_ = sendKeys(
		model,
		enterKey(),
		enterKey(),
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
		enterKey(),
		escKey(),
		escKey(),
		downKey(), downKey(), downKey(), downKey(), downKey(), downKey(), downKey(), enterKey(),
	)

	// Read saved settings
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read settings file: %v", err)
	}

	var savedSettings types.Settings
	_ = json.Unmarshal(data, &savedSettings)

	if len(savedSettings.Lines) == 0 {
		t.Fatalf("Expected non-empty Lines in saved settings")
	}
}
