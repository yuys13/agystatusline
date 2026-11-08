package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestResolveGitCwd(t *testing.T) {
	tests := []struct {
		name string
		ctx  RenderContextDummy
		want string
	}{
		{
			name: "CWD preferred over CurrentDir and ProjectDir",
			ctx: RenderContextDummy{
				CWD:        "/cwd",
				CurrentDir: "/current",
				ProjectDir: "/project",
			},
			want: "/cwd",
		},
		{
			name: "Fallback to CurrentDir when CWD is empty",
			ctx: RenderContextDummy{
				CurrentDir: "/current",
				ProjectDir: "/project",
			},
			want: "/current",
		},
		{
			name: "Fallback to ProjectDir when CWD and CurrentDir are empty",
			ctx: RenderContextDummy{
				ProjectDir: "/project",
			},
			want: "/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveGitCwd(tt.ctx)
			if got != tt.want {
				t.Errorf("ResolveGitCwd() = %q, want %q", got, tt.want)
			}
		})
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
		t.Errorf("Expected cache hit with %q, got %q (hit=%t)", output, cachedOutput, hit)
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
	t.Run("non-existent file", func(t *testing.T) {
		tempDir := t.TempDir()
		_, err := readPersistentCache(filepath.Join(tempDir, "nonexistent.json"))
		if err == nil {
			t.Fatalf("readPersistentCache() expected error for nonexistent file, got nil")
		}
	})

	t.Run("invalid JSON file", func(t *testing.T) {
		tempDir := t.TempDir()
		invalidJsonPath := filepath.Join(tempDir, "invalid.json")
		_ = os.WriteFile(invalidJsonPath, []byte("invalid json{"), 0644)
		_, err := readPersistentCache(invalidJsonPath)
		if err == nil {
			t.Fatalf("readPersistentCache() expected error for invalid JSON, got nil")
		}
	})

	t.Run("invalid schema version", func(t *testing.T) {
		tempDir := t.TempDir()
		invalidVerPath := filepath.Join(tempDir, "invalid_ver.json")
		_ = os.WriteFile(invalidVerPath, []byte(`{"version": 999}`), 0644)
		_, err := readPersistentCache(invalidVerPath)
		if err == nil {
			t.Fatalf("readPersistentCache() expected error, got nil")
		}
		if err.Error() != "invalid cache version" {
			t.Errorf("readPersistentCache() error = %v, want %q", err, "invalid cache version")
		}
	})
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
			t.Errorf("Expected no temporary files remaining in %q, but found %q", tempDir, f.Name())
		}
	}
}

func TestReadGitDirFile(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			name  string
			setup func(dir string) (filePath string, want string)
		}{
			{
				name: "Valid relative gitdir file",
				setup: func(dir string) (string, string) {
					relFile := filepath.Join(dir, "rel_git")
					_ = os.WriteFile(relFile, []byte("gitdir: ../target_git"), 0644)
					expectedTarget, _ := filepath.Abs(filepath.Join(dir, "../target_git"))
					return relFile, expectedTarget
				},
			},
			{
				name: "Valid absolute gitdir file",
				setup: func(dir string) (string, string) {
					absTarget := filepath.Join(dir, "abs_target")
					absFile := filepath.Join(dir, "abs_git")
					_ = os.WriteFile(absFile, []byte("gitdir: "+absTarget), 0644)
					return absFile, absTarget
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tempDir := t.TempDir()
				path, want := tt.setup(tempDir)
				got, err := readGitDirFile(path)
				if err != nil {
					t.Fatalf("readGitDirFile() unexpected error: %v", err)
				}
				if got != want {
					t.Errorf("readGitDirFile() = %q, want %q", got, want)
				}
			})
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Run("non-existent file", func(t *testing.T) {
			tempDir := t.TempDir()
			path := filepath.Join(tempDir, "nonexistent")
			_, err := readGitDirFile(path)
			if err == nil {
				t.Fatalf("readGitDirFile() expected error for nonexistent file, got nil")
			}
		})

		t.Run("invalid content without gitdir prefix", func(t *testing.T) {
			tempDir := t.TempDir()
			invalidFile := filepath.Join(tempDir, "invalid_git")
			_ = os.WriteFile(invalidFile, []byte("invalid content"), 0644)
			_, err := readGitDirFile(invalidFile)
			if err == nil {
				t.Fatalf("readGitDirFile() expected error, got nil")
			}
			if err.Error() != "not a gitdir file" {
				t.Errorf("readGitDirFile() error = %v, want %q", err, "not a gitdir file")
			}
		})
	})
}

