package widgets

import (
	"fmt"
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

func TestWidgets_PreserveColors(t *testing.T) {
	RegisterAll()
	settings := types.DefaultSettings()

	widgetsList := []string{
		"model", "git-branch", "sandbox", "agent-state",
		"context-bar", "artifacts", "subagents", "tasks",
	}

	valTrue := true
	countVal := 3
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{ID: "test-model"},
			VCS:   &types.VCSInfo{Branch: "main", Dirty: &valTrue},
			Sandbox: &types.SandboxInfo{
				Enabled: &valTrue,
			},
			AgentState:    "working",
			ArtifactCount: &countVal,
			Subagents:     float64(2),
			TaskCount:     &countVal,
			ContextWindow: &types.ContextWindowInfo{
				UsedPercentage: func() *float64 { v := 75.0; return &v }(),
			},
		},
	}

	for _, wName := range widgetsList {
		w := GetWidget(wName)
		if w == nil {
			t.Fatalf("Widget %s not found", wName)
		}
		item := types.WidgetItem{Type: wName, PreserveColors: &valTrue}
		title, body, err := w.Render(item, ctx, settings)
		if err != nil {
			t.Errorf("Widget %s Render failed with PreserveColors: %v", wName, err)
		}
		if !strings.Contains(body, "\x1b[") && !strings.Contains(title, "\x1b[") {
			t.Errorf("Widget %s output does not contain ANSI escape sequences under PreserveColors: title=%q, body=%q", wName, title, body)
		}
	}
}

func TestWidgets_RawValue(t *testing.T) {
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
		w := GetWidget(wName)
		if w == nil {
			t.Fatalf("Widget %s not found", wName)
		}
		item := types.WidgetItem{Type: wName, RawValue: &valTrue, Metadata: map[string]string{"key": "gemini-5h"}}
		title, body, err := w.Render(item, ctx, settings)
		if err != nil {
			t.Errorf("Widget %s Render failed with RawValue: %v", wName, err)
		}
		if title != "" {
			t.Errorf("Widget %s with RawValue returned non-empty title %q", wName, title)
		}
		if strings.Contains(body, "\x1b[") {
			t.Errorf("Widget %s with RawValue returned ANSI codes in body %q", wName, body)
		}
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

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			<-startSignal

			count := workerID
			pct := float64(workerID % 100)
			ctx := types.RenderContext{
				Data: types.StatusJSON{
					Model:         types.ModelInfo{ID: fmt.Sprintf("model-%d", workerID)},
					ArtifactCount: &count,
					ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pct},
				},
			}

			for _, name := range widgetsList {
				w := GetWidget(name)
				item := types.WidgetItem{Type: name, Metadata: map[string]string{"key": "gemini-5h"}}
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
