package tui

import (
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestDefaultPreviewContext(t *testing.T) {
	settings := types.DefaultSettings()
	settings.General.Minimalist = true
	settings.General.GitCacheTTL = 10

	t.Run("ContextFields", func(t *testing.T) {
		ctx := DefaultPreviewContext(settings)

		if ctx.TerminalWidth == nil || *ctx.TerminalWidth != 160 {
			t.Errorf("expected terminal width 160, got %v", ctx.TerminalWidth)
		}
		if ctx.Minimalist != true {
			t.Errorf("expected minimalist true, got %v", ctx.Minimalist)
		}
		if ctx.GitCacheTTLSeconds != 10 {
			t.Errorf("expected git cache TTL 10, got %v", ctx.GitCacheTTLSeconds)
		}
		if ctx.Data.Model.ID != "gemini-3.5-flash-medium" {
			t.Errorf("expected model id gemini-3.5-flash-medium, got %q", ctx.Data.Model.ID)
		}
		if ctx.GitProvider == nil {
			t.Fatalf("expected non-nil GitProvider in preview context")
		}
	})

	t.Run("MockGitProvider", func(t *testing.T) {
		ctx := DefaultPreviewContext(settings)
		if ctx.GitProvider == nil {
			t.Fatalf("expected non-nil GitProvider in preview context")
		}

		branch, dirty, err := ctx.GitProvider.GetBranch()
		if err != nil || branch != "main" || dirty != false {
			t.Errorf("expected GetBranch() = (main, false, nil), got (%q, %v, %v)", branch, dirty, err)
		}

		changes, err := ctx.GitProvider.GetChanges()
		if err != nil || changes.Insertions != 42 || changes.Deletions != 10 {
			t.Errorf("expected GetChanges() = (+42, -10), got (%+v, %v)", changes, err)
		}

		ab, err := ctx.GitProvider.GetAheadBehind()
		if err != nil || ab.Ahead != 2 || ab.Behind != 1 {
			t.Errorf("expected GetAheadBehind() = (2, 1), got (%+v, %v)", ab, err)
		}

		repo, err := ctx.GitProvider.GetRepoName()
		if err != nil || repo != "agystatusline" {
			t.Errorf("expected GetRepoName() = agystatusline, got (%q, %v)", repo, err)
		}
	})
}