func TestDiscoverGitDir(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			name  string
			setup func(dir string) (searchDir string, want string)
		}{
			{
				name: "Standard .git directory",
				setup: func(dir string) (string, string) {
					gitDir := filepath.Join(dir, ".git")
					_ = os.Mkdir(gitDir, 0755)
					return dir, gitDir
				},
			},
			{
				name: "Git dir as file (worktree)",
				setup: func(dir string) (string, string) {
					worktreeDir := filepath.Join(dir, "worktree")
					_ = os.Mkdir(worktreeDir, 0755)

					realGitDir := filepath.Join(dir, "real_git")
					_ = os.Mkdir(realGitDir, 0755)

					gitFile := filepath.Join(worktreeDir, ".git")
					_ = os.WriteFile(gitFile, []byte("gitdir: "+realGitDir), 0644)
					return worktreeDir, realGitDir
				},
			},
			{
				name: "Parent directory contains .git directory",
				setup: func(dir string) (string, string) {
					gitDir := filepath.Join(dir, ".git")
					_ = os.Mkdir(gitDir, 0755)
					subDir := filepath.Join(dir, "sub", "child")
					_ = os.MkdirAll(subDir, 0755)
					return subDir, gitDir
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tempDir := t.TempDir()
				searchDir, want := tt.setup(tempDir)
				got, err := discoverGitDir(searchDir)
				if err != nil {
					t.Fatalf("discoverGitDir() unexpected error: %v", err)
				}
				if got != want {
					t.Errorf("discoverGitDir() = %q, want %q", got, want)
				}
			})
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := []struct {
			name  string
			setup func(dir string) string
		}{
			{
				name: "No git dir found",
				setup: func(dir string) string {
					return dir
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tempDir := t.TempDir()
				searchDir := tt.setup(tempDir)
				got, err := discoverGitDir(searchDir)
				if err == nil {
					t.Fatalf("discoverGitDir() expected error for directory without git, got %q", got)
				}
			})
		}
	})
}

func TestNormalizeDirectory(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			name  string
			setup func(dir string) (inputPath string, want string)
		}{
			{
				name: "Directory path",
				setup: func(dir string) (string, string) {
					return dir, dir
				},
			},
			{
				name: "File path returns parent directory",
				setup: func(dir string) (string, string) {
					filePath := filepath.Join(dir, "file.txt")
					_ = os.WriteFile(filePath, []byte("test"), 0644)
					return filePath, dir
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tempDir := t.TempDir()
				input, want := tt.setup(tempDir)
				got, err := normalizeDirectory(input)
				if err != nil {
					t.Fatalf("normalizeDirectory() unexpected error: %v", err)
				}
				if got != want {
					t.Errorf("normalizeDirectory() = %q, want %q", got, want)
				}
			})
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := []struct {
			name  string
			setup func(dir string) string
		}{
			{
				name: "Non-existent path",
				setup: func(dir string) string {
					return filepath.Join(dir, "nonexistent")
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tempDir := t.TempDir()
				input := tt.setup(tempDir)
				got, err := normalizeDirectory(input)
				if err == nil {
					t.Fatalf("normalizeDirectory() expected error for nonexistent path, got %q", got)
				}
			})
		}
	})
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
	t.Run("argument validation", func(t *testing.T) {
		tests := []struct {
			name    string
			command string
			ctx     CwdResolver
		}{
			{
				name:    "Empty CWD",
				command: "status",
				ctx:     RenderContextDummy{CWD: ""},
			},
			{
				name:    "Empty command",
				command: "",
				ctx:     RenderContextDummy{CWD: "/tmp", GitCacheTTLSeconds: 5},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := RunGit(tt.command, tt.ctx, 5, nil)
				if err == nil {
					t.Fatalf("RunGit() expected error for %s, got nil", tt.name)
				}
			})
		}
	})

	t.Run("successful execution and caching", func(t *testing.T) {
		tempDir := t.TempDir()
		SetCacheDir(tempDir)
		ClearGitCache()

		ctx := RenderContextDummy{CWD: tempDir, GitCacheTTLSeconds: 5}
		execCount := 0
		mockExec := func(args []string, cwd string) (string, error) {
			execCount++
			return "branch-name", nil
		}

		got, err := RunGit("symbolic-ref --short HEAD", ctx, 5, mockExec)
		if err != nil {
			t.Fatalf("RunGit() unexpected error: %v", err)
		}
		if got != "branch-name" {
			t.Errorf("RunGit() = %q, want %q", got, "branch-name")
		}
		if execCount != 1 {
			t.Errorf("Expected mockExec to be called once, got %d", execCount)
		}

		// Second call should hit cache and not invoke mockExec again
		out2, err := RunGit("symbolic-ref --short HEAD", ctx, 5, mockExec)
		if err != nil {
			t.Fatalf("RunGit() second call unexpected error: %v", err)
		}
		if out2 != "branch-name" {
			t.Errorf("Expected 'branch-name', got %q", out2)
		}
		if execCount != 1 {
			t.Errorf("Expected mockExec execution count to remain 1 (cache hit), got %d", execCount)
		}
	})
}

