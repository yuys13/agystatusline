package widgets

import (
	"fmt"
	"math"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yuys13/agystatusline/types"
)

func TestRunGitExec(t *testing.T) {
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git binary not found in PATH, skipping runGitExec test")
	}

	tempDir := t.TempDir()

	// Test initializing a git repo using runGitExec
	out, err := runGitExec([]string{"init"}, tempDir)
	if err != nil {
		t.Fatalf("runGitExec init failed: %v", err)
	}
	if !strings.Contains(out, "Initialized") && !strings.Contains(out, "Reinitialized") {
		t.Logf("runGitExec init output: %q", out)
	}

	// Test checking if inside work tree
	out, err = runGitExec([]string{"rev-parse", "--is-inside-work-tree"}, tempDir)
	if err != nil {
		t.Fatalf("runGitExec rev-parse failed: %v", err)
	}
	if strings.TrimSpace(out) != "true" {
		t.Errorf("Expected 'true', got %q", out)
	}

	// Test invalid command returning error
	_, err = runGitExec([]string{"non-existent-git-subcommand-12345"}, tempDir)
	if err == nil {
		t.Errorf("Expected error for invalid git subcommand, got nil")
	}
}

func TestGetWidget(t *testing.T) {
	RegisterAll()

	tests := []struct {
		name      string
		wName     string
		expectNil bool
	}{
		{name: "Valid widget model", wName: "model", expectNil: false},
		{name: "Valid widget git-branch", wName: "git-branch", expectNil: false},
		{name: "Valid widget git-changes", wName: "git-changes", expectNil: false},
		{name: "Valid widget custom-text", wName: "custom-text", expectNil: false},
		{name: "Valid widget sandbox", wName: "sandbox", expectNil: false},
		{name: "Valid widget agent-state", wName: "agent-state", expectNil: false},
		{name: "Valid widget context-bar", wName: "context-bar", expectNil: false},
		{name: "Valid widget quota", wName: "quota", expectNil: false},
		{name: "Valid widget quota-bar", wName: "quota-bar", expectNil: false},
		{name: "Valid widget artifacts", wName: "artifacts", expectNil: false},
		{name: "Valid widget subagents", wName: "subagents", expectNil: false},
		{name: "Valid widget tasks", wName: "tasks", expectNil: false},
		{name: "Non-existent widget", wName: "does-not-exist-widget", expectNil: true},
		{name: "Empty widget name", wName: "", expectNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := GetWidget(tt.wName)
			if tt.expectNil && w != nil {
				t.Errorf("Expected nil for widget %q, got %v", tt.wName, w)
			}
			if !tt.expectNil && w == nil {
				t.Errorf("Expected non-nil widget for %q, got nil", tt.wName)
			}
		})
	}
}

func TestWidgets_Raw(t *testing.T) {
	RegisterAll()
	settings := types.DefaultSettings()

	widgetsList := []string{
		"model", "git-branch", "sandbox", "context-bar",
		"artifacts", "subagents", "tasks", "quota", "quota-bar",
	}

	valTrue := true
	countVal := 5
	remFrac := 0.8
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{DisplayName: "Claude 3.5 Sonnet"},
			VCS:   &types.VCSInfo{Branch: "feature-branch", Dirty: &valTrue},
			Sandbox: &types.SandboxInfo{
				Enabled: &valTrue,
			},
			ArtifactCount: &countVal,
			Subagents:     float64(4),
			TaskCount:     &countVal,
			ContextWindow: &types.ContextWindowInfo{
				UsedPercentage: func() *float64 { v := 40.0; return &v }(),
			},
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {RemainingFraction: &remFrac},
			},
		},
	}

	for _, wName := range widgetsList {
		t.Run(wName, func(t *testing.T) {
			w := GetWidget(wName)
			if w == nil {
				t.Fatalf("Widget %s not found", wName)
			}
			item := types.WidgetItem{Type: wName, Raw: true, Key: "gemini-5h"}
			title, body, err := w.Render(item, ctx, settings)
			if err != nil {
				t.Errorf("Widget %s Render failed with Raw: %v", wName, err)
			}
			if title != "" {
				t.Errorf("Widget %s with Raw=true returned non-empty title %q", wName, title)
			}
			if strings.Contains(body, "\x1b[") {
				t.Errorf("Widget %s with Raw=true returned ANSI codes in body %q", wName, body)
			}
			if body == "" {
				t.Errorf("Widget %s with Raw=true returned empty body", wName)
			}
		})
	}
}

