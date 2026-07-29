package utils

import (
	"os"
	"path/filepath"
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
	tests := []struct {
		name        string
		setup       func(dir string) string
		wantErr     bool
		expectedErr string
	}{
		{
			name: "Non-existent file",
			setup: func(dir string) string {
				return filepath.Join(dir, "nonexistent.json")
			},
			wantErr: true,
		},
		{
			name: "Invalid JSON file",
			setup: func(dir string) string {
				invalidJsonPath := filepath.Join(dir, "invalid.json")
				_ = os.WriteFile(invalidJsonPath, []byte("invalid json{"), 0644)
				return invalidJsonPath
			},
			wantErr: true,
		},
		{
			name: "Invalid schema version",
			setup: func(dir string) string {
				invalidVerPath := filepath.Join(dir, "invalid_ver.json")
				_ = os.WriteFile(invalidVerPath, []byte(`{"version": 999}`), 0644)
				return invalidVerPath
			},
			wantErr:     true,
			expectedErr: "invalid cache version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			path := tt.setup(tempDir)
			_, err := readPersistentCache(path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("readPersistentCache() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.expectedErr != "" && (err == nil || err.Error() != tt.expectedErr) {
				t.Errorf("readPersistentCache() error = %v, want %q", err, tt.expectedErr)
			}
		})
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
			t.Errorf("Expected no temporary files remaining in %q, but found %q", tempDir, f.Name())
		}
	}
}

func TestReadGitDirFile(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(dir string) (filePath string, want string)
		wantErr     bool
		expectedErr string
	}{
		{
			name: "Non-existent file",
			setup: func(dir string) (string, string) {
				return filepath.Join(dir, "nonexistent"), ""
			},
			wantErr: true,
		},
		{
			name: "Invalid content without gitdir prefix",
			setup: func(dir string) (string, string) {
				invalidFile := filepath.Join(dir, "invalid_git")
				_ = os.WriteFile(invalidFile, []byte("invalid content"), 0644)
				return invalidFile, ""
			},
			wantErr:     true,
			expectedErr: "not a gitdir file",
		},
		{
			name: "Valid relative gitdir file",
			setup: func(dir string) (string, string) {
				relFile := filepath.Join(dir, "rel_git")
				_ = os.WriteFile(relFile, []byte("gitdir: ../target_git"), 0644)
				expectedTarget, _ := filepath.Abs(filepath.Join(dir, "../target_git"))
				return relFile, expectedTarget
			},
			wantErr: false,
		},
		{
			name: "Valid absolute gitdir file",
			setup: func(dir string) (string, string) {
				absTarget := filepath.Join(dir, "abs_target")
				absFile := filepath.Join(dir, "abs_git")
				_ = os.WriteFile(absFile, []byte("gitdir: "+absTarget), 0644)
				return absFile, absTarget
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			path, want := tt.setup(tempDir)
			got, err := readGitDirFile(path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("readGitDirFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.expectedErr != "" && (err == nil || err.Error() != tt.expectedErr) {
				t.Errorf("readGitDirFile() error = %v, want %q", err, tt.expectedErr)
			}
			if !tt.wantErr && got != want {
				t.Errorf("readGitDirFile() = %q, want %q", got, want)
			}
		})
	}
}

func TestDiscoverGitDir(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(dir string) (searchDir string, want string)
		wantErr bool
	}{
		{
			name: "No git dir found",
			setup: func(dir string) (string, string) {
				return dir, ""
			},
			wantErr: true,
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
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			searchDir, want := tt.setup(tempDir)
			got, err := discoverGitDir(searchDir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("discoverGitDir() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != want {
				t.Errorf("discoverGitDir() = %q, want %q", got, want)
			}
		})
	}
}

func TestNormalizeDirectory(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(dir string) (inputPath string, want string)
		wantErr bool
	}{
		{
			name: "Non-existent path",
			setup: func(dir string) (string, string) {
				return filepath.Join(dir, "nonexistent"), ""
			},
			wantErr: true,
		},
		{
			name: "Directory path",
			setup: func(dir string) (string, string) {
				return dir, dir
			},
			wantErr: false,
		},
		{
			name: "File path returns parent directory",
			setup: func(dir string) (string, string) {
				filePath := filepath.Join(dir, "file.txt")
				_ = os.WriteFile(filePath, []byte("test"), 0644)
				return filePath, dir
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			input, want := tt.setup(tempDir)
			got, err := normalizeDirectory(input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != want {
				t.Errorf("normalizeDirectory() = %q, want %q", got, want)
			}
		})
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
	type gitExecFunc func(args []string, cwd string) (string, error)

	tests := []struct {
		name    string
		command string
		getCtx  func(dir string) CwdResolver
		want    string
		wantErr bool
		verify  func(t *testing.T, ctx CwdResolver, mockExec gitExecFunc, execCount *int)
	}{
		{
			name:    "Empty CWD",
			command: "status",
			getCtx: func(dir string) CwdResolver {
				return RenderContextDummy{CWD: ""}
			},
			wantErr: true,
		},
		{
			name:    "Empty command",
			command: "",
			getCtx: func(dir string) CwdResolver {
				return RenderContextDummy{CWD: dir, GitCacheTTLSeconds: 5}
			},
			wantErr: true,
		},
		{
			name:    "Successful execution and caching",
			command: "symbolic-ref --short HEAD",
			getCtx: func(dir string) CwdResolver {
				return RenderContextDummy{CWD: dir, GitCacheTTLSeconds: 5}
			},
			want: "branch-name",
			verify: func(t *testing.T, ctx CwdResolver, mockExec gitExecFunc, execCount *int) {
				// Second call should hit cache and not invoke mockExec again
				out2, err := RunGit("symbolic-ref --short HEAD", ctx, 5, mockExec)
				if err != nil || out2 != "branch-name" {
					t.Errorf("Expected 'branch-name', got %q, err=%v", out2, err)
				}
				if *execCount != 1 {
					t.Errorf("Expected mockExec execution count to remain 1 (cache hit), got %d", *execCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			SetCacheDir(tempDir)
			ClearGitCache()

			ctx := tt.getCtx(tempDir)
			execCount := 0
			var execFunc gitExecFunc
			if tt.want != "" || tt.verify != nil {
				execFunc = func(args []string, cwd string) (string, error) {
					execCount++
					return tt.want, nil
				}
			}

			got, err := RunGit(tt.command, ctx, 5, execFunc)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RunGit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("RunGit() = %q, want %q", got, tt.want)
			}
			if tt.want != "" && execCount != 1 {
				t.Errorf("Expected mockExec to be called once, got %d", execCount)
			}
			if tt.verify != nil {
				tt.verify(t, ctx, execFunc, &execCount)
			}
		})
	}
}
