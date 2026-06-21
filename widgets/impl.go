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

func (m *ModelWidget) GetDefaultColor() string { return "brightMagenta" }
func (m *ModelWidget) GetDisplayName() string  { return "Model" }
func (m *ModelWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightMagenta"
}

func (m *ModelWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	displayName := ctx.Data.Model.DisplayName
	if displayName == "" {
		displayName = ctx.Data.Model.ID
	}

	if displayName == "" {
		return "", "", nil
	}

	modelName := strings.TrimSpace(displayName)

	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		// Under statusline.sh spec, model name is italic magenta
		return "", "\x1b[3m\x1b[95m" + modelName + "\x1b[23m\x1b[39m", nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", modelName, nil
	}
	return "", modelName, nil
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
	symbol := "⎇"
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

func formatResetInSeconds(resetInSeconds *float64) string {
	if resetInSeconds == nil {
		return ""
	}
	secs := max(int(*resetInSeconds), 0)
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	} else if secs < 3600 {
		m := secs / 60
		s := secs % 60
		if s > 0 {
			return fmt.Sprintf("%dm %ds", m, s)
		} else {
			return fmt.Sprintf("%dm", m)
		}
	} else if secs < 86400 {
		h := secs / 3600
		m := (secs % 3600) / 60
		if m > 0 {
			return fmt.Sprintf("%dh %dm", h, m)
		} else {
			return fmt.Sprintf("%dh", h)
		}
	} else {
		d := secs / 86400
		h := (secs % 86400) / 3600
		if h > 0 {
			return fmt.Sprintf("%dd %dh", d, h)
		} else {
			return fmt.Sprintf("%dd", d)
		}
	}
}

// QuotaWidget displays quota limits and usage.
type QuotaWidget struct{}

func (q *QuotaWidget) GetDefaultColor() string { return "brightWhite" }
func (q *QuotaWidget) GetDisplayName() string  { return "Quota" }
func (q *QuotaWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightWhite"
}

func (q *QuotaWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	if ctx.Data.Quota == nil {
		return "", "", nil
	}

	key := item.Metadata["key"]
	if key == "" {
		return "", "", nil
	}

	quota, ok := ctx.Data.Quota[key]
	if !ok {
		return "", "", nil
	}

	displayMode := item.Metadata["display"]
	var valueStr string

	var pctStr string
	if quota.RemainingFraction != nil {
		pct := (*quota.RemainingFraction) * 100.0
		pctStr = fmt.Sprintf("%.2f%%", pct)
	}

	resetStr := formatResetInSeconds(quota.ResetInSeconds)

	if displayMode == "reset" {
		if resetStr == "" {
			return "", "", nil
		}
		valueStr = resetStr
	} else if displayMode == "quota" {
		if pctStr == "" {
			return "", "", nil
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
			return "", "", nil
		}
	}

	if item.RawValue != nil && *item.RawValue {
		return "", valueStr, nil
	}

	label := item.CustomText
	if label == "" {
		label = key
	}

	if displayMode == "reset" {
		return label + " (reset)", valueStr, nil
	}
	return label, valueStr, nil
}

// CustomTextWidget displays custom user-defined text.
type CustomTextWidget struct{}

func (c *CustomTextWidget) GetDefaultColor() string { return "white" }
func (c *CustomTextWidget) GetDisplayName() string  { return "Custom Text" }
func (c *CustomTextWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "white"
}

func (c *CustomTextWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	return "", item.CustomText, nil
}

// SandboxWidget displays the sandbox enabled status.
type SandboxWidget struct{}

func (s *SandboxWidget) GetDefaultColor() string { return "yellow" }
func (s *SandboxWidget) GetDisplayName() string  { return "Sandbox" }
func (s *SandboxWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	if ctx.Data.Sandbox != nil && ctx.Data.Sandbox.Enabled != nil && *ctx.Data.Sandbox.Enabled {
		return "brightGreen"
	}
	return "brightBlack"
}

func (s *SandboxWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	if ctx.Data.Sandbox == nil || ctx.Data.Sandbox.Enabled == nil {
		return "", "", nil
	}

	enabled := *ctx.Data.Sandbox.Enabled
	valStr := "off"
	if enabled {
		valStr = "on"
	}

	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		// statusline.sh style coloring:
		// sandbox is gray (ansi 90), ON is green and bold (ansi 92), off is gray (ansi 90)
		var bodyStr string
		if enabled {
			bodyStr = "\x1b[90msandbox\x1b[39m \x1b[92m\x1b[1mON\x1b[22m\x1b[39m"
		} else {
			bodyStr = "\x1b[90msandbox off\x1b[39m"
		}
		return "", bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", valStr, nil
	}

	if enabled {
		return "sandbox", "on", nil
	}
	return "sandbox", "off", nil
}