func TestGitCache_AdversarialRaceConditions(t *testing.T) {
	baseDir := t.TempDir()

	// Setup multiple mock git repositories
	repoDirs := make([]string, 3)
	for i := range repoDirs {
		rDir := filepath.Join(baseDir, fmt.Sprintf("repo_%d", i))
		if err := os.MkdirAll(filepath.Join(rDir, ".git"), 0755); err != nil {
			t.Fatalf("Failed to create mock repo dir: %v", err)
		}
		_ = os.WriteFile(filepath.Join(rDir, ".git", "HEAD"), []byte("ref: refs/heads/main"), 0644)
		_ = os.WriteFile(filepath.Join(rDir, ".git", "index"), []byte("index binary data"), 0644)
		repoDirs[i] = rDir
	}

	cacheDir1 := filepath.Join(baseDir, "cache_1")
	cacheDir2 := filepath.Join(baseDir, "cache_2")
	SetCacheDir(cacheDir1)

	var wg sync.WaitGroup
	numWorkers := 20
	opsPerWorker := 30

	commands := []string{
		"status --porcelain",
		"symbolic-ref --short HEAD",
		"rev-parse HEAD",
		"diff --stat",
		"",
	}

	for i := range numWorkers {
		wg.Add(1)
		workerID := i
		go func() {
			defer wg.Done()

			for j := range opsPerWorker {
				repo := repoDirs[(workerID+j)%len(repoDirs)]
				cmd := commands[(workerID+j)%len(commands)]
				ctx := RenderContextDummy{
					CWD:                repo,
					GitCacheTTLSeconds: 5,
				}

				switch (workerID + j) % 6 {
				case 0: // RunGit execution
					mockExec := func(args []string, cwd string) (string, error) {
						return fmt.Sprintf("output_%d_%d", workerID, j), nil
					}
					out, err := RunGit(cmd, ctx, 5, mockExec)
					if cmd != "" && err != nil {
						t.Errorf("Worker %d op %d: RunGit unexpected error: %v", workerID, j, err)
					}
					if cmd != "" && out == "" {
						t.Errorf("Worker %d op %d: RunGit expected non-empty output", workerID, j)
					}

				case 1: // GetCacheEntry & SetCacheEntry
					now := time.Now().UnixNano() / int64(time.Millisecond)
					_, _ = GetCacheEntry(repo, cmd, 5)
					SetCacheEntry(repo, cmd, fmt.Sprintf("val_%d_%d", workerID, j), now, 5)

				case 2: // ClearGitCache
					if j%5 == 0 {
						ClearGitCache()
					}

				case 3: // SetCacheDir toggling
					if j%10 == 0 {
						if j%20 == 0 {
							SetCacheDir(cacheDir1)
						} else {
							SetCacheDir(cacheDir2)
						}
					}

				case 4: // Mutate repo git files (HEAD / index modification)
					headPath := filepath.Join(repo, ".git", "HEAD")
					_ = os.WriteFile(headPath, []byte(fmt.Sprintf("ref: refs/heads/branch_%d_%d", workerID, j)), 0644)

				case 5: // Mutate persistent cache file corruptively
					meta, err := getGitRepoMetadata(repo)
					if err == nil && meta != nil {
						_ = os.WriteFile(meta.CachePath, []byte("{invalid_json_corrupt_data"), 0644)
					}
				}
			}
		}()
	}

	wg.Wait()
	// Restore clean cache dir
	SetCacheDir(t.TempDir())
	ClearGitCache()
}

func TestGitCache_RapidConcurrentCacheAccess(t *testing.T) {
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	_ = os.MkdirAll(gitDir, 0755)
	_ = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main"), 0644)
	_ = os.WriteFile(filepath.Join(gitDir, "index"), []byte("index content"), 0644)

	SetCacheDir(tempDir)
	ClearGitCache()

	ctx := RenderContextDummy{
		CWD:                tempDir,
		GitCacheTTLSeconds: 10,
	}

	execCount := 0
	var execMutex sync.Mutex
	mockExec := func(args []string, cwd string) (string, error) {
		execMutex.Lock()
		execCount++
		execMutex.Unlock()
		time.Sleep(1 * time.Millisecond)
		return "main", nil
	}

	// Prime initial cache entry
	_, err := RunGit("symbolic-ref --short HEAD", ctx, 10, mockExec)
	if err != nil {
		t.Fatalf("Initial RunGit failed: %v", err)
	}

	numGoroutines := 50
	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines)

	for range numGoroutines {
		wg.Go(func() {
			res, err := RunGit("symbolic-ref --short HEAD", ctx, 10, mockExec)
			if err != nil {
				errCh <- err
				return
			}
			if res != "main" {
				errCh <- fmt.Errorf("expected 'main', got %q", res)
				return
			}
		})
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("Concurrent RunGit error: %v", err)
	}

	execMutex.Lock()
	count := execCount
	execMutex.Unlock()

	if count != 1 {
		t.Errorf("Expected mockExec to be called exactly 1 time due to cache hit, got %d", count)
	}
}

