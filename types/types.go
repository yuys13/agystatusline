package types

import (
	"encoding/json"
	"errors"
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
type WidgetItem struct {
	ID              string            `json:"id"`
	Type            string            `json:"type"`
	Color           string            `json:"color,omitempty"`
	BackgroundColor string            `json:"backgroundColor,omitempty"`
	Bold            *bool             `json:"bold,omitempty"`
	Dim             any               `json:"dim,omitempty"` // bool or string ("parens")
	Character       string            `json:"character,omitempty"`
	RawValue        *bool             `json:"rawValue,omitempty"`
	CustomText      string            `json:"customText,omitempty"`
	CustomSymbol    string            `json:"customSymbol,omitempty"`
	CommandPath     string            `json:"commandPath,omitempty"`
	MaxWidth        *int              `json:"maxWidth,omitempty"`
	PreserveColors  *bool             `json:"preserveColors,omitempty"`
	Timeout         *int              `json:"timeout,omitempty"`
	Merge           any               `json:"merge,omitempty"` // bool or string ("no-padding")
	Hide            *bool             `json:"hide,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type PowerlineConfig struct {
	Enabled                   bool     `json:"enabled"`
	Separators                []string `json:"separators"`
	SeparatorInvertBackground []bool   `json:"separatorInvertBackground"`
	StartCaps                 []string `json:"startCaps"`
	EndCaps                   []string `json:"endCaps"`
	Theme                     string   `json:"theme,omitempty"`
	AutoAlign                 bool     `json:"autoAlign"`
	ContinueThemeAcrossLines  bool     `json:"continueThemeAcrossLines"`
}

type UpdateMessage struct {
	Message   string `json:"message,omitempty"`
	Remaining *int   `json:"remaining,omitempty"`
}

type InstallationMetadata struct {
	Method           string `json:"method"`
	PackageManager   string `json:"packageManager,omitempty"`
	InstalledVersion string `json:"installedVersion,omitempty"`
}

// Settings represents ccstatusline configurations in settings.json.
type Settings struct {
	Version                 int                   `json:"version"`
	Lines                   [][]WidgetItem        `json:"lines"`
	FlexMode                string                `json:"flexMode"`
	CompactThreshold        int                   `json:"compactThreshold"`
	ColorLevel              int                   `json:"colorLevel"`
	DefaultSeparator        string                `json:"defaultSeparator,omitempty"`
	DefaultPadding          string                `json:"defaultPadding,omitempty"`
	InheritSeparatorColors  bool                  `json:"inheritSeparatorColors"`
	OverrideBackgroundColor string                `json:"overrideBackgroundColor,omitempty"`
	OverrideForegroundColor string                `json:"overrideForegroundColor,omitempty"`
	GlobalBold              bool                  `json:"globalBold"`
	GitCacheTTLSeconds      int                   `json:"gitCacheTtlSeconds"`
	MinimalistMode          bool                  `json:"minimalistMode"`
	Powerline               PowerlineConfig       `json:"powerline"`
	UpdateMessage           *UpdateMessage        `json:"updatemessage,omitempty"`
	Installation            *InstallationMetadata `json:"installation,omitempty"`
}

// DefaultSettings returns configuration defaults mapped from Zod schema defaults.
func DefaultSettings() Settings {
	return Settings{
		Version: 3,
		Lines: [][]WidgetItem{
			{
				{ID: "1", Type: "agent-state", Color: "brightGreen"},
				{ID: "2", Type: "model", Color: "brightMagenta"},
				{ID: "3", Type: "context-bar", Color: "brightWhite"},
				{ID: "4", Type: "artifacts", Color: "brightWhite"},
				{ID: "5", Type: "subagents", Color: "brightWhite"},
				{ID: "6", Type: "tasks", Color: "brightWhite"},
				{ID: "7", Type: "sandbox", Color: "yellow"},
			},
			{},
			{},
		},
		FlexMode:           "full-minus-40",
		CompactThreshold:   60,
		ColorLevel:         2,
		GitCacheTTLSeconds: 5,
		Powerline: PowerlineConfig{
			Enabled:                   false,
			Separators:                []string{"\uE0B0"},
			SeparatorInvertBackground: []bool{false},
			StartCaps:                 []string{},
			EndCaps:                   []string{},
			Theme:                     "nord-aurora",
			AutoAlign:                 false,
			ContinueThemeAcrossLines:  false,
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
