package widgets

import (
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestModelWidget(t *testing.T) {
	RegisterAll()

	w := GetWidget("model")
	if w == nil {
		t.Fatalf("Model widget not found in registry")
	}

	if w.GetDefaultColor() != "cyan" {
		t.Errorf("Expected default color 'cyan', got '%s'", w.GetDefaultColor())
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
	output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	expected := "Model: Claude 3.5 Sonnet"
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, output)
	}

	// RawValue render
	rawVal := true
	itemRaw := types.WidgetItem{
		Type:     "model",
		RawValue: &rawVal,
	}
	outputRaw, err := w.Render(itemRaw, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	expectedRaw := "Claude 3.5 Sonnet"
	if outputRaw != expectedRaw {
		t.Errorf("Expected '%s', got '%s'", expectedRaw, outputRaw)
	}

	// Test parenthesized suffixes: (New) should be stripped, but (Medium) should be kept.
	ctxWithNew := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID:          "claude-3-5-sonnet-new",
				DisplayName: "Claude 3.5 Sonnet (New)",
			},
		},
	}
	outputNew, err := w.Render(item, ctxWithNew, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	expectedNew := "Model: Claude 3.5 Sonnet"
	if outputNew != expectedNew {
		t.Errorf("Expected '%s', got '%s'", expectedNew, outputNew)
	}

	ctxWithMedium := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID:          "gemini-3.5-flash-medium",
				DisplayName: "Gemini 3.5 Flash (Medium)",
			},
		},
	}
	outputMedium, err := w.Render(item, ctxWithMedium, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	expectedMedium := "Model: Gemini 3.5 Flash (Medium)"
	if outputMedium != expectedMedium {
		t.Errorf("Expected '%s', got '%s'", expectedMedium, outputMedium)
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

	output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if output != "12.5k" {
		t.Errorf("Expected '12.5k', got '%s'", output)
	}
}

func TestGitBranchWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("git-branch")
	if w == nil {
		t.Fatalf("Git branch widget not found")
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

	output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if output != "⎇feature/tdd" {
		t.Errorf("Expected '⎇feature/tdd', got '%s'", output)
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

	output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if output != "(+13,-5)" {
		t.Errorf("Expected '(+13,-5)', got '%s'", output)
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
	output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if output != "Used: 0.01%" {
		t.Errorf("Expected 'Used: 0.01%%', got '%s'", output)
	}

	// RawValue render
	rawVal := true
	itemRaw := types.WidgetItem{Type: "context-used-pct", RawValue: &rawVal}
	outputRaw, err := w.Render(itemRaw, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if outputRaw != "0.01%" {
		t.Errorf("Expected '0.01%%', got '%s'", outputRaw)
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
	output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if output != "Remaining: 99.99%" {
		t.Errorf("Expected 'Remaining: 99.99%%', got '%s'", output)
	}

	// RawValue render
	rawVal := true
	itemRaw := types.WidgetItem{Type: "context-remaining-pct", RawValue: &rawVal}
	outputRaw, err := w.Render(itemRaw, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if outputRaw != "99.99%" {
		t.Errorf("Expected '99.99%%', got '%s'", outputRaw)
	}
}

func TestQuotaWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota")
	if w == nil {
		t.Fatalf("Quota widget not found")
	}

	if w.GetDefaultColor() != "brightBlack" {
		t.Errorf("Expected default color 'brightBlack', got '%s'", w.GetDefaultColor())
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
	output1, err := w.Render(item1, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if output1 != "gemini-5h: 50.19% (2h 28m)" {
		t.Errorf("Expected 'gemini-5h: 50.19%% (2h 28m)', got '%s'", output1)
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
	output2, err := w.Render(item2, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if output2 != "50.19% (2h 28m)" {
		t.Errorf("Expected '50.19%% (2h 28m)', got '%s'", output2)
	}

	// Case 3: Custom Text label + Reset
	item3 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
		CustomText: "Gemini Q",
	}
	output3, err := w.Render(item3, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if output3 != "Gemini Q: 50.19% (2h 28m)" {
		t.Errorf("Expected 'Gemini Q: 50.19%% (2h 28m)', got '%s'", output3)
	}

	// Case 3b: display="quota" (Percentage only, labeled)
	itemQuotaOnly := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "quota",
		},
	}
	outputQuotaOnly, err := w.Render(itemQuotaOnly, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if outputQuotaOnly != "gemini-5h: 50.19%" {
		t.Errorf("Expected 'gemini-5h: 50.19%%', got '%s'", outputQuotaOnly)
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
	outputQuotaOnlyRaw, err := w.Render(itemQuotaOnlyRaw, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if outputQuotaOnlyRaw != "50.19%" {
		t.Errorf("Expected '50.19%%', got '%s'", outputQuotaOnlyRaw)
	}

	// Case 4: Reset time labeled
	item4 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "reset",
		},
	}
	output4, err := w.Render(item4, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if output4 != "gemini-5h (reset): 2h 28m" {
		t.Errorf("Expected 'gemini-5h (reset): 2h 28m', got '%s'", output4)
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
	output5, err := w.Render(item5, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if output5 != "2h 28m" {
		t.Errorf("Expected '2h 28m', got '%s'", output5)
	}

	// Case 6: Reset time other durations
	// 45 seconds -> 45s
	// 125 seconds -> 2m 5s
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
	output45, _ := w.Render(itemSecs45, ctxDur, settings)
	if output45 != "45s" {
		t.Errorf("Expected '45s', got '%s'", output45)
	}

	itemSecs125 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "secs125",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	output125, _ := w.Render(itemSecs125, ctxDur, settings)
	if output125 != "2m 5s" {
		t.Errorf("Expected '2m 5s', got '%s'", output125)
	}

	itemSecs567440 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "secs567440",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	output567440, _ := w.Render(itemSecs567440, ctxDur, settings)
	if output567440 != "6d 13h" {
		t.Errorf("Expected '6d 13h', got '%s'", output567440)
	}

	// Case 7: Key not found or empty
	itemEmpty := types.WidgetItem{
		Type: "quota",
	}
	outputEmpty, _ := w.Render(itemEmpty, ctx, settings)
	if outputEmpty != "" {
		t.Errorf("Expected empty string for empty key, got '%s'", outputEmpty)
	}

	itemInvalid := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "invalid-key",
		},
	}
	outputInvalid, _ := w.Render(itemInvalid, ctx, settings)
	if outputInvalid != "" {
		t.Errorf("Expected empty string for invalid key, got '%s'", outputInvalid)
	}

	// Case 8: Quota map is nil
	ctxNilQuota := types.RenderContext{
		Data: types.StatusJSON{},
	}
	outputNil, _ := w.Render(item1, ctxNilQuota, settings)
	if outputNil != "" {
		t.Errorf("Expected empty string for nil quota map, got '%s'", outputNil)
	}
}
