package widgets

import (
	"fmt"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestGitBranchWidget_Normal(t *testing.T) {
	w := initTestWidget(t, "git-branch")
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

func TestGitChangesWidget_Normal(t *testing.T) {
	w := initTestWidget(t, "git-changes")

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

func TestGitBranchWidget_EdgeCase_CommandError_HideTrue(t *testing.T) {
	wBranch := initTestWidget(t, "git-branch")
	settings := types.DefaultSettings()

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()
	runGitCommand = func(cmd string, ctx CwdResolver, ttl int) (string, error) {
		return "", fmt.Errorf("git not installed")
	}

	ctx := types.RenderContext{Data: types.StatusJSON{CWD: "/no-git-dir"}}
	hideTrue := true
	itemHide := types.WidgetItem{Type: "git-branch", Hide: &hideTrue}

	_, bodyBranchHide, err := wBranch.Render(itemHide, ctx, settings)
	if err != nil || bodyBranchHide != "" {
		t.Errorf("Expected empty string when git missing and hide=true, got %q, err=%v", bodyBranchHide, err)
	}
}

func TestGitBranchWidget_EdgeCase_CommandError_HideFalse(t *testing.T) {
	wBranch := initTestWidget(t, "git-branch")
	settings := types.DefaultSettings()

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()
	runGitCommand = func(cmd string, ctx CwdResolver, ttl int) (string, error) {
		return "", fmt.Errorf("git not installed")
	}

	ctx := types.RenderContext{Data: types.StatusJSON{CWD: "/no-git-dir"}}
	hideFalse := false
	itemNoHide := types.WidgetItem{Type: "git-branch", Hide: &hideFalse}

	_, bodyBranchNoHide, err := wBranch.Render(itemNoHide, ctx, settings)
	if err != nil || bodyBranchNoHide != "⎇ no git" {
		t.Errorf("Expected '⎇ no git', got %q, err=%v", bodyBranchNoHide, err)
	}
}

func TestGitBranchWidget_GetBodyColor_Dirty(t *testing.T) {
	wBranch := initTestWidget(t, "git-branch")
	hideTrue := true
	itemNoHide := types.WidgetItem{Type: "git-branch"}
	ctxDirty := types.RenderContext{
		Data: types.StatusJSON{
			VCS: &types.VCSInfo{Dirty: &hideTrue},
		},
	}

	if wBranch.GetBodyColor(itemNoHide, ctxDirty) != "brightRed" {
		t.Errorf("Expected brightRed for dirty git branch")
	}
}

func TestGitChangesWidget_EdgeCase_CommandError_HideTrue(t *testing.T) {
	wChanges := initTestWidget(t, "git-changes")
	settings := types.DefaultSettings()

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()
	runGitCommand = func(cmd string, ctx CwdResolver, ttl int) (string, error) {
		return "", fmt.Errorf("git not installed")
	}

	ctx := types.RenderContext{Data: types.StatusJSON{CWD: "/no-git-dir"}}
	hideTrue := true
	itemChangesHide := types.WidgetItem{Type: "git-changes", Hide: &hideTrue}

	_, bodyChangesHide, err := wChanges.Render(itemChangesHide, ctx, settings)
	if err != nil || bodyChangesHide != "" {
		t.Errorf("Expected empty string for git-changes when git missing and hide=true, got %q, err=%v", bodyChangesHide, err)
	}
}

func TestGitBranchWidget_NonGit_HideFalse(t *testing.T) {
	w := initTestWidget(t, "git-branch")
	settings := types.DefaultSettings()

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	runGitCommand = func(cmd string, ctx CwdResolver, ttl int) (string, error) {
		return "false", nil
	}

	ctx := types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}}
	item := types.WidgetItem{Type: "git-branch"}

	_, body, err := w.Render(item, ctx, settings)
	if err != nil || body != "⎇ no git" {
		t.Errorf("Expected '⎇ no git' in non-git dir, got body=%q, err=%v", body, err)
	}
}

func TestGitBranchWidget_NonGit_HideTrue(t *testing.T) {
	w := initTestWidget(t, "git-branch")
	settings := types.DefaultSettings()

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	runGitCommand = func(cmd string, ctx CwdResolver, ttl int) (string, error) {
		return "false", nil
	}

	ctx := types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}}
	hideVal := true
	itemHide := types.WidgetItem{Type: "git-branch", Hide: &hideVal}

	titleHide, bodyHide, err := w.Render(itemHide, ctx, settings)
	if err != nil || titleHide != "" || bodyHide != "" {
		t.Errorf("Expected empty result when Hide=true and no git, got title=%q, body=%q, err=%v", titleHide, bodyHide, err)
	}
}

func TestGitChangesWidget_NonGit(t *testing.T) {
	w := initTestWidget(t, "git-changes")
	settings := types.DefaultSettings()

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

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

func TestParseShortStat_Empty(t *testing.T) {
	ins, del := parseShortStat("")
	if ins != 0 || del != 0 {
		t.Errorf("Expected 0, 0 for empty stat, got %d, %d", ins, del)
	}
}

func TestParseShortStat_InsertionsOnly(t *testing.T) {
	insOnly, delOnly := parseShortStat("5 insertions(+)")
	if insOnly != 5 || delOnly != 0 {
		t.Errorf("Expected 5, 0 for insertion only, got %d, %d", insOnly, delOnly)
	}
}

func TestParseShortStat_DeletionsOnly(t *testing.T) {
	insDel, delDel := parseShortStat("10 deletions(-)")
	if insDel != 0 || delDel != 10 {
		t.Errorf("Expected 0, 10 for deletion only, got %d, %d", insDel, delDel)
	}
}

func TestGitBranchWidget_Interface(t *testing.T) {
	w := initTestWidget(t, "git-branch")
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "git-branch"}

	if nameStr := w.GetDisplayName(); nameStr == "" {
		t.Errorf("GetDisplayName() returned empty for git-branch")
	}
	if defaultColor := w.GetDefaultColor(); defaultColor == "" {
		t.Errorf("GetDefaultColor() returned empty for git-branch")
	}
	if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
		t.Errorf("GetBodyColor() returned empty for git-branch")
	}
}

func TestGitChangesWidget_Interface(t *testing.T) {
	w := initTestWidget(t, "git-changes")
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "git-changes"}

	if nameStr := w.GetDisplayName(); nameStr == "" {
		t.Errorf("GetDisplayName() returned empty for git-changes")
	}
	if defaultColor := w.GetDefaultColor(); defaultColor == "" {
		t.Errorf("GetDefaultColor() returned empty for git-changes")
	}
	if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
		t.Errorf("GetBodyColor() returned empty for git-changes")
	}
}