// AgentStateWidget displays the active agent state.
type AgentStateWidget struct{}

func (a *AgentStateWidget) GetDefaultColor() string { return "brightGreen" }
func (a *AgentStateWidget) GetDisplayName() string  { return "Agent State" }

func (a *AgentStateWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	state := ctx.Data.AgentState
	if state == "" {
		state = "idle"
	}
	switch state {
	case "idle":
		return "brightGreen"
	case "thinking":
		return "brightYellow"
	case "working":
		return "brightCyan"
	case "tool_use":
		return "brightMagenta"
	}
	return "white"
}

func (a *AgentStateWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	state := ctx.Data.AgentState
	if state == "" {
		state = "idle"
	}

	var symbolText string
	switch state {
	case "idle":
		symbolText = "● READY"
	case "thinking":
		symbolText = "◆ THINKING"
	case "working":
		symbolText = "⚙ WORKING"
	case "tool_use":
		symbolText = "🔧 TOOL"
	default:
		symbolText = "⏳ " + strings.ToUpper(state)
	}

	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		boldCode := "\x1b[1m"
		resetCode := "\x1b[22m\x1b[39m"
		var colorCode string
		switch state {
		case "idle":
			colorCode = "\x1b[92m"
		case "thinking":
			colorCode = "\x1b[93m"
		case "working":
			colorCode = "\x1b[96m"
		case "tool_use":
			colorCode = "\x1b[95m"
		default:
			colorCode = "\x1b[97m"
		}
		return "", boldCode + colorCode + symbolText + resetCode, nil
	}

	return "", symbolText, nil
}

// ContextBarWidget displays a progress bar representing context window usage.
type ContextBarWidget struct{}

func (c *ContextBarWidget) GetDefaultColor() string { return "brightWhite" }
func (c *ContextBarWidget) GetDisplayName() string  { return "Context Bar" }

func (c *ContextBarWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	var pct float64
	if ctx.Data.ContextWindow != nil && ctx.Data.ContextWindow.UsedPercentage != nil {
		pct = *ctx.Data.ContextWindow.UsedPercentage
	}
	if pct >= 90 {
		return "brightRed"
	} else if pct >= 60 {
		return "brightYellow"
	}
	return "brightWhite"
}

func (c *ContextBarWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	var pct float64
	if ctx.Data.ContextWindow != nil && ctx.Data.ContextWindow.UsedPercentage != nil {
		pct = *ctx.Data.ContextWindow.UsedPercentage
	} else {
		return "ctx", "", nil
	}

	pctInt := int(pct)
	barLen := 15
	filled := pctInt * barLen / 100
	remainder := (pctInt * barLen) % 100

	var barBuilder strings.Builder
	for i := range barLen {
		if i < filled {
			barBuilder.WriteString("█")
		} else if i == filled {
			if remainder >= 75 {
				barBuilder.WriteString("▓")
			} else if remainder >= 50 {
				barBuilder.WriteString("▒")
			} else if remainder >= 25 {
				barBuilder.WriteString("░")
			} else {
				barBuilder.WriteString("·")
			}
		} else {
			barBuilder.WriteString("·")
		}
	}
	bar := barBuilder.String()

	pctFmt := fmt.Sprintf("%.1f%%", pct)

	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		var barColor string
		if pctInt >= 90 {
			barColor = "\x1b[91m"
		} else if pctInt >= 60 {
			barColor = "\x1b[93m"
		} else {
			barColor = "\x1b[97m"
		}
		titleStr := "\x1b[90mctx\x1b[39m"
		bodyStr := barColor + bar + "\x1b[39m \x1b[97m\x1b[1m" + pctFmt + "\x1b[22m\x1b[39m"
		return titleStr, bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", bar + " " + pctFmt, nil
	}
	return "ctx", bar + " " + pctFmt, nil
}

// ArtifactsWidget displays count of artifacts.
type ArtifactsWidget struct{}

func (a *ArtifactsWidget) GetDefaultColor() string { return "brightWhite" }
func (a *ArtifactsWidget) GetDisplayName() string  { return "Artifacts" }
func (a *ArtifactsWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightWhite"
}

