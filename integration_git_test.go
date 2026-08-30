package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/renderer"
	"github.com/yuys13/agystatusline/types"
	"github.com/yuys13/agystatusline/utils"
	"github.com/yuys13/agystatusline/widgets"
)

func setupGitRepo(t *testing.T, initialBranch string) string {
	t.Helper()
	repoDir := t.TempDir()

	cmd := exec.Command("git", "init", "-b", initialBranch)
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		cmdInit := exec.Command("git", "init")
		cmdInit.Dir = repoDir
		if err := cmdInit.Run(); err != nil {
			t.Fatalf("Failed to git init: %v", err)
		}
		cmdCheckout := exec.Command("git", "checkout", "-b", initialBranch)
		cmdCheckout.Dir = repoDir
		_ = cmdCheckout.Run()
	}

	_ = exec.Command("git", "-C", repoDir, "config", "user.name", "Test User").Run()
	_ = exec.Command("git", "-C", repoDir, "config", "user.email", "test@example.com").Run()

	readmePath := filepath.Join(repoDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repo\nInitial line 1\nInitial line 2\n"), 0644); err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}
	_ = exec.Command("git", "-C", repoDir, "add", ".").Run()
	_ = exec.Command("git", "-C", repoDir, "commit", "-m", "Initial commit").Run()

	return repoDir
}

func setupTestCacheDir(t *testing.T) string {
	cacheDir := t.TempDir()
	utils.SetCacheDir(cacheDir)
	utils.ClearGitCache()
	t.Cleanup(func() {
		utils.ClearGitCache()
		utils.SetCacheDir("")
	})
	return cacheDir
}

// TestIntegration_Git_CleanRepo tests Git branch & change widgets in a clean real Git repo.
func TestIntegration_Git_CleanRepo(t *testing.T) {
	cacheDir := setupTestCacheDir(t)
	widgets.RegisterAll()

	repoDir := setupGitRepo(t, "main")
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			CWD: repoDir,
		},
		GitCacheTTLSeconds: 5,
	}
	settings := types.DefaultSettings()

	branchWidget := widgets.GetWidget("git-branch")
	if branchWidget == nil {
		t.Fatalf("git-branch widget not found")
	}
	branchItem := types.WidgetItem{Type: "git-branch"}
	title, body, err := branchWidget.Render(branchItem, ctx, settings)
	if err != nil {
		t.Fatalf("Unexpected error rendering branch widget: %v", err)
	}
	plainBranch := renderer.StripAnsi(title + body)

	if !strings.Contains(plainBranch, "main") {
		t.Errorf("Expected branch widget to contain 'main', got %q", plainBranch)
	}
	if strings.Contains(plainBranch, "*") {
		t.Errorf("Expected clean repo branch to not have dirty asterisk, got %q", plainBranch)
	}

	changesWidget := widgets.GetWidget("git-changes")
	if changesWidget == nil {
		t.Fatalf("git-changes widget not found")
	}
	changesItem := types.WidgetItem{Type: "git-changes"}
	titleC, bodyC, errC := changesWidget.Render(changesItem, ctx, settings)
	if errC != nil {
		t.Fatalf("Unexpected error rendering changes widget: %v", errC)
	}
	plainChanges := renderer.StripAnsi(titleC + bodyC)

	if plainChanges != "(+0,-0)" {
		t.Errorf("Expected changes widget to be '(+0,-0)' for clean repo, got %q", plainChanges)
	}

	gitCachePath := filepath.Join(cacheDir, "git-cache")
	files, err := os.ReadDir(gitCachePath)
	if err != nil || len(files) == 0 {
		t.Errorf("Expected persistent git cache files created under %s, got err=%v, count=%d", gitCachePath, err, len(files))
	}
}