func TestGitCache_CorruptDiskCacheRecovery(t *testing.T) {
	corruptPayloads := []struct {
		name    string
		content []byte
	}{
		{name: "Truncated JSON", content: []byte(`{"version": 1, "cwd":`)},
		{name: "Null JSON", content: []byte(`null`)},
		{name: "Array JSON instead of Object", content: []byte(`[1, 2, 3]`)},
		{name: "Wrong Schema Version", content: []byte(`{"version": 99, "cwd": "dir", "entries": {}}`)},
		{name: "Binary Garbage", content: []byte{0x00, 0xFF, 0xFE, 0xFD, 0x12, 0x34}},
		{name: "Empty File", content: []byte{}},
		{name: "Invalid Entry Structure", content: []byte(`{"version": 1, "cwd": "dir", "entries": {"cmd": "not_an_object"}}`)},
	}

	for _, tc := range corruptPayloads {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			gitDir := filepath.Join(tempDir, ".git")
			_ = os.MkdirAll(gitDir, 0755)
			_ = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main"), 0644)

			SetCacheDir(tempDir)
			ClearGitCache()

			meta, err := getGitRepoMetadata(tempDir)
			if err != nil {
				t.Fatalf("Failed to get repo metadata: %v", err)
			}

			// Pre-create corrupt file
			_ = os.MkdirAll(filepath.Dir(meta.CachePath), 0755)
			if err := os.WriteFile(meta.CachePath, tc.content, 0644); err != nil {
				t.Fatalf("Failed to write corrupt cache file: %v", err)
			}

			// 1. GetCacheEntry should handle corrupt file gracefully and return cache miss
			out, hit := GetCacheEntry(tempDir, "status", 5)
			if hit {
				t.Errorf("Expected cache miss for corrupt payload %s, but hit=true with out=%q", tc.name, out)
			}

			// 2. SetCacheEntry should overwrite the corrupt file safely
			now := time.Now().UnixNano() / int64(time.Millisecond)
			SetCacheEntry(tempDir, "status", "valid_output", now, 5)

			// 3. Clear memory cache and re-read from persistent cache to verify recovery
			ClearGitCache()
			outRecovered, hitRecovered := GetCacheEntry(tempDir, "status", 5)
			if !hitRecovered || outRecovered != "valid_output" {
				t.Errorf("Failed to recover from corrupt cache file: hit=%v, out=%q", hitRecovered, outRecovered)
			}
		})
	}
}

func TestGitCache_BoundaryCasesAndEdgeInputs(t *testing.T) {
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	_ = os.MkdirAll(gitDir, 0755)
	_ = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main"), 0644)

	SetCacheDir(tempDir)
	ClearGitCache()

	t.Run("Negative TTL", func(t *testing.T) {
		now := time.Now().UnixNano() / int64(time.Millisecond)
		SetCacheEntry(tempDir, "cmd_neg_ttl", "out", now, -5)
		// Negative TTL means now - createdAt > ttlMs (-5000), so entry should be expired immediately
		_, hit := GetCacheEntry(tempDir, "cmd_neg_ttl", -5)
		if hit {
			t.Errorf("Expected cache miss for negative TTL, got hit")
		}
	})

	t.Run("Extreme Max TTL", func(t *testing.T) {
		now := time.Now().UnixNano() / int64(time.Millisecond)
		SetCacheEntry(tempDir, "cmd_max_ttl", "out_max", now, math.MaxInt32)
		out, hit := GetCacheEntry(tempDir, "cmd_max_ttl", math.MaxInt32)
		if !hit || out != "out_max" {
			t.Errorf("Expected cache hit for max TTL, got hit=%v, out=%q", hit, out)
		}
	})

	t.Run("Extreme Timestamps", func(t *testing.T) {
		SetCacheEntry(tempDir, "cmd_ts_zero", "out_zero", 0, 5)
		_, hit := GetCacheEntry(tempDir, "cmd_ts_zero", 5)
		if hit {
			t.Errorf("Expected cache miss for timestamp 0 with TTL 5, got hit")
		}

		SetCacheEntry(tempDir, "cmd_ts_future", "out_future", math.MaxInt64, 5)
		// Future timestamp: now - createdAt is negative, which is <= ttlMs (5000)
		out, hitFuture := GetCacheEntry(tempDir, "cmd_ts_future", 5)
		if !hitFuture || out != "out_future" {
			t.Errorf("Expected cache hit for future timestamp, got hit=%v, out=%q", hitFuture, out)
		}
	})

	t.Run("Special Characters in CWD and Command", func(t *testing.T) {
		specialDir := filepath.Join(tempDir, "テスト ディレクトリ containing spaces & symbols $#@!")
		_ = os.MkdirAll(filepath.Join(specialDir, ".git"), 0755)
		_ = os.WriteFile(filepath.Join(specialDir, ".git", "HEAD"), []byte("ref: refs/heads/feature/special-1"), 0644)

		specialCmd := "log --graph --pretty=format:'%h - %s (%cr) <%an>' -n 5"

		ctx := RenderContextDummy{CWD: specialDir, GitCacheTTLSeconds: 5}
		mockExec := func(args []string, cwd string) (string, error) {
			return "commit_log_output", nil
		}

		out, err := RunGit(specialCmd, ctx, 5, mockExec)
		if err != nil || out != "commit_log_output" {
			t.Fatalf("RunGit failed on special chars: out=%q, err=%v", out, err)
		}

		// Subsequent call should hit cache
		outCached, err := RunGit(specialCmd, ctx, 5, func(args []string, cwd string) (string, error) {
			t.Fatalf("mockExec should not be called on cache hit")
			return "", nil
		})
		if err != nil || outCached != "commit_log_output" {
			t.Errorf("RunGit cache hit failed on special chars: out=%q, err=%v", outCached, err)
		}
	})

	t.Run("Cache directory is a regular file (Permission/Mkdir error)", func(t *testing.T) {
		blockDir := filepath.Join(tempDir, "file_as_cache_dir")
		_ = os.WriteFile(blockDir, []byte("I am a file, not a directory"), 0644)

		SetCacheDir(blockDir)
		ClearGitCache()

		// SetCacheEntry should fail to write persistent cache but still update memory cache cleanly without crashing
		now := time.Now().UnixNano() / int64(time.Millisecond)
		SetCacheEntry(tempDir, "status", "mem_only", now, 5)

		out, hit := GetCacheEntry(tempDir, "status", 5)
		if !hit || out != "mem_only" {
			t.Errorf("Memory cache failed when persistent cache dir is a file: hit=%v, out=%q", hit, out)
		}
	})
}

