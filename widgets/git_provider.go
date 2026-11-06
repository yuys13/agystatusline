package widgets

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/yuys13/agystatusline/types"
)

// GitProvider implements types.GitDataProvider using git CLI commands with caching.
type GitProvider struct{}

func (p *GitProvider) checkInsideWorkTree(ctx types.RenderContext) error {
	isGit, err := runGitCommand("rev-parse --is-inside-work-tree", ctx, ctx.GitCacheTTLSeconds)
	if err != nil || strings.TrimSpace(isGit) != "true" {
		return types.ErrNotGitRepository
	}
	return nil
}

// GetBranch retrieves current branch name and whether the working tree is dirty.
func (p *GitProvider) GetBranch(ctx types.RenderContext) (string, bool, error) {
	if err := p.checkInsideWorkTree(ctx); err != nil {
		return "", false, err
	}

	b, err := runGitCommand("symbolic-ref --short HEAD", ctx, ctx.GitCacheTTLSeconds)
	if err != nil {
		return "", false, err
	}
	branch := strings.TrimSpace(b)

	status, _ := runGitCommand("status --porcelain", ctx, ctx.GitCacheTTLSeconds)
	dirty := strings.TrimSpace(status) != ""

	return branch, dirty, nil
}

// GetChanges retrieves uncommitted insertion and deletion counts.
func (p *GitProvider) GetChanges(ctx types.RenderContext) (types.GitChanges, error) {
	if err := p.checkInsideWorkTree(ctx); err != nil {
		return types.GitChanges{}, err
	}

	unstagedStat, _ := runGitCommand("diff --shortstat", ctx, ctx.GitCacheTTLSeconds)
	stagedStat, _ := runGitCommand("diff --cached --shortstat", ctx, ctx.GitCacheTTLSeconds)

	uIns, uDel := parseShortStat(unstagedStat)
	sIns, sDel := parseShortStat(stagedStat)

	return types.GitChanges{
		Insertions: uIns + sIns,
		Deletions:  uDel + sDel,
	}, nil
}

// GetAheadBehind retrieves ahead/behind commit counts relative to upstream.
func (p *GitProvider) GetAheadBehind(ctx types.RenderContext) (types.GitAheadBehind, error) {
	if err := p.checkInsideWorkTree(ctx); err != nil {
		return types.GitAheadBehind{}, err
	}

	out, err := runGitCommand("rev-list --left-right --count HEAD...@{upstream}", ctx, ctx.GitCacheTTLSeconds)
	if err != nil {
		return types.GitAheadBehind{}, types.ErrNoUpstream
	}

	ahead, behind, err := parseAheadBehind(out)
	if err != nil {
		return types.GitAheadBehind{}, err
	}

	return types.GitAheadBehind{
		Ahead:  ahead,
		Behind: behind,
	}, nil
}

// GetRepoName retrieves the repository base directory name.
func (p *GitProvider) GetRepoName(ctx types.RenderContext) (string, error) {
	if err := p.checkInsideWorkTree(ctx); err != nil {
		return "", err
	}

	topLevel, err := runGitCommand("rev-parse --show-toplevel", ctx, ctx.GitCacheTTLSeconds)
	if err != nil || strings.TrimSpace(topLevel) == "" {
		return "", errors.New("could not determine repository root")
	}

	return filepath.Base(strings.TrimSpace(topLevel)), nil
}
