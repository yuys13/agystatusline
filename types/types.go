package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ModelInfo wraps the model string or object representation.
type ModelInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// UnmarshalJSON implements custom unmarshaling for ModelInfo,
// which can be either a plain string or a JSON object.
func (m *ModelInfo) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		m.ID = s
		m.DisplayName = s
		return nil
	}

	type alias ModelInfo
	var obj alias
	if err := json.Unmarshal(data, &obj); err == nil {
		*m = ModelInfo(obj)
		return nil
	}

	return errors.New("invalid model info format")
}

// ContextUsage wraps the context usage which can be a number or a JSON object.
type ContextUsage struct {
	InputTokens              float64 `json:"input_tokens"`
	OutputTokens             float64 `json:"output_tokens"`
	CacheCreationInputTokens float64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     float64 `json:"cache_read_input_tokens"`
}

// UnmarshalJSON implements custom unmarshaling for ContextUsage,
// which can be either a plain number or a JSON object.
func (c *ContextUsage) UnmarshalJSON(data []byte) error {
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		c.InputTokens = num
		return nil
	}

	type alias ContextUsage
	var obj alias
	if err := json.Unmarshal(data, &obj); err == nil {
		*c = ContextUsage(obj)
		return nil
	}

	return errors.New("invalid context usage format")
}

type WorkspaceInfo struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

type OutputStyleInfo struct {
	Name string `json:"name"`
}

type EffortInfo struct {
	Level string `json:"level"`
}

type CostInfo struct {
	TotalCostUSD       *float64 `json:"total_cost_usd,omitempty"`
	TotalDurationMS    *float64 `json:"total_duration_ms,omitempty"`
	TotalAPIDurationMS *float64 `json:"total_api_duration_ms,omitempty"`
	TotalLinesAdded    *float64 `json:"total_lines_added,omitempty"`
	TotalLinesRemoved  *float64 `json:"total_lines_removed,omitempty"`
}

type ContextWindowInfo struct {
	ContextWindowSize   *float64      `json:"context_window_size,omitempty"`
	TotalInputTokens    *float64      `json:"total_input_tokens,omitempty"`
	TotalOutputTokens   *float64      `json:"total_output_tokens,omitempty"`
	CurrentUsage        *ContextUsage `json:"current_usage,omitempty"`
	UsedPercentage      *float64      `json:"used_percentage,omitempty"`
	RemainingPercentage *float64      `json:"remaining_percentage,omitempty"`
}

type VimInfo struct {
	Mode string `json:"mode"`
}

type WorktreeInfo struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	Branch         string `json:"branch"`
	OriginalCwd    string `json:"original_cwd"`
	OriginalBranch string `json:"original_branch"`
}

type VCSInfo struct {
	Type   string `json:"type,omitempty"`
	Branch string `json:"branch,omitempty"`
	Dirty  *bool  `json:"dirty,omitempty"`
}

type RateLimitPeriod struct {
	UsedPercentage *float64 `json:"used_percentage,omitempty"`
	ResetsAt       *float64 `json:"resets_at,omitempty"`
}

type RateLimitsInfo struct {
	FiveHour       *RateLimitPeriod `json:"five_hour,omitempty"`
	SevenDay       *RateLimitPeriod `json:"seven_day,omitempty"`
	SevenDaySonnet *RateLimitPeriod `json:"seven_day_sonnet,omitempty"`
	SevenDayOpus   *RateLimitPeriod `json:"seven_day_opus,omitempty"`
}

type QuotaInfo struct {
	RemainingFraction *float64 `json:"remaining_fraction,omitempty"`
	ResetTime         string   `json:"reset_time,omitempty"`
	ResetInSeconds    *float64 `json:"reset_in_seconds,omitempty"`
}