func TestGitCache_FuzzGenerator(t *testing.T) {
	tempDir := t.TempDir()
	SetCacheDir(tempDir)
	ClearGitCache()

	seed := time.Now().UnixNano()
	t.Logf("Fuzz test seed: %d", seed)

	repoDir := filepath.Join(tempDir, "fuzz_repo")
	gitDir := filepath.Join(repoDir, ".git")
	_ = os.MkdirAll(gitDir, 0755)
	_ = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main"), 0644)
	_ = os.WriteFile(filepath.Join(gitDir, "index"), []byte("index_fuzz"), 0644)

	ctx := RenderContextDummy{
		CWD:                repoDir,
		GitCacheTTLSeconds: 5,
	}

	randomBytes := func(n int) string {
		b := make([]byte, n)
		_, _ = rand.Read(b)
		return hex.EncodeToString(b)
	}

	for i := range 100 {
		cmd := fmt.Sprintf("cmd_%d", i%10)
		val := randomBytes(16)
		ttl := (i % 7) - 2 // includes negative, zero, positive TTLs

		mockExec := func(args []string, cwd string) (string, error) {
			return val, nil
		}

		switch i % 4 {
		case 0:
			out, err := RunGit(cmd, ctx, ttl, mockExec)
			if err != nil {
				t.Errorf("Fuzz RunGit op %d failed with error: %v", i, err)
			}
			if ttl > 0 && out == "" {
				t.Errorf("Fuzz RunGit op %d expected non-empty output", i)
			}
		case 1:
			now := time.Now().UnixNano() / int64(time.Millisecond)
			SetCacheEntry(repoDir, cmd, val, now, ttl)
		case 2:
			cachedVal, hit := GetCacheEntry(repoDir, cmd, ttl)
			if hit && ttl < 0 {
				t.Errorf("Fuzz GetCacheEntry op %d: hit=true for negative TTL %d", i, ttl)
			}
			if hit && cachedVal == "" {
				t.Errorf("Fuzz GetCacheEntry op %d: hit=true but empty cached value", i)
			}
		case 3:
			if i%20 == 0 {
				ClearGitCache()
			}
		}
	}
}

func TestIsCacheEntryFresh_TTL(t *testing.T) {
	now := int64(100000)
	tests := []struct {
		name      string
		createdAt int64
		ttlMs     int64
		wantFresh bool
	}{
		{
			name:      "positive TTL within duration is fresh",
			createdAt: 98000, // age 2000ms
			ttlMs:     5000,
			wantFresh: true,
		},
		{
			name:      "positive TTL exactly at boundary is fresh",
			createdAt: 95000, // age 5000ms
			ttlMs:     5000,
			wantFresh: true,
		},
		{
			name:      "positive TTL expired is not fresh",
			createdAt: 94000, // age 6000ms
			ttlMs:     5000,
			wantFresh: false,
		},
		{
			name:      "zero TTL is never fresh (caching disabled)",
			createdAt: now,
			ttlMs:     0,
			wantFresh: false,
		},
		{
			name:      "negative TTL is never fresh",
			createdAt: now,
			ttlMs:     -1000,
			wantFresh: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := GitCacheEntryInternal{
				CreatedAt: tt.createdAt,
			}
			got := isCacheEntryFresh(entry, nil, tt.ttlMs, now)
			if got != tt.wantFresh {
				t.Errorf("isCacheEntryFresh(createdAt=%d, ttlMs=%d, now=%d) = %v, want %v",
					tt.createdAt, tt.ttlMs, now, got, tt.wantFresh)
			}
		})
	}
}

