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
