package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type CwdResolver interface {
	GetCwd() string
	GetWorkspaceCurrentDir() string
	GetWorkspaceProjectDir() string
}

type GitRepoMetadata struct {
	CachePath    string
	HeadMtimeMS  int64
	IndexMtimeMS int64
}

type GitCacheEntryInternal struct {
	Output       *string `json:"output"`
	CreatedAt    int64   `json:"createdAt"`
	HeadMtimeMS  int64   `json:"headMtimeMs"`
	IndexMtimeMS int64   `json:"indexMtimeMs"`
}

type PersistentGitCacheInternal struct {
	Version int                              `json:"version"`
	CWD     string                           `json:"cwd"`
	Entries map[string]GitCacheEntryInternal `json:"entries"`
}

const GitCacheSchemaVersion = 1

var (
	cacheDir        = ""
	cacheDirMutex   sync.RWMutex
	gitCommandCache = make(map[string]GitCacheEntryInternal)
	gitCacheMutex   sync.Mutex
)

// SetCacheDir sets the root cache directory (primarily for testing).
func SetCacheDir(path string) {
	cacheDirMutex.Lock()
	defer cacheDirMutex.Unlock()
	cacheDir = path
}

func getCacheDir() string {
	cacheDirMutex.RLock()
	defer cacheDirMutex.RUnlock()
	if cacheDir != "" {
		return cacheDir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "agystatusline")
}

func getRepoCachePath(gitDir string) string {
	hash := sha256.New()
	hash.Write([]byte(gitDir))
	repoHash := hex.EncodeToString(hash.Sum(nil))[:16]
	return filepath.Join(getCacheDir(), "git-cache", fmt.Sprintf("git-%s.json", repoHash))
}

func getMtimeMs(filePath string) int64 {
	stat, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return stat.ModTime().UnixNano() / int64(time.Millisecond)
}

func normalizeDirectory(candidate string) (string, error) {
	resolved, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	stat, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if stat.IsDir() {
		return resolved, nil
	}
	return filepath.Dir(resolved), nil
}

func readGitDirFile(gitFilePath string) (string, error) {
	contentBytes, err := os.ReadFile(gitFilePath)
	if err != nil {
		return "", err
	}
	content := strings.TrimSpace(string(contentBytes))
	if strings.HasPrefix(strings.ToLower(content), "gitdir:") {
		target := strings.TrimSpace(content[7:])
		if filepath.IsAbs(target) {
			return target, nil
		}
		return filepath.Abs(filepath.Join(filepath.Dir(gitFilePath), target))
	}
	return "", errors.New("not a gitdir file")
}