func TestGetCacheEntry_TTLZeroAndNegative(t *testing.T) {
	tempDir := t.TempDir()
	SetCacheDir(tempDir)
	ClearGitCache()

	now := time.Now().UnixNano() / int64(time.Millisecond)
	SetCacheEntry(tempDir, "status", "clean", now, 5)

	tests := []struct {
		name       string
		ttlSeconds int
		wantHit    bool
	}{
		{
			name:       "ttl 0 disables cache hit",
			ttlSeconds: 0,
			wantHit:    false,
		},
		{
			name:       "negative ttl disables cache hit",
			ttlSeconds: -1,
			wantHit:    false,
		},
		{
			name:       "positive ttl returns cache hit",
			ttlSeconds: 5,
			wantHit:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, hit := GetCacheEntry(tempDir, "status", tt.ttlSeconds)
			if hit != tt.wantHit {
				t.Errorf("GetCacheEntry(ttlSeconds=%d) hit = %v, want %v", tt.ttlSeconds, hit, tt.wantHit)
			}
		})
	}
}

func TestRunGit_TTLZero_AlwaysExecutes(t *testing.T) {
	tempDir := t.TempDir()
	SetCacheDir(tempDir)
	ClearGitCache()

	ctx := RenderContextDummy{
		CWD:                tempDir,
		GitCacheTTLSeconds: 0,
	}

	callCount := 0
	mockExec := func(args []string, cwd string) (string, error) {
		callCount++
		return fmt.Sprintf("output-%d", callCount), nil
	}

	out1, err := RunGit("status", ctx, 0, mockExec)
	if err != nil {
		t.Fatalf("RunGit call 1 failed: %v", err)
	}
	if out1 != "output-1" {
		t.Errorf("RunGit call 1 = %q, want %q", out1, "output-1")
	}

	out2, err := RunGit("status", ctx, 0, mockExec)
	if err != nil {
		t.Fatalf("RunGit call 2 failed: %v", err)
	}
	if out2 != "output-2" {
		t.Errorf("RunGit call 2 = %q, want %q (cache was erroneously hit with ttl=0)", out2, "output-2")
	}

	if callCount != 2 {
		t.Errorf("mockExec called %d times, want 2 (ttl=0 should never cache)", callCount)
	}
}

