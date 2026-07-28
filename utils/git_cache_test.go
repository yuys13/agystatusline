package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveGitCwd(t *testing.T) {
	// Dummy data setup
	context := RenderContextDummy{
		CWD:        "/cwd",
		CurrentDir: "/current",
		ProjectDir: "/project",
	}

	actual := ResolveGitCwd(context)
	if actual != "/cwd" {
		t.Errorf("Expected '/cwd', got '%s'", actual)
	}

	context2 := RenderContextDummy{
		CurrentDir: "/current",
	}
	actual2 := ResolveGitCwd(context2)
	if actual2 != "/current" {
		t.Errorf("Expected '/current', got '%s'", actual2)
	}
}

func TestGitCache_TTLAndMtime(t *testing.T) {
	tempDir := t.TempDir()

	// Create a dummy git repo directory structure
	gitDir := filepath.Join(tempDir, ".git")
	err := os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	headPath := filepath.Join(gitDir, "HEAD")
	indexPath := filepath.Join(gitDir, "index")

	err = os.WriteFile(headPath, []byte("ref: refs/heads/main"), 0644)
	if err != nil {
		t.Fatalf("Failed to write HEAD: %v", err)
	}
	err = os.WriteFile(indexPath, []byte("dummy index"), 0644)
	if err != nil {
		t.Fatalf("Failed to write index: %v", err)
	}

	// Set cache directory to temp
	SetCacheDir(tempDir)

	// Create context
	ctx := RenderContextDummy{
		CWD:                tempDir,
		GitCacheTTLSeconds: 5,
	}

	// Call git command wrapper (which will invoke git, but since git isn't a mock,
	// we will test the caching mechanism directly on top of raw outputs)
	cacheKey := "status --porcelain"
	now := time.Now().UnixNano() / int64(time.Millisecond)

	// Set initial cache
	output := "M file.go"
	SetCacheEntry(tempDir, cacheKey, output, now, ctx.GitCacheTTLSeconds)

	// Read cache - should be hit (same mtimes and within TTL)
	cachedOutput, hit := GetCacheEntry(tempDir, cacheKey, ctx.GitCacheTTLSeconds)
	if !hit || cachedOutput != output {
		t.Errorf("Expected cache hit with '%s', got '%s' (hit=%t)", output, cachedOutput, hit)
	}

	// Wait or force TTL expiration
	expiredTime := now - 6000 // 6 seconds ago (TTL is 5s)
	SetCacheEntry(tempDir, cacheKey, output, expiredTime, ctx.GitCacheTTLSeconds)
	_, hit = GetCacheEntry(tempDir, cacheKey, ctx.GitCacheTTLSeconds)
	if hit {
		t.Errorf("Expected cache miss due to TTL expiration, but it hit")
	}

	// Reset cache and update file mtime to test mtime invalidation
	SetCacheEntry(tempDir, cacheKey, output, time.Now().UnixNano()/int64(time.Millisecond), ctx.GitCacheTTLSeconds)

	// Change mtime of HEAD
	newTime := time.Now().Add(1 * time.Hour)
	err = os.Chtimes(headPath, newTime, newTime)
	if err != nil {
		t.Fatalf("Failed to change mtime: %v", err)
	}

	_, hit = GetCacheEntry(tempDir, cacheKey, ctx.GitCacheTTLSeconds)
	if hit {
		t.Errorf("Expected cache miss due to HEAD mtime change, but it hit")
	}
}

// Dummy helper structs matching required fields in main program
type RenderContextDummy struct {
	CWD                string
	CurrentDir         string
	ProjectDir         string
	GitCacheTTLSeconds int
}

func (r RenderContextDummy) GetCwd() string {
	return r.CWD
}

func (r RenderContextDummy) GetWorkspaceCurrentDir() string {
	return r.CurrentDir
}

func (r RenderContextDummy) GetWorkspaceProjectDir() string {
	return r.ProjectDir
}

type PersistentGitCache struct {
	Version int                      `json:"version"`
	CWD     string                   `json:"cwd"`
	Entries map[string]GitCacheEntry `json:"entries"`
}

type GitCacheEntry struct {
	Output       *string `json:"output"`
	CreatedAt    int64   `json:"createdAt"`
	HeadMtimeMS  int64   `json:"headMtimeMs"`
	IndexMtimeMS int64   `json:"indexMtimeMs"`
}

func TestReadPersistentCache_ErrorHandling(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Non-existent file
	_, err := readPersistentCache(filepath.Join(tempDir, "nonexistent.json"))
	if err == nil {
		t.Errorf("Expected error when reading non-existent cache file, got nil")
	}

	// 2. Invalid JSON file
	invalidJsonPath := filepath.Join(tempDir, "invalid.json")
	_ = os.WriteFile(invalidJsonPath, []byte("invalid json{"), 0644)
	_, err = readPersistentCache(invalidJsonPath)
	if err == nil {
		t.Errorf("Expected error when reading invalid JSON cache file, got nil")
	}

	// 3. Invalid schema version
	invalidVerPath := filepath.Join(tempDir, "invalid_ver.json")
	_ = os.WriteFile(invalidVerPath, []byte(`{"version": 999}`), 0644)
	_, err = readPersistentCache(invalidVerPath)
	if err == nil || err.Error() != "invalid cache version" {
		t.Errorf("Expected 'invalid cache version' error, got %v", err)
	}
}

func TestWritePersistentCache_CleanupOnFailure(t *testing.T) {
	tempDir := t.TempDir()
	invalidPath := filepath.Join(tempDir, "dir_target")
	_ = os.Mkdir(invalidPath, 0755)

	cache := &PersistentGitCacheInternal{
		Version: GitCacheSchemaVersion,
		CWD:     tempDir,
		Entries: make(map[string]GitCacheEntryInternal),
	}

	// Writing to a directory path should fail gracefully without leaving temporary files
	writePersistentCache(invalidPath, cache)

	files, _ := os.ReadDir(tempDir)
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".tmp" {
			t.Errorf("Expected no temporary files remaining in %s, but found %s", tempDir, f.Name())
		}
	}
}