type SandboxInfo struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// StatusJSON defines the schema for telemetry input streamed on stdin.
type StatusJSON struct {
	HookEventName  string               `json:"hook_event_name"`
	SessionID      string               `json:"session_id"`
	TranscriptPath string               `json:"transcript_path"`
	CWD            string               `json:"cwd"`
	Model          ModelInfo            `json:"model"`
	Workspace      *WorkspaceInfo       `json:"workspace,omitempty"`
	Version        string               `json:"version"`
	OutputStyle    *OutputStyleInfo     `json:"output_style,omitempty"`
	Effort         *EffortInfo          `json:"effort,omitempty"`
	Cost           *CostInfo            `json:"cost,omitempty"`
	ContextWindow  *ContextWindowInfo   `json:"context_window,omitempty"`
	Vim            *VimInfo             `json:"vim,omitempty"`
	Worktree       *WorktreeInfo        `json:"worktree,omitempty"`
	RateLimits     *RateLimitsInfo      `json:"rate_limits,omitempty"`
	Quota          map[string]QuotaInfo `json:"quota,omitempty"`
	Sandbox        *SandboxInfo         `json:"sandbox,omitempty"`
	TerminalWidth  *int                 `json:"terminal_width,omitempty"`
	AgentState     string               `json:"agent_state,omitempty"`
	ArtifactCount  *int                 `json:"artifact_count,omitempty"`
	Subagents      any                  `json:"subagents,omitempty"`
	TaskCount      *int                 `json:"task_count,omitempty"`
	VCS            *VCSInfo             `json:"vcs,omitempty"`
}

// WidgetItem configures a single widget in the statusline.
// It supports plain string representations ("model"), shorthand strings
// ("quota:gemini-5h", "custom-text:PROD", "git-branch: "), and inline tables.
type WidgetItem struct {
	Type   string `toml:"type" json:"type"`
	Key    string `toml:"key,omitempty" json:"key,omitempty"`
	Text   string `toml:"text,omitempty" json:"text,omitempty"`
	Symbol string `toml:"symbol,omitempty" json:"symbol,omitempty"`
	Color  string `toml:"color,omitempty" json:"color,omitempty"`
	Raw    bool   `toml:"raw,omitempty" json:"raw,omitempty"`
}

// UnmarshalTOML implements custom TOML unmarshaling for WidgetItem.
func (w *WidgetItem) UnmarshalTOML(data any) error {
	switch v := data.(type) {
	case string:
		str := v
		if strings.Contains(str, ":") {
			parts := strings.SplitN(str, ":", 2)
			w.Type = strings.TrimSpace(parts[0])
			param := parts[1]
			switch w.Type {
			case "quota", "quota-bar":
				w.Key = strings.TrimSpace(param)
			case "custom-text", "custom":
				w.Type = "custom-text"
				w.Text = param
			case "git-branch":
				w.Symbol = param
			default:
				w.Key = strings.TrimSpace(param)
			}
		} else {
			w.Type = strings.TrimSpace(str)
		}
		if w.Type == "custom" {
			w.Type = "custom-text"
		}
		return nil

	case map[string]any:
		if t, ok := v["type"].(string); ok {
			w.Type = strings.TrimSpace(t)
		}
		if k, ok := v["key"].(string); ok {
			w.Key = strings.TrimSpace(k)
		}
		if txt, ok := v["text"].(string); ok {
			w.Text = txt
		}
		if sym, ok := v["symbol"].(string); ok {
			w.Symbol = sym
		}
		if c, ok := v["color"].(string); ok {
			w.Color = strings.TrimSpace(c)
		}
		if r, ok := v["raw"].(bool); ok {
			w.Raw = r
		}
		if w.Type == "custom" {
			w.Type = "custom-text"
		}
		return nil

	case []byte:
		return w.UnmarshalTOML(string(v))

	default:
		return fmt.Errorf("invalid widget format: expected string or table, got %T", data)
	}
}

// UnmarshalText implements encoding.TextUnmarshaler for WidgetItem,
// allowing go-toml/v2 and standard encoders to parse string representations.
func (w *WidgetItem) UnmarshalText(text []byte) error {
	return w.UnmarshalTOML(string(text))
}

