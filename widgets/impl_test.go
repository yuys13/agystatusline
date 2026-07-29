package widgets

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestModelWidget(t *testing.T) {
	RegisterAll()

	w := GetWidget("model")
	if w == nil {
		t.Fatalf("Model widget not found in registry")
	}

	if w.GetDefaultColor() != "brightMagenta" {
		t.Errorf("Expected default color 'brightMagenta', got '%s'", w.GetDefaultColor())
	}

	settings := types.DefaultSettings()
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID:          "claude-3-5-sonnet",
				DisplayName: "Claude 3.5 Sonnet",
			},
		},
	}

	item := types.WidgetItem{
		Type: "model",
	}

	// Normal render
	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "Claude 3.5 Sonnet" {
		t.Errorf("Expected title '' and body 'Claude 3.5 Sonnet', got title '%s' and body '%s'", title, output)
	}

	// RawValue render
	rawVal := true
	itemRaw := types.WidgetItem{
		Type:     "model",
		RawValue: &rawVal,
	}
	titleRaw, outputRaw, err := w.Render(itemRaw, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleRaw != "" || outputRaw != "Claude 3.5 Sonnet" {
		t.Errorf("Expected title '' and body 'Claude 3.5 Sonnet', got title '%s' and body '%s'", titleRaw, outputRaw)
	}

	// Test that parenthesized suffixes are kept as-is.
	ctxWithNew := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID:          "claude-3-5-sonnet-new",
				DisplayName: "Claude 3.5 Sonnet (New)",
			},
		},
	}
	titleNew, outputNew, err := w.Render(item, ctxWithNew, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleNew != "" || outputNew != "Claude 3.5 Sonnet (New)" {
		t.Errorf("Expected title '' and body 'Claude 3.5 Sonnet (New)', got title '%s' and body '%s'", titleNew, outputNew)
	}

	ctxWithMedium := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID:          "gemini-3.5-flash-medium",
				DisplayName: "Gemini 3.5 Flash (Medium)",
			},
		},
	}
	titleMedium, outputMedium, err := w.Render(item, ctxWithMedium, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleMedium != "" || outputMedium != "Gemini 3.5 Flash (Medium)" {
		t.Errorf("Expected title '' and body 'Gemini 3.5 Flash (Medium)', got title '%s' and body '%s'", titleMedium, outputMedium)
	}
}

func TestGitBranchWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("git-branch")
	if w == nil {
		t.Fatalf("Git branch widget not found")
	}
	if w.GetDefaultColor() != "brightMagenta" {
		t.Errorf("Expected default color 'brightMagenta', got '%s'", w.GetDefaultColor())
	}

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	runGitCommand = func(cmd string, ctx CwdResolver, ttl int) (string, error) {
		if cmd == "rev-parse --is-inside-work-tree" {
			return "true", nil
		}
		if cmd == "symbolic-ref --short HEAD" {
			return "feature/tdd", nil
		}
		return "", nil
	}

	settings := types.DefaultSettings()
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			CWD: "/dummy/repo",
		},
	}
	item := types.WidgetItem{Type: "git-branch"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "⎇ feature/tdd" {
		t.Errorf("Expected title '' and body '⎇ feature/tdd', got title '%s' and body '%s'", title, output)
	}
}

func TestGitChangesWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("git-changes")
	if w == nil {
		t.Fatalf("Git changes widget not found")
	}

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	runGitCommand = func(cmd string, ctx CwdResolver, ttl int) (string, error) {
		if cmd == "rev-parse --is-inside-work-tree" {
			return "true", nil
		}
		if cmd == "diff --shortstat" {
			return " 2 files changed, 10 insertions(+), 5 deletions(-)", nil
		}
		if cmd == "diff --cached --shortstat" {
			return " 1 file changed, 3 insertions(+)", nil
		}
		return "", nil
	}

	settings := types.DefaultSettings()
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			CWD: "/dummy/repo",
		},
	}
	item := types.WidgetItem{Type: "git-changes"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "(+13,-5)" {
		t.Errorf("Expected title '' and body '(+13,-5)', got title '%s' and body '%s'", title, output)
	}
}

func TestQuotaWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota")
	if w == nil {
		t.Fatalf("Quota widget not found")
	}

	if w.GetDefaultColor() != "brightWhite" {
		t.Errorf("Expected default color 'brightWhite', got '%s'", w.GetDefaultColor())
	}

	remaining1 := float64(0.5019274)
	resetInSecs1 := float64(8891)
	remaining2 := float64(1.0)
	resetInSecs2 := float64(17956)

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &remaining1,
					ResetTime:         "2026-06-20T11:27:27Z",
					ResetInSeconds:    &resetInSecs1,
				},
				"3p-5h": {
					RemainingFraction: &remaining2,
					ResetTime:         "2026-06-20T13:58:32Z",
					ResetInSeconds:    &resetInSecs2,
				},
			},
		},
	}
	settings := types.DefaultSettings()

	// Case 1: Labeled Percentage + Reset (default)
	item1 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
	}
	title1, output1, err := w.Render(item1, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title1 != "gemini-5h" || output1 != "50.19% (2h 28m)" {
		t.Errorf("Expected title 'gemini-5h' and body '50.19%% (2h 28m)', got title '%s' and body '%s'", title1, output1)
	}

	// Case 2: Raw Percentage + Reset (rawValue = true, default)
	rawVal := true
	item2 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
		RawValue: &rawVal,
	}
	title2, output2, err := w.Render(item2, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title2 != "" || output2 != "50.19% (2h 28m)" {
		t.Errorf("Expected title '' and body '50.19%% (2h 28m)', got title '%s' and body '%s'", title2, output2)
	}

	// Case 3: Custom Text label + Reset
	item3 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
		CustomText: "Gemini Q",
	}
	title3, output3, err := w.Render(item3, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title3 != "Gemini Q" || output3 != "50.19% (2h 28m)" {
		t.Errorf("Expected title 'Gemini Q' and body '50.19%% (2h 28m)', got title '%s' and body '%s'", title3, output3)
	}

	// Case 3b: display="quota" (Percentage only, labeled)
	itemQuotaOnly := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "quota",
		},
	}
	titleQuotaOnly, outputQuotaOnly, err := w.Render(itemQuotaOnly, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleQuotaOnly != "gemini-5h" || outputQuotaOnly != "50.19%" {
		t.Errorf("Expected title 'gemini-5h' and body '50.19%%', got title '%s' and body '%s'", titleQuotaOnly, outputQuotaOnly)
	}

	// Case 3c: display="quota" (Percentage only, raw)
	itemQuotaOnlyRaw := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "quota",
		},
		RawValue: &rawVal,
	}
	titleQuotaOnlyRaw, outputQuotaOnlyRaw, err := w.Render(itemQuotaOnlyRaw, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleQuotaOnlyRaw != "" || outputQuotaOnlyRaw != "50.19%" {
		t.Errorf("Expected title '' and body '50.19%%', got title '%s' and body '%s'", titleQuotaOnlyRaw, outputQuotaOnlyRaw)
	}

	// Case 4: Reset time labeled
	item4 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "reset",
		},
	}
	title4, output4, err := w.Render(item4, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title4 != "gemini-5h (reset)" || output4 != "2h 28m" {
		t.Errorf("Expected title 'gemini-5h (reset)' and body '2h 28m', got title '%s' and body '%s'", title4, output4)
	}

	// Case 5: Reset time raw
	item5 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title5, output5, err := w.Render(item5, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title5 != "" || output5 != "2h 28m" {
		t.Errorf("Expected title '' and body '2h 28m', got title '%s' and body '%s'", title5, output5)
	}

	// Case 6: Reset time other durations
	secs45 := float64(45)
	secs125 := float64(125)
	secs567440 := float64(567440)
	ctxDur := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"secs45": {
					RemainingFraction: &remaining2,
					ResetInSeconds:    &secs45,
				},
				"secs125": {
					RemainingFraction: &remaining2,
					ResetInSeconds:    &secs125,
				},
				"secs567440": {
					RemainingFraction: &remaining2,
					ResetInSeconds:    &secs567440,
				},
			},
		},
	}
	itemSecs45 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "secs45",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title45, output45, _ := w.Render(itemSecs45, ctxDur, settings)
	if title45 != "" || output45 != "45s" {
		t.Errorf("Expected title '' and body '45s', got title '%s' and body '%s'", title45, output45)
	}

	itemSecs125 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "secs125",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title125, output125, _ := w.Render(itemSecs125, ctxDur, settings)
	if title125 != "" || output125 != "2m 5s" {
		t.Errorf("Expected title '' and body '2m 5s', got title '%s' and body '%s'", title125, output125)
	}

	itemSecs567440 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "secs567440",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title567440, output567440, _ := w.Render(itemSecs567440, ctxDur, settings)
	if title567440 != "" || output567440 != "6d 13h" {
		t.Errorf("Expected title '' and body '6d 13h', got title '%s' and body '%s'", title567440, output567440)
	}

	// Case 7: Key not found or empty
	itemEmpty := types.WidgetItem{
		Type: "quota",
	}
	titleEmpty, outputEmpty, _ := w.Render(itemEmpty, ctx, settings)
	if titleEmpty != "" || outputEmpty != "" {
		t.Errorf("Expected empty title/body for empty key, got title '%s' and body '%s'", titleEmpty, outputEmpty)
	}

	itemInvalid := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "invalid-key",
		},
	}
	titleInvalid, outputInvalid, _ := w.Render(itemInvalid, ctx, settings)
	if titleInvalid != "" || outputInvalid != "" {
		t.Errorf("Expected empty title/body for invalid key, got title '%s' and body '%s'", titleInvalid, outputInvalid)
	}

	// Case 8: Quota map is nil
	ctxNilQuota := types.RenderContext{
		Data: types.StatusJSON{},
	}
	titleNil, outputNil, _ := w.Render(item1, ctxNilQuota, settings)
	if titleNil != "" || outputNil != "" {
		t.Errorf("Expected empty title/body for nil quota map, got title '%s' and body '%s'", titleNil, outputNil)
	}
}

func TestSandboxWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("sandbox")
	if w == nil {
		t.Fatalf("Sandbox widget not found in registry")
	}

	if w.GetDefaultColor() != "yellow" {
		t.Errorf("Expected default color 'yellow', got '%s'", w.GetDefaultColor())
	}

	if w.GetDisplayName() != "Sandbox" {
		t.Errorf("Expected display name 'Sandbox', got '%s'", w.GetDisplayName())
	}

	settings := types.DefaultSettings()

	// Case 1: Sandbox info is nil
	ctxNil := types.RenderContext{
		Data: types.StatusJSON{},
	}
	item := types.WidgetItem{Type: "sandbox"}
	titleNil, outNil, err := w.Render(item, ctxNil, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleNil != "" || outNil != "" {
		t.Errorf("Expected empty title/body when sandbox is nil, got title '%s' and body '%s'", titleNil, outNil)
	}

	// Case 2: Sandbox.Enabled is nil
	ctxNilEnabled := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{},
		},
	}
	titleNilEnabled, outNilEnabled, err := w.Render(item, ctxNilEnabled, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleNilEnabled != "" || outNilEnabled != "" {
		t.Errorf("Expected empty title/body when sandbox.enabled is nil, got title '%s' and body '%s'", titleNilEnabled, outNilEnabled)
	}

	// Case 3: Sandbox.Enabled is true (normal and raw)
	trueVal := true
	ctxTrue := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{
				Enabled: &trueVal,
			},
		},
	}
	titleTrue, outTrue, err := w.Render(item, ctxTrue, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleTrue != "sandbox" || outTrue != "on" {
		t.Errorf("Expected title 'sandbox' and body 'on', got title '%s' and body '%s'", titleTrue, outTrue)
	}

	itemRaw := types.WidgetItem{Type: "sandbox", RawValue: &trueVal}
	titleTrueRaw, outTrueRaw, err := w.Render(itemRaw, ctxTrue, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleTrueRaw != "" || outTrueRaw != "on" {
		t.Errorf("Expected title '' and body 'on', got title '%s' and body '%s'", titleTrueRaw, outTrueRaw)
	}

	// Case 4: Sandbox.Enabled is false (normal and raw)
	falseVal := false
	ctxFalse := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{
				Enabled: &falseVal,
			},
		},
	}
	titleFalse, outFalse, err := w.Render(item, ctxFalse, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleFalse != "sandbox" || outFalse != "off" {
		t.Errorf("Expected title 'sandbox' and body 'off', got title '%s' and body '%s'", titleFalse, outFalse)
	}

	titleFalseRaw, outFalseRaw, err := w.Render(itemRaw, ctxFalse, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleFalseRaw != "" || outFalseRaw != "off" {
		t.Errorf("Expected title '' and body 'off', got title '%s' and body '%s'", titleFalseRaw, outFalseRaw)
	}
}

func TestAgentStateWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("agent-state")
	if w == nil {
		t.Fatalf("Agent state widget not found")
	}

	settings := types.DefaultSettings()
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			AgentState: "thinking",
		},
	}
	item := types.WidgetItem{Type: "agent-state"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "◆ THINKING" {
		t.Errorf("Expected body '◆ THINKING', got title '%s' and body '%s'", title, output)
	}
}

func TestContextBarWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("context-bar")
	if w == nil {
		t.Fatalf("Context bar widget not found")
	}

	settings := types.DefaultSettings()
	pct := 50.0
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			ContextWindow: &types.ContextWindowInfo{
				UsedPercentage: &pct,
			},
		},
	}
	item := types.WidgetItem{Type: "context-bar"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "ctx" || !strings.Contains(output, "50.0%") {
		t.Errorf("Expected title 'ctx' and body containing '50.0%%', got title '%s' and body '%s'", title, output)
	}
}

func TestArtifactsWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("artifacts")
	if w == nil {
		t.Fatalf("Artifacts widget not found")
	}

	settings := types.DefaultSettings()
	count := 5
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			ArtifactCount: &count,
		},
	}
	item := types.WidgetItem{Type: "artifacts"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "artifacts" || output != "5" {
		t.Errorf("Expected title 'artifacts' and body '5', got title '%s' and body '%s'", title, output)
	}
}

func TestSubagentsWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("subagents")
	if w == nil {
		t.Fatalf("Subagents widget not found")
	}

	settings := types.DefaultSettings()
	count := 3
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Subagents: float64(count),
		},
	}
	item := types.WidgetItem{Type: "subagents"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "subagents" || output != "3" {
		t.Errorf("Expected title 'subagents' and body '3', got title '%s' and body '%s'", title, output)
	}
}

func TestTasksWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("tasks")
	if w == nil {
		t.Fatalf("Tasks widget not found")
	}

	settings := types.DefaultSettings()
	count := 2
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			TaskCount: &count,
		},
	}
	item := types.WidgetItem{Type: "tasks"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "tasks" || output != "2" {
		t.Errorf("Expected title 'tasks' and body '2', got title '%s' and body '%s'", title, output)
	}
}

func TestQuotaBarWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota-bar")
	if w == nil {
		t.Fatalf("Quota bar widget not found")
	}

	settings := types.DefaultSettings()

	// Test 1: No quota data
	ctxNoData := types.RenderContext{
		Data: types.StatusJSON{},
	}
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}
	title, output, err := w.Render(item, ctxNoData, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "" {
		t.Errorf("Expected empty output for missing quota data, got title %q and body %q", title, output)
	}

	// Test 2: Standard rendering with different percentages
	pcts := []float64{0.8, 0.3, 0.05} // 80%, 30%, 5%
	expectedColors := []string{"brightGreen", "brightYellow", "brightRed"}

	for i, pct := range pcts {
		ctx := types.RenderContext{
			Data: types.StatusJSON{
				Quota: map[string]types.QuotaInfo{
					"gemini-5h": {
						RemainingFraction: &pct,
					},
				},
			},
		}

		// Verify GetBodyColor
		color := w.GetBodyColor(item, ctx)
		if color != expectedColors[i] {
			t.Errorf("For pct %.1f, expected body color %s, got %s", pct*100, expectedColors[i], color)
		}

		// Verify normal Render
		title, output, err = w.Render(item, ctx, settings)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}
		expectedPctStr := fmt.Sprintf("%.1f%%", pct*100)
		if title != "5h" || !strings.Contains(output, expectedPctStr) {
			t.Errorf("Expected title '5h' and body containing %q, got title %q and body %q", expectedPctStr, title, output)
		}
		// Verify gemini-weekly maps to '7d'
		itemWeekly := types.WidgetItem{
			Type:     "quota-bar",
			Metadata: map[string]string{"key": "gemini-weekly"},
		}
		ctxWeekly := types.RenderContext{
			Data: types.StatusJSON{
				Quota: map[string]types.QuotaInfo{
					"gemini-weekly": {
						RemainingFraction: &pct,
					},
				},
			},
		}
		titleW, _, err := w.Render(itemWeekly, ctxWeekly, settings)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}
		if titleW != "7d" {
			t.Errorf("Expected title '7d', got %q", titleW)
		}

		// Verify 3p-weekly maps to '3p-7d'
		item3PWeekly := types.WidgetItem{
			Type:     "quota-bar",
			Metadata: map[string]string{"key": "3p-weekly"},
		}
		ctx3PWeekly := types.RenderContext{
			Data: types.StatusJSON{
				Quota: map[string]types.QuotaInfo{
					"3p-weekly": {
						RemainingFraction: &pct,
					},
				},
			},
		}
		title3PW, _, err := w.Render(item3PWeekly, ctx3PWeekly, settings)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}
		if title3PW != "3p-7d" {
			t.Errorf("Expected title '3p-7d', got %q", title3PW)
		}

		// Verify RawValue
		itemRaw := types.WidgetItem{
			Type:     "quota-bar",
			Metadata: map[string]string{"key": "gemini-5h"},
			RawValue: func(b bool) *bool { return &b }(true),
		}
		titleR, outputR, err := w.Render(itemRaw, ctx, settings)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}
		if titleR != "" || !strings.Contains(outputR, expectedPctStr) {
			t.Errorf("Expected empty title and body containing %q, got title %q and body %q", expectedPctStr, titleR, outputR)
		}
	}

	// Test 3: Specific boundaries (50% should be green, 49% yellow, 10% yellow, 9% red)
	boundaryTests := []struct {
		fraction float64
		expected string
	}{
		{0.50, "brightGreen"},
		{0.49, "brightYellow"},
		{0.10, "brightYellow"},
		{0.09, "brightRed"},
	}
	for _, tc := range boundaryTests {
		ctx := types.RenderContext{
			Data: types.StatusJSON{
				Quota: map[string]types.QuotaInfo{
					"gemini-5h": {
						RemainingFraction: &tc.fraction,
					},
				},
			},
		}
		color := w.GetBodyColor(item, ctx)
		if color != tc.expected {
			t.Errorf("For boundary fraction %.2f, expected color %s, got %s", tc.fraction, tc.expected, color)
		}
	}

	// Test 4: Reset time inclusion
	resetSecs := 750.0 // 12m 30s
	ctxWithReset := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: func(f float64) *float64 { return &f }(0.5019),
					ResetInSeconds:    &resetSecs,
				},
			},
		},
	}
	itemWithReset := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}
	titleReset, outputReset, err := w.Render(itemWithReset, ctxWithReset, settings)
	if err != nil {
		t.Fatalf("Render error with reset: %v", err)
	}
	if titleReset != "5h" {
		t.Errorf("Expected title '5h', got %q", titleReset)
	}
	expectedOutput := "███████▒······· 50.2% (12m 30s)"
	if outputReset != expectedOutput {
		t.Errorf("Expected body %q, got %q", expectedOutput, outputReset)
	}
}

func TestModelWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("model")
	settings := types.DefaultSettings()

	// 1. Empty Model
	ctxEmpty := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "model"}
	title, body, err := w.Render(item, ctxEmpty, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty output for empty model, got title=%q, body=%q, err=%v", title, body, err)
	}

	// 2. ID only (no DisplayName)
	ctxIDOnly := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{ID: "claude-3-5-haiku"},
		},
	}
	_, bodyID, _ := w.Render(item, ctxIDOnly, settings)
	if bodyID != "claude-3-5-haiku" {
		t.Errorf("Expected fallback to ID 'claude-3-5-haiku', got %q", bodyID)
	}

	// 3. PreserveColors
	trueVal := true
	itemPreserve := types.WidgetItem{Type: "model", PreserveColors: &trueVal}
	_, bodyPreserve, _ := w.Render(itemPreserve, ctxIDOnly, settings)
	if !strings.Contains(bodyPreserve, "\x1b[95m") {
		t.Errorf("Expected ANSI colors in preserve mode, got %q", bodyPreserve)
	}
}

func TestGitWidgets_EdgeCases(t *testing.T) {
	RegisterAll()
	wBranch := GetWidget("git-branch")
	wChanges := GetWidget("git-changes")
	settings := types.DefaultSettings()

	// Mock git fail
	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()
	runGitCommand = func(cmd string, ctx CwdResolver, ttl int) (string, error) {
		return "", fmt.Errorf("git not installed")
	}

	ctx := types.RenderContext{Data: types.StatusJSON{CWD: "/no-git-dir"}}

	// GitBranch with hide = true
	hideTrue := true
	itemHide := types.WidgetItem{Type: "git-branch", Hide: &hideTrue}
	_, bodyBranchHide, _ := wBranch.Render(itemHide, ctx, settings)
	if bodyBranchHide != "" {
		t.Errorf("Expected empty string when git missing and hide=true, got %q", bodyBranchHide)
	}

	// GitBranch with hide = false
	hideFalse := false
	itemNoHide := types.WidgetItem{Type: "git-branch", Hide: &hideFalse}
	_, bodyBranchNoHide, _ := wBranch.Render(itemNoHide, ctx, settings)
	if bodyBranchNoHide != "⎇ no git" {
		t.Errorf("Expected '⎇ no git', got %q", bodyBranchNoHide)
	}

	// GitBranch GetBodyColor with Dirty
	ctxDirty := types.RenderContext{
		Data: types.StatusJSON{
			VCS: &types.VCSInfo{Dirty: &hideTrue},
		},
	}
	if wBranch.GetBodyColor(itemNoHide, ctxDirty) != "brightRed" {
		t.Errorf("Expected brightRed for dirty git branch")
	}

	// GitChanges with hide = true
	itemChangesHide := types.WidgetItem{Type: "git-changes", Hide: &hideTrue}
	_, bodyChangesHide, _ := wChanges.Render(itemChangesHide, ctx, settings)
	if bodyChangesHide != "" {
		t.Errorf("Expected empty string for git-changes when git missing and hide=true, got %q", bodyChangesHide)
	}

	// parseShortStat edge cases
	ins, del := parseShortStat("")
	if ins != 0 || del != 0 {
		t.Errorf("Expected 0, 0 for empty stat, got %d, %d", ins, del)
	}

	insOnly, delOnly := parseShortStat("5 insertions(+)")
	if insOnly != 5 || delOnly != 0 {
		t.Errorf("Expected 5, 0 for insertion only, got %d, %d", insOnly, delOnly)
	}

	insDel, delDel := parseShortStat("10 deletions(-)")
	if insDel != 0 || delDel != 10 {
		t.Errorf("Expected 0, 10 for deletion only, got %d, %d", insDel, delDel)
	}
}

