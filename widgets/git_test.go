package widgets

import (
	"fmt"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

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
