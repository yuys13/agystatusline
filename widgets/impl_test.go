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