func TestFormatResetInSeconds(t *testing.T) {
	tests := []struct {
		secs     *float64
		expected string
	}{
		{nil, ""},
		{func(f float64) *float64 { return &f }(-10), "0s"},
		{func(f float64) *float64 { return &f }(0), "0s"},
		{func(f float64) *float64 { return &f }(45), "45s"},
		{func(f float64) *float64 { return &f }(60), "1m"},
		{func(f float64) *float64 { return &f }(90), "1m 30s"},
		{func(f float64) *float64 { return &f }(3600), "1h"},
		{func(f float64) *float64 { return &f }(3660), "1h 1m"},
		{func(f float64) *float64 { return &f }(86400), "1d"},
		{func(f float64) *float64 { return &f }(90000), "1d 1h"},
	}

	for _, tc := range tests {
		res := formatResetInSeconds(tc.secs)
		if res != tc.expected {
			t.Errorf("For secs=%v expected %q, got %q", tc.secs, tc.expected, res)
		}
	}
}

func TestQuotaWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota")
	settings := types.DefaultSettings()

	// Nil Quota
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h"}}
	title, body, err := w.Render(item, ctxNil, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result for nil Quota map")
	}

	// Missing Key in Metadata
	ctxData := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {RemainingFraction: func(f float64) *float64 { return &f }(0.8)},
			},
		},
	}
	itemNoKey := types.WidgetItem{Type: "quota"}
	_, bodyNoKey, _ := w.Render(itemNoKey, ctxData, settings)
	if bodyNoKey != "" {
		t.Errorf("Expected empty body when metadata key is missing")
	}

	// Unknown Key
	itemBadKey := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "unknown"}}
	_, bodyBadKey, _ := w.Render(itemBadKey, ctxData, settings)
	if bodyBadKey != "" {
		t.Errorf("Expected empty body for unknown quota key")
	}

	// Display mode "quota"
	itemDisplayQuota := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h", "display": "quota"}}
	titleQ, bodyQ, _ := w.Render(itemDisplayQuota, ctxData, settings)
	if titleQ != "gemini-5h" || bodyQ != "80.00%" {
		t.Errorf("Expected 'gemini-5h' and '80.00%%', got title=%q body=%q", titleQ, bodyQ)
	}

	// Display mode "reset"
	itemDisplayReset := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h", "display": "reset"}}
	_, bodyReset, _ := w.Render(itemDisplayReset, ctxData, settings)
	if bodyReset != "" {
		t.Errorf("Expected empty reset body when ResetInSeconds is nil")
	}
}

func TestSandboxWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("sandbox")
	settings := types.DefaultSettings()

	// Sandbox nil
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "sandbox"}
	_, bodyNil, _ := w.Render(item, ctxNil, settings)
	if bodyNil != "" {
		t.Errorf("Expected empty body when sandbox is nil")
	}

	// Sandbox off
	falseVal := false
	ctxOff := types.RenderContext{
		Data: types.StatusJSON{
			Sandbox: &types.SandboxInfo{Enabled: &falseVal},
		},
	}
	titleOff, bodyOff, _ := w.Render(item, ctxOff, settings)
	if titleOff != "sandbox" || bodyOff != "off" {
		t.Errorf("Expected 'sandbox' and 'off', got %q, %q", titleOff, bodyOff)
	}
	if w.GetBodyColor(item, ctxOff) != "brightBlack" {
		t.Errorf("Expected brightBlack for sandbox off")
	}

	// Preserve colors
	trueVal := true
	itemPreserve := types.WidgetItem{Type: "sandbox", PreserveColors: &trueVal}
	_, bodyPreserve, _ := w.Render(itemPreserve, ctxOff, settings)
	if !strings.Contains(bodyPreserve, "sandbox off") {
		t.Errorf("Expected preserve colors text to contain 'sandbox off'")
	}
}

func TestAgentStateWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("agent-state")
	settings := types.DefaultSettings()

	states := []struct {
		state         string
		expectedColor string
		expectedText  string
	}{
		{"", "brightGreen", "● READY"},
		{"thinking", "brightYellow", "◆ THINKING"},
		{"working", "brightCyan", "⚙ WORKING"},
		{"tool_use", "brightMagenta", "🔧 TOOL"},
		{"custom_state", "white", "⏳ CUSTOM_STATE"},
	}

	for _, tc := range states {
		ctx := types.RenderContext{
			Data: types.StatusJSON{AgentState: tc.state},
		}
		item := types.WidgetItem{Type: "agent-state"}

		color := w.GetBodyColor(item, ctx)
		if color != tc.expectedColor {
			t.Errorf("For state %q expected color %s, got %s", tc.state, tc.expectedColor, color)
		}

		_, body, _ := w.Render(item, ctx, settings)
		if body != tc.expectedText {
			t.Errorf("For state %q expected text %q, got %q", tc.state, tc.expectedText, body)
		}
	}
}

func TestContextBarWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("context-bar")
	settings := types.DefaultSettings()

	// Nil context window
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "context-bar"}
	title, body, _ := w.Render(item, ctxNil, settings)
	if title != "ctx" || body != "" {
		t.Errorf("Expected 'ctx' and empty body when ContextWindow is nil, got %q, %q", title, body)
	}

	// High percentage colors
	pctHigh := 95.0
	ctxHigh := types.RenderContext{
		Data: types.StatusJSON{
			ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pctHigh},
		},
	}
	if w.GetBodyColor(item, ctxHigh) != "brightRed" {
		t.Errorf("Expected brightRed for 95%% context bar")
	}

	pctMid := 65.0
	ctxMid := types.RenderContext{
		Data: types.StatusJSON{
			ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pctMid},
		},
	}
	if w.GetBodyColor(item, ctxMid) != "brightYellow" {
		t.Errorf("Expected brightYellow for 65%% context bar")
	}
}

func TestSubagentsWidget_Types(t *testing.T) {
	RegisterAll()
	w := GetWidget("subagents")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "subagents"}

	// Array type
	ctxSlice := types.RenderContext{
		Data: types.StatusJSON{Subagents: []any{"a1", "a2"}},
	}
	_, bodySlice, _ := w.Render(item, ctxSlice, settings)
	if bodySlice != "2" {
		t.Errorf("Expected '2' subagents, got %q", bodySlice)
	}

	// Number type
	ctxNum := types.RenderContext{
		Data: types.StatusJSON{Subagents: float64(5)},
	}
	_, bodyNum, _ := w.Render(item, ctxNum, settings)
	if bodyNum != "5" {
		t.Errorf("Expected '5' subagents, got %q", bodyNum)
	}
}

func TestTasksWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("tasks")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "tasks"}

	// Nil task count
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	title, body, _ := w.Render(item, ctxNil, settings)
	if title != "tasks" || body != "0" {
		t.Errorf("Expected 'tasks' and '0' for nil TaskCount, got %q, %q", title, body)
	}

	// Valid task count
	taskCount := 3
	ctxVal := types.RenderContext{Data: types.StatusJSON{TaskCount: &taskCount}}
	_, bodyVal, _ := w.Render(item, ctxVal, settings)
	if bodyVal != "3" {
		t.Errorf("Expected '3' for TaskCount=3, got %q", bodyVal)
	}
}

func TestWidgetInterfaces(t *testing.T) {
	RegisterAll()
	ctx := types.RenderContext{
		Data: types.StatusJSON{},
	}
	item := types.WidgetItem{
		Type: "model",
	}

	for _, name := range []string{
		"model", "git-branch", "git-changes", "quota", "custom-text",
		"sandbox", "agent-state", "context-bar", "quota-bar", "artifacts",
		"subagents", "tasks",
	} {
		w := GetWidget(name)
		if w == nil {
			t.Fatalf("Widget %s not registered", name)
		}
		if nameStr := w.GetDisplayName(); nameStr == "" {
			t.Errorf("GetDisplayName() returned empty for %s", name)
		}
		if defaultColor := w.GetDefaultColor(); defaultColor == "" {
			t.Errorf("GetDefaultColor() returned empty for %s", name)
		}
		if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
			t.Errorf("GetBodyColor() returned empty for %s", name)
		}
	}
}

func TestModelWidget_IDOnlyAndEmpty(t *testing.T) {
	RegisterAll()
	w := GetWidget("model")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "model"}

	// DisplayName empty, ID present
	ctxIDOnly := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{ID: "gemini-flash"},
		},
	}
	_, body, err := w.Render(item, ctxIDOnly, settings)
	if err != nil || body != "gemini-flash" {
		t.Errorf("Expected 'gemini-flash', got body=%q, err=%v", body, err)
	}

	// Both DisplayName and ID empty
	ctxEmpty := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{},
		},
	}
	titleEmpty, bodyEmpty, err := w.Render(item, ctxEmpty, settings)
	if err != nil || titleEmpty != "" || bodyEmpty != "" {
		t.Errorf("Expected empty title/body for empty model info, got title=%q, body=%q", titleEmpty, bodyEmpty)
	}
}

