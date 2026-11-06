package widgets

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/yuys13/agystatusline/types"
)

var defaultGitProvider types.GitDataProvider = &GitProvider{}

func getGitProvider(ctx types.RenderContext) types.GitDataProvider {
	if ctx.GitProvider != nil {
		return ctx.GitProvider
	}
	return defaultGitProvider
}

// GitBranchWidget displays the current Git branch name.
type GitBranchWidget struct{}

func (g *GitBranchWidget) GetDefaultColor() string { return "brightMagenta" }
func (g *GitBranchWidget) GetDisplayName() string  { return "Git Branch" }
func (g *GitBranchWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	if ctx.Data.VCS != nil && ctx.Data.VCS.Dirty != nil && *ctx.Data.VCS.Dirty {
		return "brightRed"
	}
	return "brightBlue"
}

func (g *GitBranchWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	symbol := "⎇ "
	if item.Symbol != "" {
		symbol = item.Symbol
	}

	// Try to get branch from VCS telemetry first
	var branch string
	var dirty bool
	if ctx.Data.VCS != nil {
		branch = ctx.Data.VCS.Branch
		if ctx.Data.VCS.Dirty != nil {
			dirty = *ctx.Data.VCS.Dirty
		}
	}

	if branch == "" {
		p := getGitProvider(ctx)
		b, d, err := p.GetBranch(ctx)
		if err == nil {
			branch = b
			dirty = d
		}
	}

	if branch == "" {
		return "", symbol + "no git", nil
	}

	bodyStr := branch
	if !item.Raw {
		bodyStr = symbol + branch
	}
	if dirty {
		bodyStr += "*"
	}

	return "", bodyStr, nil
}

// GitChangesWidget displays the counts of Git insertions and deletions.
type GitChangesWidget struct{}

func (g *GitChangesWidget) GetDefaultColor() string { return "yellow" }
func (g *GitChangesWidget) GetDisplayName() string  { return "Git Changes" }
func (g *GitChangesWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "yellow"
}

func (g *GitChangesWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	p := getGitProvider(ctx)
	changes, err := p.GetChanges(ctx)
	if err != nil {
		return "", "(no git)", nil
	}

	return "", fmt.Sprintf("(+%d,-%d)", changes.Insertions, changes.Deletions), nil
}

func parseShortStat(stat string) (int, int) {
	insertMatch := regexp.MustCompile(`(\d+)\s+insertions?`).FindStringSubmatch(stat)
	deleteMatch := regexp.MustCompile(`(\d+)\s+deletions?`).FindStringSubmatch(stat)

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

// GitAheadBehindWidget displays Git ahead/behind commits count against remote upstream.
type GitAheadBehindWidget struct{}

func (g *GitAheadBehindWidget) GetDefaultColor() string { return "brightCyan" }
func (g *GitAheadBehindWidget) GetDisplayName() string  { return "Git Ahead/Behind" }
func (g *GitAheadBehindWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightCyan"
}

func (g *GitAheadBehindWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	p := getGitProvider(ctx)
	ab, err := p.GetAheadBehind(ctx)
	if err != nil {
		body := "(no upstream)"
		if errors.Is(err, types.ErrNotGitRepository) {
			body = "(no git)"
		}
		if item.Symbol != "" && !item.Raw {
			body = item.Symbol + body
		}
		return "", body, nil
	}

	var body string
	if ab.Ahead > 0 && ab.Behind > 0 {
		body = fmt.Sprintf("↑%d ↓%d", ab.Ahead, ab.Behind)
	} else if ab.Ahead > 0 {
		body = fmt.Sprintf("↑%d", ab.Ahead)
	} else if ab.Behind > 0 {
		body = fmt.Sprintf("↓%d", ab.Behind)
	} else {
		body = "≡"
	}

	if item.Symbol != "" && !item.Raw {
		body = item.Symbol + body
	}

	return "", body, nil
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

// GitRepoWidget displays the current Git repository name.
type GitRepoWidget struct{}

func (g *GitRepoWidget) GetDefaultColor() string { return "white" }
func (g *GitRepoWidget) GetDisplayName() string  { return "Git Repository" }
func (g *GitRepoWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "white"
}

func (g *GitRepoWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	p := getGitProvider(ctx)
	repoName, err := p.GetRepoName(ctx)
	if err != nil {
		body := "(no git)"
		if item.Symbol != "" && !item.Raw {
			body = item.Symbol + body
		}
		return "", body, nil
	}

	body := repoName
	if item.Symbol != "" && !item.Raw {
		body = item.Symbol + body
	}

	return "", body, nil
}
