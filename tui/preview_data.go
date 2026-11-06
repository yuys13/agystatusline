package tui

import (
	"github.com/yuys13/agystatusline/types"
)

// MockGitProvider provides deterministic dummy Git metrics for TUI live preview.
type MockGitProvider struct {
	Branch     string
	Dirty      bool
	Insertions int
	Deletions  int
	Ahead      int
	Behind     int
	RepoName   string
}

func (m *MockGitProvider) GetBranch(ctx types.RenderContext) (string, bool, error) {
	branch := m.Branch
	if branch == "" {
		branch = "main"
	}
	return branch, m.Dirty, nil
}

func (m *MockGitProvider) GetChanges(ctx types.RenderContext) (types.GitChanges, error) {
	return types.GitChanges{
		Insertions: m.Insertions,
		Deletions:  m.Deletions,
	}, nil
}

func (m *MockGitProvider) GetAheadBehind(ctx types.RenderContext) (types.GitAheadBehind, error) {
	return types.GitAheadBehind{
		Ahead:  m.Ahead,
		Behind: m.Behind,
	}, nil
}

func (m *MockGitProvider) GetRepoName(ctx types.RenderContext) (string, error) {
	repo := m.RepoName
	if repo == "" {
		repo = "agystatusline"
	}
	return repo, nil
}

// DefaultPreviewContext constructs a realistic dummy RenderContext for live preview.
func DefaultPreviewContext(settings types.Settings) types.RenderContext {
	width := 160
	inputTokens := float64(14200)
	usedPct := float64(20.0)
	remainingPct := float64(80.0)
	g5hFraction := 0.5019274
	g5hReset := 8891.0
	p35hFraction := 1.0
	p35hReset := 17956.0
	gWkFraction := 0.9090967
	gWkReset := 567440.0
	gArgonFraction := 1.0
	gArgonReset := 604724.0
	p3WkFraction := 1.0
	p3WkReset := 604756.0
	sandboxEnabled := true
	artCount := 3
	taskCnt := 1

	return types.RenderContext{
		TerminalWidth:      &width,
		IsPreview:          true,
		Minimalist:         settings.General.Minimalist,
		GitCacheTTLSeconds: settings.General.GitCacheTTL,
		GitProvider: &MockGitProvider{
			Branch:     "main",
			Dirty:      false,
			Insertions: 42,
			Deletions:  10,
			Ahead:      2,
			Behind:     1,
			RepoName:   "agystatusline",
		},
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID:          "gemini-3.5-flash-medium",
				DisplayName: "Gemini 3.5 Flash (Medium)",
			},
			ContextWindow: &types.ContextWindowInfo{
				TotalInputTokens:    &inputTokens,
				UsedPercentage:      &usedPct,
				RemainingPercentage: &remainingPct,
			},
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &g5hFraction,
					ResetInSeconds:    &g5hReset,
				},
				"gemini-weekly": {
					RemainingFraction: &gWkFraction,
					ResetInSeconds:    &gWkReset,
				},
				"gemini-argon-wtu-weekly": {
					RemainingFraction: &gArgonFraction,
					ResetInSeconds:    &gArgonReset,
				},
				"3p-5h": {
					RemainingFraction: &p35hFraction,
					ResetInSeconds:    &p35hReset,
				},
				"3p-weekly": {
					RemainingFraction: &p3WkFraction,
					ResetInSeconds:    &p3WkReset,
				},
			},
			Sandbox: &types.SandboxInfo{
				Enabled: &sandboxEnabled,
			},
			AgentState:    "idle",
			ArtifactCount: &artCount,
			Subagents:     []string{"subagent-1", "subagent-2"},
			TaskCount:     &taskCnt,
		},
	}
}
