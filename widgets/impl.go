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

// ContextUsedPctWidget displays context window used percentage.
type ContextUsedPctWidget struct{}

func (c *ContextUsedPctWidget) GetDefaultColor() string { return "brightBlack" }
func (c *ContextUsedPctWidget) GetDisplayName() string  { return "Context Used %" }

func (c *ContextUsedPctWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, error) {
	var pct float64
	if ctx.Data.ContextWindow != nil && ctx.Data.ContextWindow.UsedPercentage != nil {
		pct = *ctx.Data.ContextWindow.UsedPercentage
	} else {
		return "", nil
	}

	if item.RawValue != nil && *item.RawValue {
		return fmt.Sprintf("%.2f%%", pct), nil
	}
	return fmt.Sprintf("Used: %.2f%%", pct), nil
}

// ContextRemainingPctWidget displays context window remaining percentage.
type ContextRemainingPctWidget struct{}

func (c *ContextRemainingPctWidget) GetDefaultColor() string { return "brightBlack" }
func (c *ContextRemainingPctWidget) GetDisplayName() string  { return "Context Remaining %" }

func (c *ContextRemainingPctWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, error) {
	var pct float64
	if ctx.Data.ContextWindow != nil && ctx.Data.ContextWindow.RemainingPercentage != nil {
		pct = *ctx.Data.ContextWindow.RemainingPercentage
	} else {
		return "", nil
	}

	if item.RawValue != nil && *item.RawValue {
		return fmt.Sprintf("%.2f%%", pct), nil
	}
	return fmt.Sprintf("Remaining: %.2f%%", pct), nil
}

// QuotaWidget displays quota limits and usage.
type QuotaWidget struct{}

func (q *QuotaWidget) GetDefaultColor() string { return "brightBlack" }
func (q *QuotaWidget) GetDisplayName() string  { return "Quota" }

func (q *QuotaWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, error) {
	if ctx.Data.Quota == nil {
		return "", nil
	}

	key := item.Metadata["key"]
	if key == "" {
		return "", nil
	}

	quota, ok := ctx.Data.Quota[key]
	if !ok {
		return "", nil
	}

	displayMode := item.Metadata["display"]
	var valueStr string

	var pctStr string
	if quota.RemainingFraction != nil {
		pct := (*quota.RemainingFraction) * 100.0
		pctStr = fmt.Sprintf("%.2f%%", pct)
	}

	var resetStr string
	if quota.ResetInSeconds != nil {
		secs := max(int(*quota.ResetInSeconds), 0)
		if secs < 60 {
			resetStr = fmt.Sprintf("%ds", secs)
		} else if secs < 3600 {
			m := secs / 60
			s := secs % 60
			if s > 0 {
				resetStr = fmt.Sprintf("%dm %ds", m, s)
			} else {
				resetStr = fmt.Sprintf("%dm", m)
			}
		} else if secs < 86400 {
			h := secs / 3600
			m := (secs % 3600) / 60
			if m > 0 {
				resetStr = fmt.Sprintf("%dh %dm", h, m)
			} else {
				resetStr = fmt.Sprintf("%dh", h)
			}
		} else {
			d := secs / 86400
			h := (secs % 86400) / 3600
			if h > 0 {
				resetStr = fmt.Sprintf("%dd %dh", d, h)
			} else {
				resetStr = fmt.Sprintf("%dd", d)
			}
		}
	}

	if displayMode == "reset" {
		if resetStr == "" {
			return "", nil
		}
		valueStr = resetStr
	} else if displayMode == "quota" {
		if pctStr == "" {
			return "", nil
		}
		valueStr = pctStr
	} else {
		// Default: quota % + reset countdown
		if pctStr != "" && resetStr != "" {
			valueStr = fmt.Sprintf("%s (%s)", pctStr, resetStr)
		} else if pctStr != "" {
			valueStr = pctStr
		} else if resetStr != "" {
			valueStr = resetStr
		} else {
			return "", nil
		}
	}

	if item.RawValue != nil && *item.RawValue {
		return valueStr, nil
	}

	label := item.CustomText
	if label == "" {
		label = key
	}

	if displayMode == "reset" {
		return fmt.Sprintf("%s (reset): %s", label, valueStr), nil
	}
	return fmt.Sprintf("%s: %s", label, valueStr), nil
}

// CustomTextWidget displays custom user-defined text.
type CustomTextWidget struct{}

func (c *CustomTextWidget) GetDefaultColor() string { return "white" }
func (c *CustomTextWidget) GetDisplayName() string  { return "Custom Text" }

func (c *CustomTextWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, error) {
	return item.CustomText, nil
}

