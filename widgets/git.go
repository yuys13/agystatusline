package widgets

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/yuys13/agystatusline/types"
)

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
	if item.CustomSymbol != "" {
		symbol = item.CustomSymbol
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

	if branch == "" && !ctx.IsPreview {
		// Fallback to git command
		isGit, err := runGitCommand("rev-parse --is-inside-work-tree", ctx, ctx.GitCacheTTLSeconds)
		if err == nil && isGit == "true" {
			branch, _ = runGitCommand("symbolic-ref --short HEAD", ctx, ctx.GitCacheTTLSeconds)
			status, _ := runGitCommand("status --porcelain", ctx, ctx.GitCacheTTLSeconds)
			dirty = strings.TrimSpace(status) != ""
		}
	}

	if ctx.IsPreview && branch == "" {
		branch = "main"
	}

	if branch == "" {
		if item.Hide != nil && *item.Hide {
			return "", "", nil
		}
		return "", symbol + "no git", nil
	}

	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		// statusline.sh style coloring:
		// branch name is brightBlue (or brightRed if dirty with a brightYellow '*' appended)
		var bodyStr string
		if dirty {
			bodyStr = "\x1b[91m" + branch + "\x1b[39m\x1b[93m*\x1b[39m"
		} else {
			bodyStr = "\x1b[94m" + branch + "\x1b[39m"
		}
		return "", bodyStr, nil
	}

	bodyStr := symbol + branch
	if dirty {
		bodyStr += "*"
	}

	if item.RawValue != nil && *item.RawValue {
		return "", branch, nil
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
	if ctx.IsPreview {
		return "", "(+42,-10)", nil
	}

	// Check if inside git tree
	isGit, err := runGitCommand("rev-parse --is-inside-work-tree", ctx, ctx.GitCacheTTLSeconds)
	if err != nil || isGit != "true" {
		if item.Hide != nil && *item.Hide {
			return "", "", nil
		}
		return "", "(no git)", nil
	}

	unstagedStat, _ := runGitCommand("diff --shortstat", ctx, ctx.GitCacheTTLSeconds)
	stagedStat, _ := runGitCommand("diff --cached --shortstat", ctx, ctx.GitCacheTTLSeconds)

	uIns, uDel := parseShortStat(unstagedStat)
	sIns, sDel := parseShortStat(stagedStat)

	insertions := uIns + sIns
	deletions := uDel + sDel

	return "", fmt.Sprintf("(+%d,-%d)", insertions, deletions), nil
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
