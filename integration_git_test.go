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

	repoWidget := widgets.GetWidget("git-repo")
	if repoWidget == nil {
		t.Fatalf("git-repo widget not found")
	}
	repoItem := types.WidgetItem{Type: "git-repo"}
	titleR, bodyR, errR := repoWidget.Render(repoItem, ctx, settings)
	if errR != nil {
		t.Fatalf("Unexpected error rendering repo widget: %v", errR)
	}
	plainRepo := renderer.StripAnsi(titleR + bodyR)
	expectedRepo := filepath.Base(repoDir)
	if plainRepo != expectedRepo {
		t.Errorf("Expected repo widget to be %q, got %q", expectedRepo, plainRepo)
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

	repoWidget := widgets.GetWidget("git-repo")
	tRepo, bRepo, _ := repoWidget.Render(types.WidgetItem{Type: "git-repo"}, ctx, settings)
	plainRepo := renderer.StripAnsi(tRepo + bRepo)
	if !strings.Contains(plainRepo, "no git") {
		t.Errorf("Expected repo widget to return '(no git)', got %q", plainRepo)
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

// TestIntegration_Git_AheadBehind tests git-ahead-behind with real local and upstream repos.
func TestIntegration_Git_AheadBehind(t *testing.T) {
	cacheDir := setupTestCacheDir(t)
	widgets.RegisterAll()

	repoDir := setupGitRepo(t, "main")
	remoteDir := t.TempDir()

	// 1. Initialize bare remote repository and set upstream
	cmdBare := exec.Command("git", "init", "--bare", "-b", "main", remoteDir)
	if err := cmdBare.Run(); err != nil {
		if err := exec.Command("git", "init", "--bare", remoteDir).Run(); err != nil {
			t.Fatalf("Failed to init bare remote: %v", err)
		}
		_ = exec.Command("git", "-C", remoteDir, "symbolic-ref", "HEAD", "refs/heads/main").Run()
	}
	_ = exec.Command("git", "-C", repoDir, "remote", "add", "origin", remoteDir).Run()
	if err := exec.Command("git", "-C", repoDir, "push", "-u", "origin", "main").Run(); err != nil {
		t.Fatalf("Failed to push initial commit to remote: %v", err)
	}

	clearAllCache := func() {
		_ = os.RemoveAll(filepath.Join(cacheDir, "git-cache"))
		utils.ClearGitCache()
	}

	clearAllCache()
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			CWD: repoDir,
		},
		GitCacheTTLSeconds: 5,
	}
	settings := types.DefaultSettings()

	w := widgets.GetWidget("git-ahead-behind")
	if w == nil {
		t.Fatalf("git-ahead-behind widget not found")
	}

	// Case 1: Up-to-date with upstream
	title, body, err := w.Render(types.WidgetItem{Type: "git-ahead-behind"}, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if plain := renderer.StripAnsi(title + body); plain != "≡" {
		t.Errorf("Expected up-to-date '≡', got %q", plain)
	}

	// Case 2: Ahead by 1 commit
	testFile := filepath.Join(repoDir, "ahead.txt")
	_ = os.WriteFile(testFile, []byte("ahead content\n"), 0644)
	_ = exec.Command("git", "-C", repoDir, "add", "ahead.txt").Run()
	_ = exec.Command("git", "-C", repoDir, "commit", "-m", "Commit ahead").Run()

	clearAllCache()
	title, body, _ = w.Render(types.WidgetItem{Type: "git-ahead-behind"}, ctx, settings)
	if plain := renderer.StripAnsi(title + body); plain != "↑1" {
		t.Errorf("Expected ahead '↑1', got %q", plain)
	}

	// Case 3: Diverged (1 ahead, 1 behind)
	// Clone remote to a secondary working tree and push a divergent commit
	cloneDir := t.TempDir()
	if out, err := exec.Command("git", "clone", "-b", "main", remoteDir, cloneDir).CombinedOutput(); err != nil {
		t.Fatalf("Clone failed: %v, out: %s", err, string(out))
	}
	_ = exec.Command("git", "-C", cloneDir, "config", "user.name", "Test User").Run()
	_ = exec.Command("git", "-C", cloneDir, "config", "user.email", "test@example.com").Run()
	remoteCommitFile := filepath.Join(cloneDir, "remote.txt")
	_ = os.WriteFile(remoteCommitFile, []byte("remote commit\n"), 0644)
	_ = exec.Command("git", "-C", cloneDir, "add", "remote.txt").Run()
	_ = exec.Command("git", "-C", cloneDir, "commit", "-m", "Remote commit").Run()
	if out, err := exec.Command("git", "-C", cloneDir, "push", "origin", "main").CombinedOutput(); err != nil {
		t.Fatalf("Push failed: %v, out: %s", err, string(out))
	}

	// Fetch remote in primary repo
	if out, err := exec.Command("git", "-C", repoDir, "fetch", "origin").CombinedOutput(); err != nil {
		t.Fatalf("Fetch failed: %v, out: %s", err, string(out))
	}

	clearAllCache()
	title, body, _ = w.Render(types.WidgetItem{Type: "git-ahead-behind"}, ctx, settings)
	if plain := renderer.StripAnsi(title + body); plain != "↑1 ↓1" {
		t.Errorf("Expected diverged '↑1 ↓1', got %q", plain)
	}

	// Case 4: Behind only (reset local branch to origin/main^ to be behind)
	_ = exec.Command("git", "-C", repoDir, "reset", "--hard", "origin/main").Run()
	_ = exec.Command("git", "-C", repoDir, "reset", "--hard", "HEAD~1").Run()

	clearAllCache()
	title, body, _ = w.Render(types.WidgetItem{Type: "git-ahead-behind"}, ctx, settings)
	if plain := renderer.StripAnsi(title + body); plain != "↓1" {
		t.Errorf("Expected behind '↓1', got %q", plain)
	}

	// Case 5: Branch with no upstream configured
	_ = exec.Command("git", "-C", repoDir, "checkout", "-b", "local-only-branch").Run()

	clearAllCache()
	title, body, _ = w.Render(types.WidgetItem{Type: "git-ahead-behind"}, ctx, settings)
	if plain := renderer.StripAnsi(title + body); plain != "(no upstream)" {
		t.Errorf("Expected '(no upstream)', got %q", plain)
	}

	// Case 6: Custom symbol
	title, body, _ = w.Render(types.WidgetItem{Type: "git-ahead-behind", Symbol: "⇅ "}, ctx, settings)
	if plain := renderer.StripAnsi(title + body); plain != "⇅ (no upstream)" {
		t.Errorf("Expected '⇅ (no upstream)', got %q", plain)
	}

	// Case 7: Custom symbol with Raw=true
	title, body, _ = w.Render(types.WidgetItem{Type: "git-ahead-behind", Symbol: "⇅ ", Raw: true}, ctx, settings)
	if plain := renderer.StripAnsi(title + body); plain != "(no upstream)" {
		t.Errorf("Expected '(no upstream)', got %q", plain)
	}

	// Case 8: Non-git directory
	nonGitCtx := types.RenderContext{
		Data: types.StatusJSON{
			CWD: t.TempDir(),
		},
		GitCacheTTLSeconds: 5,
	}
	title, body, _ = w.Render(types.WidgetItem{Type: "git-ahead-behind"}, nonGitCtx, settings)
	if plain := renderer.StripAnsi(title + body); plain != "(no git)" {
		t.Errorf("Expected '(no git)', got %q", plain)
	}
}

// TestIntegration_Git_DetachedHEAD tests Git branch widget in detached HEAD state in a real Git repo.
func TestIntegration_Git_DetachedHEAD(t *testing.T) {
	cacheDir := setupTestCacheDir(t)
	widgets.RegisterAll()

	repoDir := setupGitRepo(t, "main")
	settings := types.DefaultSettings()

	clearAllCache := func() {
		_ = os.RemoveAll(filepath.Join(cacheDir, "git-cache"))
		utils.ClearGitCache()
	}

	// Get initial commit short sha
	shaOut, err := exec.Command("git", "-C", repoDir, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		t.Fatalf("Failed to rev-parse: %v", err)
	}
	shortSHA := strings.TrimSpace(string(shaOut))

	// Detach HEAD at initial commit
	if err := exec.Command("git", "-C", repoDir, "checkout", "--detach", "HEAD").Run(); err != nil {
		t.Fatalf("Failed to detach HEAD: %v", err)
	}

	w := widgets.GetWidget("git-branch")
	if w == nil {
		t.Fatalf("Git branch widget not found")
	}

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			CWD: repoDir,
		},
		GitCacheTTLSeconds: 5,
	}

	// 1. Detached HEAD at commit SHA (clean)
	clearAllCache()
	title, body, err := w.Render(types.WidgetItem{Type: "git-branch"}, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if plain := renderer.StripAnsi(title + body); plain != "⎇ ("+shortSHA+")" {
		t.Errorf("Expected '⎇ (%s)', got %q", shortSHA, plain)
	}
	if color := w.GetBodyColor(types.WidgetItem{Type: "git-branch"}, ctx); color != "brightBlue" {
		t.Errorf("Expected body color 'brightBlue', got %q", color)
	}

	// 2. Detached HEAD at commit SHA (dirty)
	if err := os.WriteFile(filepath.Join(repoDir, "dirty.txt"), []byte("dirty content"), 0644); err != nil {
		t.Fatalf("Failed to write dirty file: %v", err)
	}
	clearAllCache()
	title, body, err = w.Render(types.WidgetItem{Type: "git-branch"}, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if plain := renderer.StripAnsi(title + body); plain != "⎇ ("+shortSHA+")*" {
		t.Errorf("Expected '⎇ (%s)*', got %q", shortSHA, plain)
	}
	if color := w.GetBodyColor(types.WidgetItem{Type: "git-branch"}, ctx); color != "brightRed" {
		t.Errorf("Expected body color 'brightRed', got %q", color)
	}

	// 3. Detached HEAD with Raw=true (dirty)
	title, body, err = w.Render(types.WidgetItem{Type: "git-branch", Raw: true}, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if plain := renderer.StripAnsi(title + body); plain != "("+shortSHA+")*" {
		t.Errorf("Expected '(%s)*', got %q", shortSHA, plain)
	}

	// 4. Detached HEAD with tag checkout (clean)
	_ = exec.Command("git", "-C", repoDir, "clean", "-fd").Run()
	if err := exec.Command("git", "-C", repoDir, "tag", "v1.2.0").Run(); err != nil {
		t.Fatalf("Failed to tag: %v", err)
	}
	clearAllCache()
	title, body, err = w.Render(types.WidgetItem{Type: "git-branch"}, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if plain := renderer.StripAnsi(title + body); plain != "⎇ (v1.2.0)" {
		t.Errorf("Expected '⎇ (v1.2.0)', got %q", plain)
	}
	if color := w.GetBodyColor(types.WidgetItem{Type: "git-branch"}, ctx); color != "brightBlue" {
		t.Errorf("Expected body color 'brightBlue', got %q", color)
	}

	// 5. Detached HEAD with tag checkout (dirty) + custom symbol
	if err := os.WriteFile(filepath.Join(repoDir, "dirty.txt"), []byte("dirty content 2"), 0644); err != nil {
		t.Fatalf("Failed to write dirty file: %v", err)
	}
	clearAllCache()
	title, body, err = w.Render(types.WidgetItem{Type: "git-branch", Symbol: "🌿 "}, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if plain := renderer.StripAnsi(title + body); plain != "🌿 (v1.2.0)*" {
		t.Errorf("Expected '🌿 (v1.2.0)*', got %q", plain)
	}
	if color := w.GetBodyColor(types.WidgetItem{Type: "git-branch", Symbol: "🌿 "}, ctx); color != "brightRed" {
		t.Errorf("Expected body color 'brightRed', got %q", color)
	}

	// 6. Detached HEAD with commit ahead of tag falls back to commit SHA
	_ = exec.Command("git", "-C", repoDir, "add", ".").Run()
	_ = exec.Command("git", "-C", repoDir, "commit", "-m", "second commit").Run()
	sha2Out, err := exec.Command("git", "-C", repoDir, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		t.Fatalf("Failed to rev-parse: %v", err)
	}
	shortSHA2 := strings.TrimSpace(string(sha2Out))

	clearAllCache()
	title, body, err = w.Render(types.WidgetItem{Type: "git-branch"}, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if plain := renderer.StripAnsi(title + body); plain != "⎇ ("+shortSHA2+")" {
		t.Errorf("Expected '⎇ (%s)', got %q", shortSHA2, plain)
	}
	if color := w.GetBodyColor(types.WidgetItem{Type: "git-branch"}, ctx); color != "brightBlue" {
		t.Errorf("Expected body color 'brightBlue', got %q", color)
	}

	// 7. Other git widgets in detached HEAD (ahead-behind and repo)
	abWidget := widgets.GetWidget("git-ahead-behind")
	if abWidget != nil {
		clearAllCache()
		_, abBody, err := abWidget.Render(types.WidgetItem{Type: "git-ahead-behind"}, ctx, settings)
		if err != nil {
			t.Fatalf("Ahead/behind render error in detached HEAD: %v", err)
		}
		if plain := renderer.StripAnsi(abBody); plain != "(no upstream)" {
			t.Errorf("Expected ahead-behind '(no upstream)' in detached HEAD, got %q", plain)
		}
	}

	repoWidget := widgets.GetWidget("git-repo")
	if repoWidget != nil {
		clearAllCache()
		_, repoBody, err := repoWidget.Render(types.WidgetItem{Type: "git-repo"}, ctx, settings)
		if err != nil {
			t.Fatalf("Repo render error in detached HEAD: %v", err)
		}
		expectedRepoName := filepath.Base(repoDir)
		if plain := renderer.StripAnsi(repoBody); plain != expectedRepoName {
			t.Errorf("Expected repo %q in detached HEAD, got %q", expectedRepoName, plain)
		}
	}

	// 8. Detached HEAD in a git worktree
	wtDir := filepath.Join(t.TempDir(), "wt-detached")
	wtCmd := exec.Command("git", "-C", repoDir, "worktree", "add", "--detach", wtDir, "HEAD")
	if err := wtCmd.Run(); err != nil {
		t.Skipf("git worktree command failed (unsupported or environment limitation): %v", err)
	}
	ctxWt := types.RenderContext{
		Data: types.StatusJSON{
			CWD: wtDir,
		},
		GitCacheTTLSeconds: 5,
	}

	// 8a. Clean worktree
	clearAllCache()
	title, body, err = w.Render(types.WidgetItem{Type: "git-branch"}, ctxWt, settings)
	if err != nil {
		t.Fatalf("Worktree render error: %v", err)
	}
	if plain := renderer.StripAnsi(title + body); plain != "⎇ ("+shortSHA2+")" {
		t.Errorf("Expected worktree branch '⎇ (%s)', got %q", shortSHA2, plain)
	}
	if color := w.GetBodyColor(types.WidgetItem{Type: "git-branch"}, ctxWt); color != "brightBlue" {
		t.Errorf("Expected worktree body color 'brightBlue', got %q", color)
	}

	// 8b. Dirty worktree with Raw=true
	if err := os.WriteFile(filepath.Join(wtDir, "wt-dirty.txt"), []byte("worktree dirty"), 0644); err != nil {
		t.Fatalf("Failed to write dirty file in worktree: %v", err)
	}
	clearAllCache()
	title, body, err = w.Render(types.WidgetItem{Type: "git-branch", Raw: true}, ctxWt, settings)
	if err != nil {
		t.Fatalf("Worktree render error (dirty): %v", err)
	}
	if plain := renderer.StripAnsi(title + body); plain != "("+shortSHA2+")*" {
		t.Errorf("Expected worktree branch '(%s)*', got %q", shortSHA2, plain)
	}
	if color := w.GetBodyColor(types.WidgetItem{Type: "git-branch", Raw: true}, ctxWt); color != "brightRed" {
		t.Errorf("Expected worktree body color 'brightRed', got %q", color)
	}
}