func TestGetHeadMtimeMs(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, dir string) (gitDir string, wantMtime int64)
	}{
		{
			name: "non-existent gitDir returns 0",
			setup: func(t *testing.T, dir string) (string, int64) {
				return filepath.Join(dir, "nonexistent"), 0
			},
		},
		{
			name: "detached HEAD returns HEAD mtime",
			setup: func(t *testing.T, dir string) (string, int64) {
				gitDir := filepath.Join(dir, ".git")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					t.Fatalf("failed to create gitDir: %v", err)
				}
				headPath := filepath.Join(gitDir, "HEAD")
				if err := os.WriteFile(headPath, []byte("0123456789abcdef\n"), 0644); err != nil {
					t.Fatalf("failed to write HEAD: %v", err)
				}
				mtime := time.Now().Add(-10 * time.Minute).Truncate(time.Second)
				if err := os.Chtimes(headPath, mtime, mtime); err != nil {
					t.Fatalf("failed to chtimes HEAD: %v", err)
				}
				return gitDir, mtime.UnixNano() / int64(time.Millisecond)
			},
		},
		{
			name: "symbolic ref with newer branch ref file returns branch ref mtime",
			setup: func(t *testing.T, dir string) (string, int64) {
				gitDir := filepath.Join(dir, ".git")
				headsDir := filepath.Join(gitDir, "refs", "heads")
				if err := os.MkdirAll(headsDir, 0755); err != nil {
					t.Fatalf("failed to create headsDir: %v", err)
				}
				headPath := filepath.Join(gitDir, "HEAD")
				if err := os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644); err != nil {
					t.Fatalf("failed to write HEAD: %v", err)
				}
				oldTime := time.Now().Add(-30 * time.Minute).Truncate(time.Second)
				if err := os.Chtimes(headPath, oldTime, oldTime); err != nil {
					t.Fatalf("failed to chtimes HEAD: %v", err)
				}
				branchPath := filepath.Join(headsDir, "main")
				if err := os.WriteFile(branchPath, []byte("1111222233334444\n"), 0644); err != nil {
					t.Fatalf("failed to write branch ref: %v", err)
				}
				newTime := time.Now().Add(-5 * time.Minute).Truncate(time.Second)
				if err := os.Chtimes(branchPath, newTime, newTime); err != nil {
					t.Fatalf("failed to chtimes branch ref: %v", err)
				}
				return gitDir, newTime.UnixNano() / int64(time.Millisecond)
			},
		},
		{
			name: "symbolic ref where HEAD is newer returns HEAD mtime",
			setup: func(t *testing.T, dir string) (string, int64) {
				gitDir := filepath.Join(dir, ".git")
				headsDir := filepath.Join(gitDir, "refs", "heads")
				if err := os.MkdirAll(headsDir, 0755); err != nil {
					t.Fatalf("failed to create headsDir: %v", err)
				}
				branchPath := filepath.Join(headsDir, "main")
				if err := os.WriteFile(branchPath, []byte("1111222233334444\n"), 0644); err != nil {
					t.Fatalf("failed to write branch ref: %v", err)
				}
				oldTime := time.Now().Add(-30 * time.Minute).Truncate(time.Second)
				if err := os.Chtimes(branchPath, oldTime, oldTime); err != nil {
					t.Fatalf("failed to chtimes branch ref: %v", err)
				}
				headPath := filepath.Join(gitDir, "HEAD")
				if err := os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644); err != nil {
					t.Fatalf("failed to write HEAD: %v", err)
				}
				newTime := time.Now().Add(-5 * time.Minute).Truncate(time.Second)
				if err := os.Chtimes(headPath, newTime, newTime); err != nil {
					t.Fatalf("failed to chtimes HEAD: %v", err)
				}
				return gitDir, newTime.UnixNano() / int64(time.Millisecond)
			},
		},
		{
			name: "packed-refs with newer mtime returns packed-refs mtime",
			setup: func(t *testing.T, dir string) (string, int64) {
				gitDir := filepath.Join(dir, ".git")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					t.Fatalf("failed to create gitDir: %v", err)
				}
				headPath := filepath.Join(gitDir, "HEAD")
				if err := os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644); err != nil {
					t.Fatalf("failed to write HEAD: %v", err)
				}
				oldTime := time.Now().Add(-20 * time.Minute).Truncate(time.Second)
				if err := os.Chtimes(headPath, oldTime, oldTime); err != nil {
					t.Fatalf("failed to chtimes HEAD: %v", err)
				}
				packedPath := filepath.Join(gitDir, "packed-refs")
				if err := os.WriteFile(packedPath, []byte("# pack-refs\n"), 0644); err != nil {
					t.Fatalf("failed to write packed-refs: %v", err)
				}
				newTime := time.Now().Add(-2 * time.Minute).Truncate(time.Second)
				if err := os.Chtimes(packedPath, newTime, newTime); err != nil {
					t.Fatalf("failed to chtimes packed-refs: %v", err)
				}
				return gitDir, newTime.UnixNano() / int64(time.Millisecond)
			},
		},
		{
			name: "linked worktree with commondir resolves common refs and packed-refs",
			setup: func(t *testing.T, dir string) (string, int64) {
				mainGitDir := filepath.Join(dir, "main", ".git")
				mainHeadsDir := filepath.Join(mainGitDir, "refs", "heads")
				if err := os.MkdirAll(mainHeadsDir, 0755); err != nil {
					t.Fatalf("failed to create mainHeadsDir: %v", err)
				}
				wtGitDir := filepath.Join(mainGitDir, "worktrees", "wt1")
				if err := os.MkdirAll(wtGitDir, 0755); err != nil {
					t.Fatalf("failed to create wtGitDir: %v", err)
				}
				commondirFile := filepath.Join(wtGitDir, "commondir")
				if err := os.WriteFile(commondirFile, []byte("../..\n"), 0644); err != nil {
					t.Fatalf("failed to write commondir: %v", err)
				}
				wtHeadPath := filepath.Join(wtGitDir, "HEAD")
				if err := os.WriteFile(wtHeadPath, []byte("ref: refs/heads/feature-wt\n"), 0644); err != nil {
					t.Fatalf("failed to write wt HEAD: %v", err)
				}
				oldTime := time.Now().Add(-25 * time.Minute).Truncate(time.Second)
				if err := os.Chtimes(wtHeadPath, oldTime, oldTime); err != nil {
					t.Fatalf("failed to chtimes wt HEAD: %v", err)
				}
				branchPath := filepath.Join(mainHeadsDir, "feature-wt")
				if err := os.WriteFile(branchPath, []byte("wt_sha_123456\n"), 0644); err != nil {
					t.Fatalf("failed to write wt branch ref: %v", err)
				}
				newTime := time.Now().Add(-3 * time.Minute).Truncate(time.Second)
				if err := os.Chtimes(branchPath, newTime, newTime); err != nil {
					t.Fatalf("failed to chtimes wt branch ref: %v", err)
				}
				return wtGitDir, newTime.UnixNano() / int64(time.Millisecond)
			},
		},
		{
			name: "malicious symbolic ref with directory traversal is ignored",
			setup: func(t *testing.T, dir string) (string, int64) {
				gitDir := filepath.Join(dir, ".git")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					t.Fatalf("failed to create gitDir: %v", err)
				}
				headPath := filepath.Join(gitDir, "HEAD")
				if err := os.WriteFile(headPath, []byte("ref: ../../../etc/passwd\n"), 0644); err != nil {
					t.Fatalf("failed to write HEAD: %v", err)
				}
				headTime := time.Now().Add(-15 * time.Minute).Truncate(time.Second)
				if err := os.Chtimes(headPath, headTime, headTime); err != nil {
					t.Fatalf("failed to chtimes HEAD: %v", err)
				}
				return gitDir, headTime.UnixNano() / int64(time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			gitDir, wantMtime := tt.setup(t, tempDir)
			got := getHeadMtimeMs(gitDir)
			if got != wantMtime {
				t.Errorf("getHeadMtimeMs(%q) = %d, want %d", gitDir, got, wantMtime)
			}
		})
	}
}