// MarshalTOML implements custom TOML marshaling for WidgetItem.
func (w WidgetItem) MarshalTOML() (any, error) {
	if w.Color == "" && !w.Raw {
		if (w.Type == "quota" || w.Type == "quota-bar") && w.Key != "" && w.Text == "" && w.Symbol == "" {
			return w.Type + ":" + w.Key, nil
		}
		if w.Type == "custom-text" && w.Text != "" && w.Key == "" && w.Symbol == "" {
			return "custom-text:" + w.Text, nil
		}
		if w.Type == "git-branch" && w.Symbol != "" && w.Key == "" && w.Text == "" {
			return "git-branch:" + w.Symbol, nil
		}
		if w.Key == "" && w.Text == "" && w.Symbol == "" && w.Type != "" {
			return w.Type, nil
		}
	}

	m := map[string]any{
		"type": w.Type,
	}
	if w.Key != "" {
		m["key"] = w.Key
	}
	if w.Text != "" {
		m["text"] = w.Text
	}
	if w.Symbol != "" {
		m["symbol"] = w.Symbol
	}
	if w.Color != "" {
		m["color"] = w.Color
	}
	if w.Raw {
		m["raw"] = w.Raw
	}
	return m, nil
}

// PowerlineConfig defines Powerline styling and separator options.
type PowerlineConfig struct {
	Enabled   bool   `toml:"enabled" json:"enabled"`
	Theme     string `toml:"theme" json:"theme"`
	Separator string `toml:"separator" json:"separator"`
	StartCaps string `toml:"start_caps" json:"start_caps"`
	EndCaps   string `toml:"end_caps" json:"end_caps"`
}

// GeneralConfig defines general statusline behavior.
type GeneralConfig struct {
	ColorLevel  int    `toml:"color_level" json:"color_level"`
	GitCacheTTL int    `toml:"git_cache_ttl" json:"git_cache_ttl"`
	Separator   string `toml:"separator" json:"separator"`
	Padding     string `toml:"padding" json:"padding"`
	Minimalist  bool   `toml:"minimalist" json:"minimalist"`
}

// Settings represents the complete agystatusline TOML configuration.
type Settings struct {
	Lines     [][]WidgetItem  `toml:"lines" json:"lines"`
	Powerline PowerlineConfig `toml:"powerline" json:"powerline"`
	General   GeneralConfig   `toml:"general" json:"general"`
}

// DefaultSettings returns the modern default configuration.
func DefaultSettings() Settings {
	return Settings{
		Lines: [][]WidgetItem{
			{
				{Type: "agent-state"},
				{Type: "model"},
				{Type: "context-bar"},
				{Type: "artifacts"},
				{Type: "subagents"},
				{Type: "tasks"},
				{Type: "sandbox"},
			},
		},
		Powerline: PowerlineConfig{
			Enabled:   false,
			Theme:     "nord-aurora",
			Separator: "\uE0B0",
			StartCaps: "",
			EndCaps:   "",
		},
		General: GeneralConfig{
			ColorLevel:  1,
			GitCacheTTL: 5,
			Separator:   " · ",
			Padding:     "",
			Minimalist:  false,
		},
	}
}

// RenderContext holds telemetry data and system metrics needed during render pass.
type RenderContext struct {
	Data               StatusJSON
	TerminalWidth      *int
	IsPreview          bool
	Minimalist         bool
	GitCacheTTLSeconds int
}

func (r RenderContext) GetCwd() string {
	return r.Data.CWD
}

func (r RenderContext) GetWorkspaceCurrentDir() string {
	if r.Data.Workspace != nil {
		return r.Data.Workspace.CurrentDir
	}
	return ""
}

func (r RenderContext) GetWorkspaceProjectDir() string {
	if r.Data.Workspace != nil {
		return r.Data.Workspace.ProjectDir
	}
	return ""
}