func TestWidgets_ConcurrentRenderStress(t *testing.T) {
	RegisterAll()
	settings := types.DefaultSettings()

	widgetsList := []string{
		"model", "git-branch", "git-changes", "quota", "custom-text",
		"sandbox", "agent-state", "context-bar", "quota-bar", "artifacts",
		"subagents", "tasks",
	}

	var wg sync.WaitGroup
	startSignal := make(chan struct{})

	for i := range 20 {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			<-startSignal

			count := workerID
			pct := float64(workerID % 100)
			remFrac := float64(workerID%100) / 100.0
			ctx := types.RenderContext{
				Data: types.StatusJSON{
					Model:         types.ModelInfo{ID: fmt.Sprintf("model-%d", workerID)},
					ArtifactCount: &count,
					ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pct},
					Quota: map[string]types.QuotaInfo{
						"gemini-5h": {RemainingFraction: &remFrac},
					},
				},
			}

			for _, name := range widgetsList {
				w := GetWidget(name)
				item := types.WidgetItem{Type: name, Key: "gemini-5h", Text: "custom"}
				_, _, err := w.Render(item, ctx, settings)
				if err != nil {
					t.Errorf("Worker %d: widget %s Render returned error: %v", workerID, name, err)
				}
				color := w.GetBodyColor(item, ctx)
				if color == "" {
					t.Errorf("Worker %d: widget %s GetBodyColor returned empty string", workerID, name)
				}
			}
		}(i)
	}

	close(startSignal)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Concurrent stress test timed out (deadlock or slow execution)")
	}
}