// TestIntegration_Git_ModifiedFiles tests git changes calculation (+/-) with staged & unstaged files.
func TestIntegration_Git_ModifiedFiles(t *testing.T) {
	setupTestCacheDir(t)
	widgets.RegisterAll()

	repoDir := setupGitRepo(t, "main")

	// Unstaged edit: append 3 lines, remove 1 line
	readmePath := filepath.Join(repoDir, "README.md")
	_ = os.WriteFile(readmePath, []byte("# Test Repo\nInitial line 1\nAdded line 1\nAdded line 2\nAdded line 3\n"), 0644)

	// Staged file: 2 new lines
	stagedPath := filepath.Join(repoDir, "staged.txt")
	_ = os.WriteFile(stagedPath, []byte("staged line 1\nstaged line 2\n"), 0644)
	_ = exec.Command("git", "-C", repoDir, "add", "staged.txt").Run()

	utils.ClearGitCache()

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			CWD: repoDir,
		},
		GitCacheTTLSeconds: 5,
	}
	settings := types.DefaultSettings()

	branchWidget := widgets.GetWidget("git-branch")
	title, body, _ := branchWidget.Render(types.WidgetItem{Type: "git-branch"}, ctx, settings)
	plainBranch := renderer.StripAnsi(title + body)

	if !strings.Contains(plainBranch, "*") {
		t.Errorf("Expected dirty branch asterisk '*', got %q", plainBranch)
	}

	changesWidget := widgets.GetWidget("git-changes")
	titleC, bodyC, _ := changesWidget.Render(types.WidgetItem{Type: "git-changes"}, ctx, settings)
	plainChanges := renderer.StripAnsi(titleC + bodyC)

	if plainChanges != "(+5,-1)" {
		t.Errorf("Expected changes (+5,-1), got %q", plainChanges)
	}
}

// TestIntegration_Git_WidgetFormattingOptions tests Symbol, Raw, and custom rendering.
func TestIntegration_Git_WidgetFormattingOptions(t *testing.T) {
	setupTestCacheDir(t)
	widgets.RegisterAll()

	repoDir := setupGitRepo(t, "feature/integration")
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			CWD: repoDir,
		},
		GitCacheTTLSeconds: 5,
	}
	settings := types.DefaultSettings()

	branchWidget := widgets.GetWidget("git-branch")

	// Custom Symbol
	itemCustom := types.WidgetItem{
		Type:   "git-branch",
		Symbol: "🌿 ",
	}
	tC, bC, _ := branchWidget.Render(itemCustom, ctx, settings)
	plainCustom := renderer.StripAnsi(tC + bC)
	if !strings.HasPrefix(plainCustom, "🌿 feature/integration") {
		t.Errorf("Expected custom symbol prefix '🌿 feature/integration', got %q", plainCustom)
	}

	// Raw
	itemRaw := types.WidgetItem{
		Type: "git-branch",
		Raw:  true,
	}
	_, bRaw, _ := branchWidget.Render(itemRaw, ctx, settings)
	if bRaw != "feature/integration" {
		t.Errorf("Expected raw value 'feature/integration', got %q", bRaw)
	}
}

// TestIntegration_Git_CacheLifecycleAndMtime tests cache hits and mtime invalidations.
func TestIntegration_Git_CacheLifecycleAndMtime(t *testing.T) {
	setupTestCacheDir(t)
	widgets.RegisterAll()

	repoDir := setupGitRepo(t, "main")
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			CWD: repoDir,
		},
		GitCacheTTLSeconds: 5,
	}
	settings := types.DefaultSettings()

	branchWidget := widgets.GetWidget("git-branch")

	// Initial render populates cache
	t1, b1, _ := branchWidget.Render(types.WidgetItem{Type: "git-branch"}, ctx, settings)

	// Modify file and git add -> updates .git/index mtime
	newFile := filepath.Join(repoDir, "new.txt")
	_ = os.WriteFile(newFile, []byte("content"), 0644)
	_ = exec.Command("git", "-C", repoDir, "add", "new.txt").Run()

	// Subsequent render should detect mtime update and show dirty state
	t2, b2, _ := branchWidget.Render(types.WidgetItem{Type: "git-branch"}, ctx, settings)
	plain2 := renderer.StripAnsi(t2 + b2)

	if !strings.Contains(plain2, "*") {
		t.Errorf("Expected cache invalidation on git add showing dirty '*', got %q (res1 was %q)", plain2, t1+b1)
	}
}