func (a *ArtifactsWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	count := 0
	if ctx.Data.ArtifactCount != nil {
		count = *ctx.Data.ArtifactCount
	}

	countStr := strconv.Itoa(count)
	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		titleStr := "\x1b[90martifacts\x1b[39m"
		bodyStr := "\x1b[97m\x1b[1m" + countStr + "\x1b[22m\x1b[39m"
		return titleStr, bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", countStr, nil
	}
	return "artifacts", countStr, nil
}

// SubagentsWidget displays count of subagents.
type SubagentsWidget struct{}

func (s *SubagentsWidget) GetDefaultColor() string { return "brightWhite" }
func (s *SubagentsWidget) GetDisplayName() string  { return "Subagents" }
func (s *SubagentsWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightWhite"
}

func (s *SubagentsWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	count := 0
	if ctx.Data.Subagents != nil {
		if list, ok := ctx.Data.Subagents.([]any); ok {
			count = len(list)
		} else if num, ok := ctx.Data.Subagents.(float64); ok {
			count = int(num)
		}
	}

	countStr := strconv.Itoa(count)
	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		titleStr := "\x1b[90msubagents\x1b[39m"
		bodyStr := "\x1b[97m\x1b[1m" + countStr + "\x1b[22m\x1b[39m"
		return titleStr, bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", countStr, nil
	}
	return "subagents", countStr, nil
}

// TasksWidget displays count of background tasks.
type TasksWidget struct{}

func (t *TasksWidget) GetDefaultColor() string { return "brightWhite" }
func (t *TasksWidget) GetDisplayName() string  { return "Tasks" }
func (t *TasksWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightWhite"
}

func (t *TasksWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	count := 0
	if ctx.Data.TaskCount != nil {
		count = *ctx.Data.TaskCount
	}

	countStr := strconv.Itoa(count)
	preserveColors := item.PreserveColors != nil && *item.PreserveColors
	if preserveColors {
		titleStr := "\x1b[90mtasks\x1b[39m"
		bodyStr := "\x1b[97m\x1b[1m" + countStr + "\x1b[22m\x1b[39m"
		return titleStr, bodyStr, nil
	}

	if item.RawValue != nil && *item.RawValue {
		return "", countStr, nil
	}
	return "tasks", countStr, nil
}

// QuotaBarWidget displays a progress bar representing remaining quota.
type QuotaBarWidget struct{}

func (q *QuotaBarWidget) GetDefaultColor() string { return "brightGreen" }
func (q *QuotaBarWidget) GetDisplayName() string  { return "Quota Bar" }

func (q *QuotaBarWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	if ctx.Data.Quota == nil {
		return "brightGreen"
	}
	key := item.Metadata["key"]
	if key == "" {
		return "brightGreen"
	}
	quota, ok := ctx.Data.Quota[key]
	if !ok || quota.RemainingFraction == nil {
		return "brightGreen"
	}
	pct := *quota.RemainingFraction * 100.0
	if pct >= 50 {
		return "brightGreen"
	} else if pct >= 10 {
		return "brightYellow"
	}
	return "brightRed"
}

func (q *QuotaBarWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	if ctx.Data.Quota == nil {
		return "", "", nil
	}
	key := item.Metadata["key"]
	if key == "" {
		return "", "", nil
	}
	quota, ok := ctx.Data.Quota[key]
	if !ok || quota.RemainingFraction == nil {
		return "", "", nil
	}

	pct := *quota.RemainingFraction * 100.0
	pctInt := int(pct)
	barLen := 15
	filled := pctInt * barLen / 100
	remainder := (pctInt * barLen) % 100

	var barBuilder strings.Builder
	for i := range barLen {
		if i < filled {
			barBuilder.WriteString("█")
		} else if i == filled {
			if remainder >= 75 {
				barBuilder.WriteString("▓")
			} else if remainder >= 50 {
				barBuilder.WriteString("▒")
			} else if remainder >= 25 {
				barBuilder.WriteString("░")
			} else {
				barBuilder.WriteString("·")
			}
		} else {
			barBuilder.WriteString("·")
		}
	}
	bar := barBuilder.String()

	pctFmt := fmt.Sprintf("%.1f%%", pct)

	label := item.CustomText
	if label == "" {
		switch key {
		case "gemini-5h":
			label = "5h"
		case "gemini-weekly":
			label = "weekly"
		default:
			label = key
		}
	}

	resetStr := formatResetInSeconds(quota.ResetInSeconds)
	bodyVal := bar + " " + pctFmt
	if resetStr != "" {
		bodyVal = bodyVal + " (" + resetStr + ")"
	}

	if item.RawValue != nil && *item.RawValue {
		return "", bodyVal, nil
	}
	return label, bodyVal, nil
}
