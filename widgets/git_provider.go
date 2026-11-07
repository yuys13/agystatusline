package widgets

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/yuys13/agystatusline/types"
)

var (
	insertShortstatRe = regexp.MustCompile(`(\d+)\s+insertions?`)
	deleteShortstatRe = regexp.MustCompile(`(\d+)\s+deletions?`)
)

func parseShortStat(stat string) (int, int) {
	insertMatch := insertShortstatRe.FindStringSubmatch(stat)
	deleteMatch := deleteShortstatRe.FindStringSubmatch(stat)

	ins := 0
	del := 0

	if len(insertMatch) > 1 {
		ins, _ = strconv.Atoi(insertMatch[1])
	}
	if len(deleteMatch) > 1 {
		del, _ = strconv.Atoi(deleteMatch[1])
	}

	return ins, del
}

func parseAheadBehind(out string) (int, int, error) {
	fields := strings.Fields(out)
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list count output: %q", out)
	}
	ahead, err1 := strconv.Atoi(fields[0])
	behind, err2 := strconv.Atoi(fields[1])
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("invalid rev-list count: %q", out)
	}
	return ahead, behind, nil
}

// staticCwd adapts a working directory string to utils.CwdResolver.
type staticCwd string

func (s staticCwd) GetCwd() string                 { return string(s) }
func (s staticCwd) GetWorkspaceCurrentDir() string { return string(s) }
func (s staticCwd) GetWorkspaceProjectDir() string { return string(s) }

// GitProvider implements types.GitDataProvider using git CLI commands with caching and intra-pass memoization.
type GitProvider struct {
	mu         sync.Mutex
	cwd        string
	ttlSeconds int

	isGitChecked bool
	isGitErr     error

	branchChecked bool
	branch        string
	branchDirty   bool
	branchErr     error

	changesChecked bool
	changes        types.GitChanges
	changesErr     error

	aheadBehindChecked bool
	aheadBehind        types.GitAheadBehind
	aheadBehindErr     error

	repoNameChecked bool
	repoName        string
	repoNameErr     error
}

// NewGitProvider creates a new GitProvider bound to the given directory and cache TTL.
func NewGitProvider(cwd string, ttlSeconds int) *GitProvider {
	return &GitProvider{
		cwd:        cwd,
		ttlSeconds: ttlSeconds,
	}
}

func (p *GitProvider) checkInsideWorkTree() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isGitChecked {
		return p.isGitErr
	}
	p.isGitChecked = true

	isGit, err := runGitCommand("rev-parse --is-inside-work-tree", staticCwd(p.cwd), p.ttlSeconds)
	if err != nil || strings.TrimSpace(isGit) != "true" {
		p.isGitErr = types.ErrNotGitRepository
		return p.isGitErr
	}
	return nil
}

// GetBranch retrieves current branch name and whether the working tree is dirty.
func (p *GitProvider) GetBranch() (string, bool, error) {
	if err := p.checkInsideWorkTree(); err != nil {
		return "", false, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.branchChecked {
		return p.branch, p.branchDirty, p.branchErr
	}
	p.branchChecked = true

	b, err := runGitCommand("symbolic-ref --short HEAD", staticCwd(p.cwd), p.ttlSeconds)
	if err == nil && strings.TrimSpace(b) != "" {
		p.branch = strings.TrimSpace(b)
	} else {
		// Fallback for detached HEAD state:
		// 1. Prioritize exact tag name if available
		tag, tagErr := runGitCommand("describe --tags --exact-match", staticCwd(p.cwd), p.ttlSeconds)
		if tagErr == nil && strings.TrimSpace(tag) != "" {
			p.branch = "(" + strings.TrimSpace(tag) + ")"
		} else {
			// 2. Fall back to short commit SHA
			sha, shaErr := runGitCommand("rev-parse --short HEAD", staticCwd(p.cwd), p.ttlSeconds)
			if shaErr != nil || strings.TrimSpace(sha) == "" {
				p.branchErr = err
				if shaErr != nil {
					p.branchErr = shaErr
				}
				return "", false, p.branchErr
			}
			p.branch = "(" + strings.TrimSpace(sha) + ")"
		}
	}

	status, _ := runGitCommand("status --porcelain", staticCwd(p.cwd), p.ttlSeconds)
	p.branchDirty = strings.TrimSpace(status) != ""

	return p.branch, p.branchDirty, nil
}

// GetChanges retrieves uncommitted insertion and deletion counts.
func (p *GitProvider) GetChanges() (types.GitChanges, error) {
	if err := p.checkInsideWorkTree(); err != nil {
		return types.GitChanges{}, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.changesChecked {
		return p.changes, p.changesErr
	}
	p.changesChecked = true

	unstagedStat, _ := runGitCommand("diff --shortstat", staticCwd(p.cwd), p.ttlSeconds)
	stagedStat, _ := runGitCommand("diff --cached --shortstat", staticCwd(p.cwd), p.ttlSeconds)

	uIns, uDel := parseShortStat(unstagedStat)
	sIns, sDel := parseShortStat(stagedStat)

	p.changes = types.GitChanges{
		Insertions: uIns + sIns,
		Deletions:  uDel + sDel,
	}
	return p.changes, nil
}

// GetAheadBehind retrieves ahead/behind commit counts relative to upstream.
func (p *GitProvider) GetAheadBehind() (types.GitAheadBehind, error) {
	if err := p.checkInsideWorkTree(); err != nil {
		return types.GitAheadBehind{}, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.aheadBehindChecked {
		return p.aheadBehind, p.aheadBehindErr
	}
	p.aheadBehindChecked = true

	out, err := runGitCommand("rev-list --left-right --count HEAD...@{upstream}", staticCwd(p.cwd), p.ttlSeconds)
	if err != nil {
		p.aheadBehindErr = types.ErrNoUpstream
		return types.GitAheadBehind{}, p.aheadBehindErr
	}

	ahead, behind, err := parseAheadBehind(out)
	if err != nil {
		p.aheadBehindErr = err
		return types.GitAheadBehind{}, err
	}

	p.aheadBehind = types.GitAheadBehind{
		Ahead:  ahead,
		Behind: behind,
	}
	return p.aheadBehind, nil
}

// GetRepoName retrieves the repository base directory name.
func (p *GitProvider) GetRepoName() (string, error) {
	if err := p.checkInsideWorkTree(); err != nil {
		return "", err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.repoNameChecked {
		return p.repoName, p.repoNameErr
	}
	p.repoNameChecked = true

	topLevel, err := runGitCommand("rev-parse --show-toplevel", staticCwd(p.cwd), p.ttlSeconds)
	if err != nil || strings.TrimSpace(topLevel) == "" {
		p.repoNameErr = errors.New("could not determine repository root")
		return "", p.repoNameErr
	}

	p.repoName = filepath.Base(strings.TrimSpace(topLevel))
	return p.repoName, nil
}