func TestQuotaWidget_MoreEdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota")
	settings := types.DefaultSettings()

	// Missing metadata key
	itemNoKey := types.WidgetItem{Type: "quota"}
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"api": {ResetTime: "2026-06-20T12:00:00Z"},
			},
		},
	}
	title, body, err := w.Render(itemNoKey, ctx, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result when metadata key is missing, got title=%q, body=%q", title, body)
	}

	// Key not in Quota map
	itemKey := types.WidgetItem{
		Type:     "quota",
		Metadata: map[string]string{"key": "nonexistent"},
	}
	title, body, err = w.Render(itemKey, ctx, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result when key is non-existent, got title=%q, body=%q", title, body)
	}

	// QuotaInfo nil fields (RemainingFraction nil, ResetTime empty)
	itemApi := types.WidgetItem{
		Type:     "quota",
		Metadata: map[string]string{"key": "api", "display": "quota"},
	}
	ctxNilFields := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"api": {},
			},
		},
	}
	title, body, err = w.Render(itemApi, ctxNilFields, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result when QuotaInfo has nil fields, got title=%q, body=%q", title, body)
	}
}

func TestQuotaBarWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota-bar")
	settings := types.DefaultSettings()

	// Nil Quota map
	ctxNil := types.RenderContext{Data: types.StatusJSON{Quota: nil}}
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "api"},
	}
	title, body, err := w.Render(item, ctxNil, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result for nil quota map, got title=%q, body=%q", title, body)
	}
	if color := w.GetBodyColor(item, ctxNil); color != "brightGreen" {
		t.Errorf("Expected default 'brightGreen' body color for nil quota, got %q", color)
	}

	// Test color thresholds (RemainingFraction: 80% -> brightGreen, 30% -> brightYellow, 5% -> brightRed)
	val80 := 0.8
	val30 := 0.3
	val05 := 0.05

	ctx80 := types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"api": {RemainingFraction: &val80}}}}
	ctx30 := types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"api": {RemainingFraction: &val30}}}}
	ctx05 := types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"api": {RemainingFraction: &val05}}}}

	if c := w.GetBodyColor(item, ctx80); c != "brightGreen" {
		t.Errorf("Expected brightGreen for 80%%, got %q", c)
	}
	if c := w.GetBodyColor(item, ctx30); c != "brightYellow" {
		t.Errorf("Expected brightYellow for 30%%, got %q", c)
	}
	if c := w.GetBodyColor(item, ctx05); c != "brightRed" {
		t.Errorf("Expected brightRed for 5%%, got %q", c)
	}
}

func TestSubagentsWidget_InvalidType(t *testing.T) {
	RegisterAll()
	w := GetWidget("subagents")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "subagents"}

	// Subagents as unexpected string type
	ctxString := types.RenderContext{
		Data: types.StatusJSON{
			Subagents: "invalid_string_data",
		},
	}
	title, body, err := w.Render(item, ctxString, settings)
	if err != nil || title != "subagents" || body != "0" {
		t.Errorf("Expected 'subagents' and '0' for string type Subagents, got title=%q, body=%q", title, body)
	}
}

func TestGitBranchWidget_NonGit(t *testing.T) {
	RegisterAll()
	w := GetWidget("git-branch")
	settings := types.DefaultSettings()

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	// Simulate non-git repository
	runGitCommand = func(cmd string, ctx CwdResolver, ttl int) (string, error) {
		return "false", nil
	}

	ctx := types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}}
	item := types.WidgetItem{Type: "git-branch"}

	_, body, err := w.Render(item, ctx, settings)
	if err != nil || body != "⎇ no git" {
		t.Errorf("Expected '⎇ no git' in non-git dir, got body=%q, err=%v", body, err)
	}

	// Test with Hide=true when no git
	hideVal := true
	itemHide := types.WidgetItem{Type: "git-branch", Hide: &hideVal}
	titleHide, bodyHide, err := w.Render(itemHide, ctx, settings)
	if err != nil || titleHide != "" || bodyHide != "" {
		t.Errorf("Expected empty result when Hide=true and no git, got title=%q, body=%q", titleHide, bodyHide)
	}
}

func TestGitChangesWidget_NonGit(t *testing.T) {
	RegisterAll()
	w := GetWidget("git-changes")
	settings := types.DefaultSettings()

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	// Simulate non-git repository
	runGitCommand = func(cmd string, ctx CwdResolver, ttl int) (string, error) {
		return "false", fmt.Errorf("not a git repository")
	}

	ctx := types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}}
	item := types.WidgetItem{Type: "git-changes"}

	_, body, err := w.Render(item, ctx, settings)
	if err != nil || body != "(no git)" {
		t.Errorf("Expected '(no git)' in non-git dir for git-changes, got body=%q, err=%v", body, err)
	}
}