func TestReadGitDirFile(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Non-existent file
	_, err := readGitDirFile(filepath.Join(tempDir, "nonexistent"))
	if err == nil {
		t.Errorf("Expected error for non-existent file, got nil")
	}

	// 2. Invalid content (no gitdir: prefix)
	invalidFile := filepath.Join(tempDir, "invalid_git")
	_ = os.WriteFile(invalidFile, []byte("invalid content"), 0644)
	_, err = readGitDirFile(invalidFile)
	if err == nil || err.Error() != "not a gitdir file" {
		t.Errorf("Expected 'not a gitdir file' error, got %v", err)
	}

	// 3. Valid relative gitdir file
	relFile := filepath.Join(tempDir, "rel_git")
	_ = os.WriteFile(relFile, []byte("gitdir: ../target_git"), 0644)
	target, err := readGitDirFile(relFile)
	if err != nil {
		t.Fatalf("Unexpected error for relative gitdir: %v", err)
	}
	expectedTarget, _ := filepath.Abs(filepath.Join(tempDir, "../target_git"))
	if target != expectedTarget {
		t.Errorf("Expected %q, got %q", expectedTarget, target)
	}

	// 4. Valid absolute gitdir file
	absTarget := filepath.Join(tempDir, "abs_target")
	absFile := filepath.Join(tempDir, "abs_git")
	_ = os.WriteFile(absFile, []byte("gitdir: "+absTarget), 0644)
	targetAbs, err := readGitDirFile(absFile)
	if err != nil || targetAbs != absTarget {
		t.Errorf("Expected %q, got %q (err=%v)", absTarget, targetAbs, err)
	}
}

func TestDiscoverGitDir(t *testing.T) {
	tempDir := t.TempDir()

	// 1. No git dir found
	_, err := discoverGitDir(tempDir)
	if err == nil {
		t.Errorf("Expected error when no git dir exists, got nil")
	}

	// 2. Git dir as file (worktree)
	worktreeDir := filepath.Join(tempDir, "worktree")
	_ = os.Mkdir(worktreeDir, 0755)

	realGitDir := filepath.Join(tempDir, "real_git")
	_ = os.Mkdir(realGitDir, 0755)

	gitFile := filepath.Join(worktreeDir, ".git")
	_ = os.WriteFile(gitFile, []byte("gitdir: "+realGitDir), 0644)

	found, err := discoverGitDir(worktreeDir)
	if err != nil {
		t.Fatalf("Unexpected error discovering git dir via worktree file: %v", err)
	}
	if found != realGitDir {
		t.Errorf("Expected %q, got %q", realGitDir, found)
	}
}

func TestNormalizeDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Non-existent path
	_, err := normalizeDirectory(filepath.Join(tempDir, "nonexistent"))
	if err == nil {
		t.Errorf("Expected error for non-existent path, got nil")
	}

	// 2. Directory path
	normDir, err := normalizeDirectory(tempDir)
	if err != nil || normDir != tempDir {
		t.Errorf("Expected %q, got %q (err=%v)", tempDir, normDir, err)
	}

	// 3. File path
	filePath := filepath.Join(tempDir, "file.txt")
	_ = os.WriteFile(filePath, []byte("test"), 0644)
	normFile, err := normalizeDirectory(filePath)
	if err != nil || normFile != tempDir {
		t.Errorf("Expected parent dir %q for file path, got %q (err=%v)", tempDir, normFile, err)
	}
}

func TestClearGitCache(t *testing.T) {
	tempDir := t.TempDir()
	SetCacheDir(tempDir)
	SetCacheEntry(tempDir, "test-cmd", "output", 1000, 10)

	ClearGitCache()

	_, hit := GetCacheEntry(tempDir, "test-cmd", 10)
	if hit {
		t.Errorf("Expected cache miss after ClearGitCache(), but hit was true")
	}
}

func TestRunGit(t *testing.T) {
	tempDir := t.TempDir()
	SetCacheDir(tempDir)
	ClearGitCache()

	ctx := RenderContextDummy{CWD: tempDir, GitCacheTTLSeconds: 5}

	// 1. Empty CWD
	emptyCtx := RenderContextDummy{CWD: ""}
	_, err := RunGit("status", emptyCtx, 5, nil)
	if err == nil {
		t.Errorf("Expected error when CWD is empty")
	}

	// 2. Empty command
	_, err = RunGit("", ctx, 5, nil)
	if err == nil {
		t.Errorf("Expected error when command is empty")
	}

	// 3. Successful execution and caching
	execCount := 0
	mockExec := func(args []string, cwd string) (string, error) {
		execCount++
		return "branch-name", nil
	}

	out1, err := RunGit("symbolic-ref --short HEAD", ctx, 5, mockExec)
	if err != nil || out1 != "branch-name" {
		t.Errorf("Expected 'branch-name', got %q, err=%v", out1, err)
	}
	if execCount != 1 {
		t.Errorf("Expected mockExec to be called once, got %d", execCount)
	}

	// 4. Second call should hit cache and not invoke mockExec again
	out2, err := RunGit("symbolic-ref --short HEAD", ctx, 5, mockExec)
	if err != nil || out2 != "branch-name" {
		t.Errorf("Expected 'branch-name', got %q, err=%v", out2, err)
	}
	if execCount != 1 {
		t.Errorf("Expected mockExec execution count to remain 1 (cache hit), got %d", execCount)
	}
}