func TestAdversarialStressAllWidgets_ExtremeEdgeCases(t *testing.T) {
	RegisterAll()
	settings := types.DefaultSettings()

	widgetsList := []string{
		"model", "context-bar", "agent-state", "artifacts", "subagents",
		"tasks", "sandbox", "quota", "quota-bar", "git-branch", "git-changes", "custom-text",
	}

	// 1. Extreme numeric and pointer fixtures
	hugeInt := 1<<31 - 1
	minInt := -1 << 31
	negInt := -9999
	hugeFloat := 1e18
	negFloat := -1e5
	nanFloat := math.NaN()
	infPos := math.Inf(1)
	infNeg := math.Inf(-1)
	boolTrue := true
	boolFalse := false

	// Unicode and boundary strings
	unicodeStrings := []string{
		"日本語ブランチ_テスト",
		"🚀✨🎉🔥🤖",
		"العربية",
		"مرحبا",
		"Mixed_English_日本語_🌟",
		strings.Repeat("A", 10000),
		"\x1b[31;1mANSI_INJECTION\x1b[0m",
		"\x00\x01\x02\r\n\t",
	}

	// 2. Telemetry contexts covering extreme variations
	testContexts := []struct {
		name string
		ctx  types.RenderContext
	}{
		{
			name: "Completely empty zero-value StatusJSON",
			ctx:  types.RenderContext{Data: types.StatusJSON{}},
		},
		{
			name: "All nil pointers with empty nested structs",
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Model:         types.ModelInfo{ID: "", DisplayName: ""},
					VCS:           &types.VCSInfo{Branch: "", Dirty: nil},
					Sandbox:       &types.SandboxInfo{Enabled: nil},
					ContextWindow: &types.ContextWindowInfo{UsedPercentage: nil},
					ArtifactCount: nil,
					TaskCount:     nil,
					Subagents:     nil,
					Quota:         map[string]types.QuotaInfo{},
				},
			},
		},
		{
			name: "Extreme positive overflow numbers",
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Model:         types.ModelInfo{DisplayName: strings.Repeat("M", 5000)},
					VCS:           &types.VCSInfo{Branch: strings.Repeat("B", 5000), Dirty: &boolTrue},
					Sandbox:       &types.SandboxInfo{Enabled: &boolTrue},
					ContextWindow: &types.ContextWindowInfo{UsedPercentage: &hugeFloat},
					ArtifactCount: &hugeInt,
					TaskCount:     &hugeInt,
					Subagents:     float64(hugeInt),
					Quota: map[string]types.QuotaInfo{
						"gemini-5h": {
							RemainingFraction: &hugeFloat,
							ResetInSeconds:    &hugeFloat,
						},
						"overflow-key": {
							RemainingFraction: &hugeFloat,
							ResetInSeconds:    &hugeFloat,
						},
					},
				},
			},
		},
		{
			name: "Extreme negative numbers",
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Model:         types.ModelInfo{DisplayName: "negative-test"},
					VCS:           &types.VCSInfo{Branch: "neg-branch", Dirty: &boolFalse},
					Sandbox:       &types.SandboxInfo{Enabled: &boolFalse},
					ContextWindow: &types.ContextWindowInfo{UsedPercentage: &negFloat},
					ArtifactCount: &negInt,
					TaskCount:     &minInt,
					Subagents:     negInt,
					Quota: map[string]types.QuotaInfo{
						"gemini-5h": {
							RemainingFraction: &negFloat,
							ResetInSeconds:    &negFloat,
						},
					},
				},
			},
		},
		{
			name: "Special float values (NaN, Inf)",
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					ContextWindow: &types.ContextWindowInfo{UsedPercentage: &nanFloat},
					Quota: map[string]types.QuotaInfo{
						"nan-quota": {
							RemainingFraction: &nanFloat,
							ResetInSeconds:    &nanFloat,
						},
						"inf-pos": {
							RemainingFraction: &infPos,
							ResetInSeconds:    &infPos,
						},
						"inf-neg": {
							RemainingFraction: &infNeg,
							ResetInSeconds:    &infNeg,
						},
					},
				},
			},
		},
		{
			name: "Subagents with diverse types in interface{}",
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Subagents: []any{1, "two", 3.0, map[string]string{"foo": "bar"}},
				},
			},
		},
		{
			name: "Subagents with unexpected struct type in interface{}",
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Subagents: struct{ Key string }{Key: "unexpected"},
				},
			},
		},
	}

	// 3. Run all widgets against all test contexts
	for _, tc := range testContexts {
		t.Run(tc.name, func(t *testing.T) {
			for _, wName := range widgetsList {
				t.Run(wName, func(t *testing.T) {
					w := GetWidget(wName)
					if w == nil {
						t.Fatalf("Widget %s not registered", wName)
					}

					// Test normal mode
					itemNormal := types.WidgetItem{
						Type:   wName,
						Key:    "gemini-5h",
						Text:   "CustomLabel",
						Symbol: "⎇ ",
					}
					title, body, err := w.Render(itemNormal, tc.ctx, settings)
					if err != nil {
						t.Errorf("Widget %s Render returned error: %v", wName, err)
					}
					_ = title
					_ = body

					color := w.GetBodyColor(itemNormal, tc.ctx)
					if color == "" {
						t.Errorf("Widget %s GetBodyColor returned empty color", wName)
					}

					// Test raw mode
					itemRaw := types.WidgetItem{
						Type: wName,
						Key:  "gemini-5h",
						Raw:  true,
					}
					rawTitle, rawBody, err := w.Render(itemRaw, tc.ctx, settings)
					if err != nil {
						t.Errorf("Widget %s Render with Raw=true returned error: %v", wName, err)
					}
					if rawTitle != "" {
						t.Errorf("Widget %s with Raw=true expected empty title, got %q", wName, rawTitle)
					}
					_ = rawBody
				})
			}
		})
	}

	// 4. Test Unicode and symbol stress
	for _, uStr := range unicodeStrings {
		t.Run("UnicodeString_"+uStr[:min(len(uStr), 10)], func(t *testing.T) {
			customWidget := GetWidget("custom-text")
			title, body, err := customWidget.Render(types.WidgetItem{Type: "custom-text", Text: uStr}, types.RenderContext{}, settings)
			if err != nil {
				t.Fatalf("custom-text error on unicode string: %v", err)
			}
			if title != "" || body != uStr {
				t.Errorf("Expected exact unicode echo in body, got body=%q title=%q", body, title)
			}

			gitBranchWidget := GetWidget("git-branch")
			dirty := true
			ctx := types.RenderContext{
				Data: types.StatusJSON{
					VCS: &types.VCSInfo{Branch: uStr, Dirty: &dirty},
				},
			}
			_, gitBody, err := gitBranchWidget.Render(types.WidgetItem{Type: "git-branch", Symbol: "🌿 "}, ctx, settings)
			if err != nil {
				t.Fatalf("git-branch error on unicode branch: %v", err)
			}
			expectedGit := "🌿 " + uStr + "*"
			if gitBody != expectedGit {
				t.Errorf("Expected gitBody=%q, got %q", expectedGit, gitBody)
			}
		})
	}

	// 5. Test Quota and Quota-Bar with unexpected / missing keys
	unexpectedKeys := []string{
		"",
		"unknown-quota-key",
		"gemini-5h/non-existent",
		"クォータキー",
		"!@#$%^&*()_+",
		"3p-custom-limit",
	}

	quotaW := GetWidget("quota")
	quotaBarW := GetWidget("quota-bar")
	remVal := 0.75
	resetSecs := 3600.0
	quotaCtx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {RemainingFraction: &remVal, ResetInSeconds: &resetSecs},
			},
		},
	}

	for _, key := range unexpectedKeys {
		t.Run("QuotaKey_"+key, func(t *testing.T) {
			item := types.WidgetItem{Type: "quota", Key: key}
			title, body, err := quotaW.Render(item, quotaCtx, settings)
			if err != nil {
				t.Errorf("quota render error on key %q: %v", key, err)
			}
			if key != "gemini-5h" && (title != "" || body != "") {
				t.Errorf("Expected empty quota output for missing key %q, got title=%q body=%q", key, title, body)
			}

			itemBar := types.WidgetItem{Type: "quota-bar", Key: key}
			barTitle, barBody, err := quotaBarW.Render(itemBar, quotaCtx, settings)
			if err != nil {
				t.Errorf("quota-bar render error on key %q: %v", key, err)
			}
			if key != "gemini-5h" && (barTitle != "" || barBody != "") {
				t.Errorf("Expected empty quota-bar output for missing key %q, got title=%q body=%q", key, barTitle, barBody)
			}
		})
	}
}
