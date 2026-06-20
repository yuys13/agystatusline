package widgets

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/yuys13/agystatusline/types"
)

// ModelWidget displays the active model name.
type ModelWidget struct{}

func (m *ModelWidget) GetDefaultColor() string { return "cyan" }
func (m *ModelWidget) GetDisplayName() string  { return "Model" }

func (m *ModelWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, error) {
	displayName := ctx.Data.Model.DisplayName
	if displayName == "" {
		displayName = ctx.Data.Model.ID
	}

	if displayName == "" {
		return "", nil
	}

	// Remove parenthesized suffixes e.g. "Claude 3.5 Sonnet (New)" -> "Claude 3.5 Sonnet"
	// but keep (Medium) if present
	re := regexp.MustCompile(`\s*(\(.*\))$`)
	matches := re.FindStringSubmatch(displayName)
	var shortName string
	if len(matches) > 1 && matches[1] == "(Medium)" {
		shortName = strings.TrimSpace(displayName)
	} else {
		shortName = strings.TrimSpace(re.ReplaceAllString(displayName, ""))
	}

	if item.RawValue != nil && *item.RawValue {
		return shortName, nil
	}
	return "Model: " + shortName, nil
}

// ContextLengthWidget displays total input tokens.
type ContextLengthWidget struct{}

func (c *ContextLengthWidget) GetDefaultColor() string { return "brightBlack" }
func (c *ContextLengthWidget) GetDisplayName() string  { return "Context Length" }

func (c *ContextLengthWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, error) {
	var tokens float64

	if ctx.Data.ContextWindow != nil && ctx.Data.ContextWindow.TotalInputTokens != nil {
		tokens = *ctx.Data.ContextWindow.TotalInputTokens
	} else {
		return "", nil
	}

	return formatTokens(tokens, 1), nil
}

// GitBranchWidget displays the current Git branch name.
type GitBranchWidget struct{}

func (g *GitBranchWidget) GetDefaultColor() string { return "magenta" }
func (g *GitBranchWidget) GetDisplayName() string  { return "Git Branch" }

func (g *GitBranchWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, error) {
	symbol := "⎇"
	if item.CustomSymbol != "" {
		symbol = item.CustomSymbol
	}

	if ctx.IsPreview {
		if item.RawValue != nil && *item.RawValue {
			return "main", nil
		}
		return symbol + "main", nil
	}

	// Check if inside git tree
	isGit, err := runGitCommand("rev-parse --is-inside-work-tree", ctx, ctx.GitCacheTTLSeconds)
	if err != nil || isGit != "true" {
		if item.Hide != nil && *item.Hide {
			return "", nil
		}
		return symbol + "no git", nil
	}

	branch, err := runGitCommand("symbolic-ref --short HEAD", ctx, ctx.GitCacheTTLSeconds)
	if err != nil || branch == "" {
		if item.Hide != nil && *item.Hide {
			return "", nil
		}
		return symbol + "no git", nil
	}

	if item.RawValue != nil && *item.RawValue {
		return branch, nil
	}
	return symbol + branch, nil
}

// GitChangesWidget displays the counts of Git insertions and deletions.
type GitChangesWidget struct{}

func (g *GitChangesWidget) GetDefaultColor() string { return "yellow" }
func (g *GitChangesWidget) GetDisplayName() string  { return "Git Changes" }

func (g *GitChangesWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, error) {
	if ctx.IsPreview {
		return "(+42,-10)", nil
	}

	// Check if inside git tree
	isGit, err := runGitCommand("rev-parse --is-inside-work-tree", ctx, ctx.GitCacheTTLSeconds)
	if err != nil || isGit != "true" {
		if item.Hide != nil && *item.Hide {
			return "", nil
		}
		return "(no git)", nil
	}

	unstagedStat, _ := runGitCommand("diff --shortstat", ctx, ctx.GitCacheTTLSeconds)
	stagedStat, _ := runGitCommand("diff --cached --shortstat", ctx, ctx.GitCacheTTLSeconds)

	uIns, uDel := parseShortStat(unstagedStat)
	sIns, sDel := parseShortStat(stagedStat)

	insertions := uIns + sIns
	deletions := uDel + sDel

	return fmt.Sprintf("(+%d,-%d)", insertions, deletions), nil
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

func formatTokens(count float64, decimals int) string {
	div := math.Pow(10, float64(decimals))
	threshold := 1000000.0 - 500.0/div
	if count >= threshold {
		val := count / 1000000.0
		return fmt.Sprintf("%.1fM", val)
	}
	if count >= 1000.0 {
		val := count / 1000.0
		return fmt.Sprintf("%.*fk", decimals, val)
	}
	return fmt.Sprintf("%.0f", count)
}