func TestGitCache_InvalidationOnRefUpdates(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(t *testing.T, gitDir string) (pathToTouch string)
	}{
		{
			name: "invalidation on branch ref update",
			setupRepo: func(t *testing.T, gitDir string) string {
				headsDir := filepath.Join(gitDir, "refs", "heads")
				if err := os.MkdirAll(headsDir, 0755); err != nil {
					t.Fatalf("failed to create headsDir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/feature\n"), 0644); err != nil {
					t.Fatalf("failed to write HEAD: %v", err)
				}
				branchPath := filepath.Join(headsDir, "feature")
				if err := os.WriteFile(branchPath, []byte("sha1\n"), 0644); err != nil {
					t.Fatalf("failed to write branch ref: %v", err)
				}
				return branchPath
			},
		},
		{
			name: "invalidation on packed-refs update",
			setupRepo: func(t *testing.T, gitDir string) string {
				if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0644); err != nil {
					t.Fatalf("failed to write HEAD: %v", err)
				}
				packedPath := filepath.Join(gitDir, "packed-refs")
				if err := os.WriteFile(packedPath, []byte("# pack-refs\n"), 0644); err != nil {
					t.Fatalf("failed to write packed-refs: %v", err)
				}
				return packedPath
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			SetCacheDir(tempDir)
			ClearGitCache()

			gitDir := filepath.Join(tempDir, ".git")
			if err := os.MkdirAll(gitDir, 0755); err != nil {
				t.Fatalf("failed to create gitDir: %v", err)
			}
			if err := os.WriteFile(filepath.Join(gitDir, "index"), []byte("index_1\n"), 0644); err != nil {
				t.Fatalf("failed to write index: %v", err)
			}

			pathToTouch := tt.setupRepo(t, gitDir)

			now := time.Now().UnixNano() / int64(time.Millisecond)
			SetCacheEntry(tempDir, "status", "cached_status", now, 10)

			// Verify initial cache hit
			out, hit := GetCacheEntry(tempDir, "status", 10)
			if !hit || out != "cached_status" {
				t.Fatalf("Initial GetCacheEntry failed: hit=%v, out=%q", hit, out)
			}

			// Simulate file update
			newTime := time.Now().Add(2 * time.Hour)
			if err := os.Chtimes(pathToTouch, newTime, newTime); err != nil {
				t.Fatalf("failed to chtimes %s: %v", pathToTouch, err)
			}

			// Cache must be invalidated
			_, hitAfterUpdate := GetCacheEntry(tempDir, "status", 10)
			if hitAfterUpdate {
				t.Errorf("Expected cache miss after touching %s, but cache hit", pathToTouch)
			}
		})
	}
}

func TestSetCacheEntry_TTLZeroAndNegative_NoDiskWrite(t *testing.T) {
	tests := []struct {
		name       string
		ttlSeconds int
	}{
		{
			name:       "ttl 0 does not write disk cache or memory cache",
			ttlSeconds: 0,
		},
		{
			name:       "negative ttl does not write disk cache or memory cache",
			ttlSeconds: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			cacheRootDir := filepath.Join(tempDir, "cache_root")
			SetCacheDir(cacheRootDir)
			ClearGitCache()

			gitDir := filepath.Join(tempDir, ".git")
			if err := os.MkdirAll(gitDir, 0755); err != nil {
				t.Fatalf("failed to create gitDir: %v", err)
			}
			if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0644); err != nil {
				t.Fatalf("failed to write HEAD: %v", err)
			}

			now := time.Now().UnixNano() / int64(time.Millisecond)
			SetCacheEntry(tempDir, "status", "clean", now, tt.ttlSeconds)

			// 1. Verify entry was not written to cache by querying with a positive TTL (60s).
			// If SetCacheEntry had erroneously written to memory or disk, a positive TTL query would hit.
			_, hit := GetCacheEntry(tempDir, "status", 60)
			if hit {
				t.Errorf("GetCacheEntry with positive TTL returned hit=true after SetCacheEntry with ttlSeconds=%d", tt.ttlSeconds)
			}

			// 2. Directly verify in-memory cache map does not contain the key
			gitCacheMutex.Lock()
			memKey := fmt.Sprintf("status|%s", tempDir)
			_, inMem := gitCommandCache[memKey]
			gitCacheMutex.Unlock()
			if inMem {
				t.Errorf("In-memory gitCommandCache contains entry for key %q after SetCacheEntry with ttlSeconds=%d", memKey, tt.ttlSeconds)
			}

			// 3. Verify no cache file was created on disk
			meta, err := getGitRepoMetadata(tempDir)
			if err != nil {
				t.Fatalf("getGitRepoMetadata failed: %v", err)
			}
			if _, statErr := os.Stat(meta.CachePath); !os.IsNotExist(statErr) {
				t.Errorf("Cache file %s exists on disk, but should not have been written for ttlSeconds=%d", meta.CachePath, tt.ttlSeconds)
			}
		})
	}
}