func discoverGitDir(startDir string) (string, error) {
	current := startDir
	for {
		gitPath := filepath.Join(current, ".git")
		stat, err := os.Stat(gitPath)
		if err == nil {
			if stat.IsDir() {
				return gitPath, nil
			}
			// check if it's a file (worktree link)
			target, err := readGitDirFile(gitPath)
			if err == nil {
				return target, nil
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", errors.New("git directory not found")
}

func getGitRepoMetadata(cwd string) (*GitRepoMetadata, error) {
	if cwd == "" {
		return nil, errors.New("empty cwd")
	}

	startDir, err := normalizeDirectory(cwd)
	if err != nil {
		return nil, err
	}

	gitDir, err := discoverGitDir(startDir)
	if err != nil {
		return nil, err
	}

	return &GitRepoMetadata{
		CachePath:    getRepoCachePath(gitDir),
		HeadMtimeMS:  getMtimeMs(filepath.Join(gitDir, "HEAD")),
		IndexMtimeMS: getMtimeMs(filepath.Join(gitDir, "index")),
	}, nil
}

func readPersistentCache(cachePath string) (*PersistentGitCacheInternal, error) {
	file, err := os.Open(cachePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cache PersistentGitCacheInternal
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cache); err != nil {
		return nil, err
	}

	if cache.Version != GitCacheSchemaVersion {
		return nil, errors.New("invalid cache version")
	}

	return &cache, nil
}

func writePersistentCache(cachePath string, cache *PersistentGitCacheInternal) {
	cacheDir := filepath.Dir(cachePath)
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return
	}

	tempPath := fmt.Sprintf("%s.%d.%d.tmp", cachePath, os.Getpid(), time.Now().UnixNano())
	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer func() {
		file.Close()
		os.Remove(tempPath) // clean up temp file if rename failed
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cache); err != nil {
		return
	}
	file.Close()

	os.Rename(tempPath, cachePath)
}

func isCacheEntryFresh(entry GitCacheEntryInternal, metadata *GitRepoMetadata, ttlMs int64, now int64) bool {
	if metadata != nil {
		if entry.HeadMtimeMS != metadata.HeadMtimeMS || entry.IndexMtimeMS != metadata.IndexMtimeMS {
			return false
		}
	}
	return ttlMs == 0 || now-entry.CreatedAt <= ttlMs
}

// ResolveGitCwd resolves the CWD from candidate paths inside the context.
func ResolveGitCwd(ctx CwdResolver) string {
	candidates := []string{
		ctx.GetCwd(),
		ctx.GetWorkspaceCurrentDir(),
		ctx.GetWorkspaceProjectDir(),
	}
	for _, c := range candidates {
		if strings.TrimSpace(c) != "" {
			return c
		}
	}
	return ""
}

// GetCacheEntry checks the cache (memory, then file) and returns the output if fresh.
func GetCacheEntry(cwd string, command string, ttlSeconds int) (string, bool) {
	gitCacheMutex.Lock()
	defer gitCacheMutex.Unlock()

	metadata, err := getGitRepoMetadata(cwd)
	now := time.Now().UnixNano() / int64(time.Millisecond)
	ttlMs := int64(ttlSeconds) * 1000

	memoryKey := fmt.Sprintf("%s|%s", command, cwd)
	if entry, exists := gitCommandCache[memoryKey]; exists {
		if isCacheEntryFresh(entry, metadata, ttlMs, now) {
			if entry.Output == nil {
				return "", true
			}
			return *entry.Output, true
		}
	}

	if err == nil && metadata != nil {
		pCache, err := readPersistentCache(metadata.CachePath)
		if err == nil && pCache != nil && pCache.CWD == cwd {
			if entry, exists := pCache.Entries[command]; exists {
				if isCacheEntryFresh(entry, metadata, ttlMs, now) {
					gitCommandCache[memoryKey] = entry // populate memory cache
					if entry.Output == nil {
						return "", true
					}
					return *entry.Output, true
				}
			}
		}
	}

	return "", false
}

// SetCacheEntry stores a cache entry both in memory and persistently in file cache.
func SetCacheEntry(cwd string, command string, output string, timestampMS int64, ttlSeconds int) {
	gitCacheMutex.Lock()
	defer gitCacheMutex.Unlock()

	metadata, err := getGitRepoMetadata(cwd)
	var headMtime, indexMtime int64
	if err == nil && metadata != nil {
		headMtime = metadata.HeadMtimeMS
		indexMtime = metadata.IndexMtimeMS
	}

	var outputPtr *string
	if output != "" {
		outVal := output
		outputPtr = &outVal
	}

	entry := GitCacheEntryInternal{
		Output:       outputPtr,
		CreatedAt:    timestampMS,
		HeadMtimeMS:  headMtime,
		IndexMtimeMS: indexMtime,
	}

	memoryKey := fmt.Sprintf("%s|%s", command, cwd)
	gitCommandCache[memoryKey] = entry

	if err == nil && metadata != nil {
		pCache, _ := readPersistentCache(metadata.CachePath)
		if pCache == nil || pCache.CWD != cwd {
			pCache = &PersistentGitCacheInternal{
				Version: GitCacheSchemaVersion,
				CWD:     cwd,
				Entries: make(map[string]GitCacheEntryInternal),
			}
		}
		pCache.Entries[command] = entry
		writePersistentCache(metadata.CachePath, pCache)
	}
}

// ClearGitCache clears in-memory caches.
func ClearGitCache() {
	gitCacheMutex.Lock()
	defer gitCacheMutex.Unlock()
	gitCommandCache = make(map[string]GitCacheEntryInternal)
}

// RunGit runs a git command and caches the result.
func RunGit(command string, ctx CwdResolver, ttlSeconds int, runCommandFunc func(args []string, cwd string) (string, error)) (string, error) {
	cwd := ResolveGitCwd(ctx)
	if cwd == "" {
		return "", errors.New("cannot resolve CWD")
	}

	if output, hit := GetCacheEntry(cwd, command, ttlSeconds); hit {
		return output, nil
	}

	args := strings.Fields(command)
	if len(args) == 0 {
		return "", errors.New("empty command")
	}

	output, err := runCommandFunc(args, cwd)
	now := time.Now().UnixNano() / int64(time.Millisecond)
	if err != nil {
		SetCacheEntry(cwd, command, "", now, ttlSeconds)
		return "", err
	}

	trimmedOutput := strings.TrimSpace(output)
	SetCacheEntry(cwd, command, trimmedOutput, now, ttlSeconds)
	return trimmedOutput, nil
}
