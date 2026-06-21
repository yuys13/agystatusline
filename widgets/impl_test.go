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

func TestContextLengthWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("context-length")
	if w == nil {
		t.Fatalf("Context length widget not found")
	}

	settings := types.DefaultSettings()
	inputTokens := float64(12500)
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			ContextWindow: &types.ContextWindowInfo{
				TotalInputTokens: &inputTokens,
			},
		},
	}
	item := types.WidgetItem{Type: "context-length"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "12.5k" {
		t.Errorf("Expected title '' and body '12.5k', got title '%s' and body '%s'", title, output)
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
	if title != "" || output != "⎇feature/tdd" {
		t.Errorf("Expected title '' and body '⎇feature/tdd', got title '%s' and body '%s'", title, output)
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

func TestContextUsedPctWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("context-used-pct")
	if w == nil {
		t.Fatalf("Context used percentage widget not found")
	}

	settings := types.DefaultSettings()
	usedPct := 0.014019012451171875
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			ContextWindow: &types.ContextWindowInfo{
				UsedPercentage: &usedPct,
			},
		},
	}
	item := types.WidgetItem{Type: "context-used-pct"}

	// Normal render
	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "Used" || output != "0.01%" {
		t.Errorf("Expected title 'Used' and body '0.01%%', got title '%s' and body '%s'", title, output)
	}

	// RawValue render
	rawVal := true
	itemRaw := types.WidgetItem{Type: "context-used-pct", RawValue: &rawVal}
	titleRaw, outputRaw, err := w.Render(itemRaw, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleRaw != "" || outputRaw != "0.01%" {
		t.Errorf("Expected title '' and body '0.01%%', got title '%s' and body '%s'", titleRaw, outputRaw)
	}
}

func TestContextRemainingPctWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("context-remaining-pct")
	if w == nil {
		t.Fatalf("Context remaining percentage widget not found")
	}

	settings := types.DefaultSettings()
	remainingPct := 99.98598098754883
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			ContextWindow: &types.ContextWindowInfo{
				RemainingPercentage: &remainingPct,
			},
		},
	}
	item := types.WidgetItem{Type: "context-remaining-pct"}

	// Normal render
	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "Remaining" || output != "99.99%" {
		t.Errorf("Expected title 'Remaining' and body '99.99%%', got title '%s' and body '%s'", title, output)
	}

	// RawValue render
	rawVal := true
	itemRaw := types.WidgetItem{Type: "context-remaining-pct", RawValue: &rawVal}
	titleRaw, outputRaw, err := w.Render(itemRaw, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleRaw != "" || outputRaw != "99.99%" {
		t.Errorf("Expected title '' and body '99.99%%', got title '%s' and body '%s'", titleRaw, outputRaw)
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
		// Verify gemini-weekly maps to 'weekly'
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
		if titleW != "weekly" {
			t.Errorf("Expected title 'weekly', got %q", titleW)
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
}