// TestIntegration_Git_NonGitDirectory tests fallback rendering for non-Git directories.
func TestIntegration_Git_NonGitDirectory(t *testing.T) {
	setupTestCacheDir(t)
	widgets.RegisterAll()

	nonGitDir := t.TempDir()
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			CWD: nonGitDir,
		},
		GitCacheTTLSeconds: 5,
	}
	settings := types.DefaultSettings()

	branchWidget := widgets.GetWidget("git-branch")
	tBranch, bBranch, _ := branchWidget.Render(types.WidgetItem{Type: "git-branch"}, ctx, settings)
	plainBranch := renderer.StripAnsi(tBranch + bBranch)
	if !strings.Contains(plainBranch, "no git") {
		t.Errorf("Expected branch widget to return 'no git', got %q", plainBranch)
	}

	changesWidget := widgets.GetWidget("git-changes")
	tChanges, bChanges, _ := changesWidget.Render(types.WidgetItem{Type: "git-changes"}, ctx, settings)
	plainChanges := renderer.StripAnsi(tChanges + bChanges)
	if !strings.Contains(plainChanges, "no git") {
		t.Errorf("Expected changes widget to return '(no git)', got %q", plainChanges)
	}
}

// TestIntegration_Git_VCSTelemetryOverride tests StatusJSON.VCS telemetry taking precedence over Git CLI.
func TestIntegration_Git_VCSTelemetryOverride(t *testing.T) {
	setupTestCacheDir(t)
	widgets.RegisterAll()

	repoDir := setupGitRepo(t, "main")
	dirtyTrue := true

	ctxWithTelemetry := types.RenderContext{
		Data: types.StatusJSON{
			CWD: repoDir,
			VCS: &types.VCSInfo{
				Type:   "git",
				Branch: "telemetry-branch",
				Dirty:  &dirtyTrue,
			},
		},
		GitCacheTTLSeconds: 5,
	}
	settings := types.DefaultSettings()

	branchWidget := widgets.GetWidget("git-branch")
	tTel, bTel, _ := branchWidget.Render(types.WidgetItem{Type: "git-branch"}, ctxWithTelemetry, settings)
	plainTelemetry := renderer.StripAnsi(tTel + bTel)

	if !strings.Contains(plainTelemetry, "telemetry-branch*") {
		t.Errorf("Expected VCS telemetry override 'telemetry-branch*', got %q", plainTelemetry)
	}
}

// TestIntegration_Git_WorktreeFile tests git worktrees with .git file instead of directory.
func TestIntegration_Git_WorktreeFile(t *testing.T) {
	setupTestCacheDir(t)
	widgets.RegisterAll()

	baseRepo := setupGitRepo(t, "main")
	wtDir := filepath.Join(t.TempDir(), "wt-test")

	cmd := exec.Command("git", "-C", baseRepo, "worktree", "add", "-b", "wt-branch", wtDir)
	if err := cmd.Run(); err != nil {
		t.Skipf("git worktree command failed (unsupported or environment limitation): %v", err)
	}

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			CWD: wtDir,
		},
		GitCacheTTLSeconds: 5,
	}
	settings := types.DefaultSettings()

	branchWidget := widgets.GetWidget("git-branch")
	tWt, bWt, _ := branchWidget.Render(types.WidgetItem{Type: "git-branch"}, ctx, settings)
	plainWt := renderer.StripAnsi(tWt + bWt)

	if !strings.Contains(plainWt, "wt-branch") {
		t.Errorf("Expected worktree branch 'wt-branch', got %q", plainWt)
	}
}
