package widgets

import (
	"fmt"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

type testMockGitProvider struct {
	branch      string
	dirty       bool
	branchErr   error
	changes     types.GitChanges
	changesErr  error
	aheadBehind types.GitAheadBehind
	abErr       error
	repo        string
	repoErr     error
}

func (m *testMockGitProvider) GetBranch(ctx types.RenderContext) (string, bool, error) {
	return m.branch, m.dirty, m.branchErr
}

func (m *testMockGitProvider) GetChanges(ctx types.RenderContext) (types.GitChanges, error) {
	return m.changes, m.changesErr
}

func (m *testMockGitProvider) GetAheadBehind(ctx types.RenderContext) (types.GitAheadBehind, error) {
	return m.aheadBehind, m.abErr
}

func (m *testMockGitProvider) GetRepoName(ctx types.RenderContext) (string, error) {
	return m.repo, m.repoErr
}

func TestGitBranchWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("git-branch")
	if w == nil {
		t.Fatalf("Git branch widget not found")
	}
	if w.GetDefaultColor() != "brightMagenta" {
		t.Errorf("Expected default color %q, got %q", "brightMagenta", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Git Branch" {
		t.Errorf("Expected display name %q, got %q", "Git Branch", w.GetDisplayName())
	}

	settings := types.DefaultSettings()
	dirtyTrue := true
	dirtyFalse := false

	tests := []struct {
		name          string
		gitRunner     func(cmd string, ctx CwdResolver, ttl int) (string, error)
		item          types.WidgetItem
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
		expectedColor string
	}{
		{
			name: "Normal git repository branch from command",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "symbolic-ref --short HEAD" {
					return "feature/tdd", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-branch"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "⎇ feature/tdd",
			expectedColor: "brightBlue",
		},
		{
			name: "Custom symbol",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "symbolic-ref --short HEAD" {
					return "main", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-branch", Symbol: "🌿 "},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "🌿 main",
			expectedColor: "brightBlue",
		},
		{
			name: "Branch from VCS telemetry with dirty",
			item: types.WidgetItem{Type: "git-branch"},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					VCS: &types.VCSInfo{Branch: "main", Dirty: &dirtyTrue},
				},
			},
			expectedTitle: "",
			expectedBody:  "⎇ main*",
			expectedColor: "brightRed",
		},
		{
			name: "Branch from VCS telemetry clean with Raw=true",
			item: types.WidgetItem{Type: "git-branch", Raw: true},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					VCS: &types.VCSInfo{Branch: "main", Dirty: &dirtyFalse},
				},
			},
			expectedTitle: "",
			expectedBody:  "main",
			expectedColor: "brightBlue",
		},
		{
			name: "Branch from VCS telemetry dirty with Raw=true",
			item: types.WidgetItem{Type: "git-branch", Raw: true},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					VCS: &types.VCSInfo{Branch: "main", Dirty: &dirtyTrue},
				},
			},
			expectedTitle: "",
			expectedBody:  "main*",
			expectedColor: "brightRed",
		},
		{
			name: "Non-git directory",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "false", nil
			},
			item:          types.WidgetItem{Type: "git-branch"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}},
			expectedTitle: "",
			expectedBody:  "⎇ no git",
			expectedColor: "brightBlue",
		},
		{
			name:          "Git provider branch fallback to main",
			item:          types.WidgetItem{Type: "git-branch"},
			ctx:           types.RenderContext{GitProvider: &testMockGitProvider{branch: "main"}},
			expectedTitle: "",
			expectedBody:  "⎇ main",
			expectedColor: "brightBlue",
		},
	}

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gitRunner != nil {
				runGitCommand = tt.gitRunner
			}
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
			if tt.expectedColor != "" {
				if color := w.GetBodyColor(tt.item, tt.ctx); color != tt.expectedColor {
					t.Errorf("Expected body color %q, got %q", tt.expectedColor, color)
				}
			}
		})
	}
}

func TestGitChangesWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("git-changes")
	if w == nil {
		t.Fatalf("Git changes widget not found")
	}

	if w.GetDefaultColor() != "yellow" {
		t.Errorf("Expected default color %q, got %q", "yellow", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Git Changes" {
		t.Errorf("Expected display name %q, got %q", "Git Changes", w.GetDisplayName())
	}

	settings := types.DefaultSettings()

	tests := []struct {
		name          string
		gitRunner     func(cmd string, ctx CwdResolver, ttl int) (string, error)
		item          types.WidgetItem
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
	}{
		{
			name: "Normal git changes",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "diff --shortstat" {
					return " 2 files changed, 10 insertions(+), 5 deletions(-)", nil
				}
				if cmd == "diff --cached --shortstat" {
					return " 1 file changed, 3 insertions(+)", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-changes"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "(+13,-5)",
		},
		{
			name:          "Git provider changes",
			item:          types.WidgetItem{Type: "git-changes"},
			ctx:           types.RenderContext{GitProvider: &testMockGitProvider{changes: types.GitChanges{Insertions: 42, Deletions: 10}}},
			expectedTitle: "",
			expectedBody:  "(+42,-10)",
		},
		{
			name: "Non-git directory",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "false", fmt.Errorf("not a git repository")
			},
			item:          types.WidgetItem{Type: "git-changes"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}},
			expectedTitle: "",
			expectedBody:  "(no git)",
		},
	}

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gitRunner != nil {
				runGitCommand = tt.gitRunner
			}
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
			if color := w.GetBodyColor(tt.item, tt.ctx); color != "yellow" {
				t.Errorf("Expected body color 'yellow', got %q", color)
			}
		})
	}
}

func TestGitAheadBehindWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("git-ahead-behind")
	if w == nil {
		t.Fatalf("Git ahead-behind widget not found")
	}

	if w.GetDefaultColor() != "brightCyan" {
		t.Errorf("Expected default color %q, got %q", "brightCyan", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Git Ahead/Behind" {
		t.Errorf("Expected display name %q, got %q", "Git Ahead/Behind", w.GetDisplayName())
	}

	settings := types.DefaultSettings()

	tests := []struct {
		name          string
		gitRunner     func(cmd string, ctx CwdResolver, ttl int) (string, error)
		item          types.WidgetItem
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
		expectedColor string
	}{
		{
			name:          "Git provider ahead-behind",
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{GitProvider: &testMockGitProvider{aheadBehind: types.GitAheadBehind{Ahead: 2, Behind: 1}}},
			expectedTitle: "",
			expectedBody:  "↑2 ↓1",
			expectedColor: "brightCyan",
		},
		{
			name: "Both ahead and behind",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "2\t1\n", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "↑2 ↓1",
			expectedColor: "brightCyan",
		},
		{
			name: "Ahead only",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "3\t0\n", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "↑3",
			expectedColor: "brightCyan",
		},
		{
			name: "Behind only",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "0\t5\n", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "↓5",
			expectedColor: "brightCyan",
		},
		{
			name: "Up-to-date with upstream (identical)",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "0\t0\n", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "≡",
			expectedColor: "brightCyan",
		},
		{
			name: "Custom symbol",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "1\t0\n", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind", Symbol: "⇅ "},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "⇅ ↑1",
			expectedColor: "brightCyan",
		},
		{
			name: "Custom symbol with Raw=true",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "1\t0\n", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind", Symbol: "⇅ ", Raw: true},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "↑1",
			expectedColor: "brightCyan",
		},
		{
			name: "Non-git directory",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "false", fmt.Errorf("not a git repository")
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}},
			expectedTitle: "",
			expectedBody:  "(no git)",
			expectedColor: "brightCyan",
		},
		{
			name: "No upstream configured",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "", fmt.Errorf("fatal: no upstream configured for branch 'feat'")
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "(no upstream)",
			expectedColor: "brightCyan",
		},
		{
			name: "Malformed rev-list output",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "corrupt data", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "(no upstream)",
			expectedColor: "brightCyan",
		},
	}

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gitRunner != nil {
				runGitCommand = tt.gitRunner
			}
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
			if color := w.GetBodyColor(tt.item, tt.ctx); color != tt.expectedColor {
				t.Errorf("Expected body color %q, got %q", tt.expectedColor, color)
			}
		})
	}
}

func TestGitRepoWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("git-repo")
	if w == nil {
		t.Fatalf("Git repo widget not found")
	}
	if w.GetDefaultColor() != "white" {
		t.Errorf("Expected default color %q, got %q", "white", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Git Repository" {
		t.Errorf("Expected display name %q, got %q", "Git Repository", w.GetDisplayName())
	}

	settings := types.DefaultSettings()

	tests := []struct {
		name          string
		gitRunner     func(cmd string, ctx CwdResolver, ttl int) (string, error)
		item          types.WidgetItem
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
		expectedColor string
	}{
		{
			name: "Normal git repository from command",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-parse --show-toplevel" {
					return "/home/user/src/agystatusline", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-repo"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/home/user/src/agystatusline"}},
			expectedTitle: "",
			expectedBody:  "agystatusline",
			expectedColor: "white",
		},
		{
			name: "Normal git repository sub directory",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-parse --show-toplevel" {
					return "/home/user/src/agystatusline", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-repo"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/home/user/src/agystatusline/widgets"}},
			expectedTitle: "",
			expectedBody:  "agystatusline",
			expectedColor: "white",
		},
		{
			name: "Custom symbol",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-parse --show-toplevel" {
					return "/workspace/awesome-project", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-repo", Symbol: "📦 "},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/workspace/awesome-project"}},
			expectedTitle: "",
			expectedBody:  "📦 awesome-project",
			expectedColor: "white",
		},
		{
			name: "Custom symbol with Raw=true",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-parse --show-toplevel" {
					return "/workspace/awesome-project", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-repo", Symbol: "📦 ", Raw: true},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/workspace/awesome-project"}},
			expectedTitle: "",
			expectedBody:  "awesome-project",
			expectedColor: "white",
		},
		{
			name: "Non-git directory",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "false", fmt.Errorf("not a git repository")
			},
			item:          types.WidgetItem{Type: "git-repo"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}},
			expectedTitle: "",
			expectedBody:  "(no git)",
			expectedColor: "white",
		},
		{
			name: "Non-git directory with custom symbol",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "false", fmt.Errorf("not a git repository")
			},
			item:          types.WidgetItem{Type: "git-repo", Symbol: "📦 "},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}},
			expectedTitle: "",
			expectedBody:  "📦 (no git)",
			expectedColor: "white",
		},
		{
			name: "Non-git directory with custom symbol and Raw=true",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "false", fmt.Errorf("not a git repository")
			},
			item:          types.WidgetItem{Type: "git-repo", Symbol: "📦 ", Raw: true},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}},
			expectedTitle: "",
			expectedBody:  "(no git)",
			expectedColor: "white",
		},
		{
			name: "Git show-toplevel failure",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-parse --show-toplevel" {
					return "", fmt.Errorf("rev-parse failed")
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-repo"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/workspace/corrupt"}},
			expectedTitle: "",
			expectedBody:  "(no git)",
			expectedColor: "white",
		},
		{
			name:          "Git provider default",
			item:          types.WidgetItem{Type: "git-repo"},
			ctx:           types.RenderContext{GitProvider: &testMockGitProvider{repo: "agystatusline"}},
			expectedTitle: "",
			expectedBody:  "agystatusline",
			expectedColor: "white",
		},
		{
			name:          "Git provider with custom symbol",
			item:          types.WidgetItem{Type: "git-repo", Symbol: "📦 "},
			ctx:           types.RenderContext{GitProvider: &testMockGitProvider{repo: "agystatusline"}},
			expectedTitle: "",
			expectedBody:  "📦 agystatusline",
			expectedColor: "white",
		},
		{
			name:          "Git provider with custom symbol and Raw=true",
			item:          types.WidgetItem{Type: "git-repo", Symbol: "📦 ", Raw: true},
			ctx:           types.RenderContext{GitProvider: &testMockGitProvider{repo: "agystatusline"}},
			expectedTitle: "",
			expectedBody:  "agystatusline",
			expectedColor: "white",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldRunner := runGitCommand
			defer func() { runGitCommand = oldRunner }()
			if tt.gitRunner != nil {
				runGitCommand = tt.gitRunner
			}

			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Unexpected render error: %v", err)
			}
			if title != tt.expectedTitle {
				t.Errorf("Expected title %q, got %q", tt.expectedTitle, title)
			}
			if body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
			color := w.GetBodyColor(tt.item, tt.ctx)
			if color != tt.expectedColor {
				t.Errorf("Expected color %q, got %q", tt.expectedColor, color)
			}
		})
	}
}

func TestGitWidgets_CommandTimeoutAndFallback(t *testing.T) {
	RegisterAll()
	settings := types.DefaultSettings()

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	// Mock a timing out or slow runner
	runGitCommand = func(command string, ctx CwdResolver, ttlSeconds int) (string, error) {
		if command == "rev-parse --is-inside-work-tree" {
			return "", fmt.Errorf("git command timed out after 200ms")
		}
		return "", fmt.Errorf("git error")
	}

	gb := GetWidget("git-branch")
	ctx := types.RenderContext{
		Data: types.StatusJSON{CWD: "/some/path"},
	}

	title, body, err := gb.Render(types.WidgetItem{Type: "git-branch"}, ctx, settings)
	if err != nil {
		t.Fatalf("GitBranch render error: %v", err)
	}
	if title != "" || body != "⎇ no git" {
		t.Errorf("Expected fallback '⎇ no git', got title %q, body %q", title, body)
	}

	gc := GetWidget("git-changes")
	title, body, err = gc.Render(types.WidgetItem{Type: "git-changes"}, ctx, settings)
	if err != nil {
		t.Fatalf("GitChanges render error: %v", err)
	}
	if title != "" || body != "(no git)" {
		t.Errorf("Expected fallback '(no git)', got title %q, body %q", title, body)
	}

	gah := GetWidget("git-ahead-behind")
	title, body, err = gah.Render(types.WidgetItem{Type: "git-ahead-behind"}, ctx, settings)
	if err != nil {
		t.Fatalf("GitAheadBehind render error: %v", err)
	}
	if title != "" || body != "(no git)" {
		t.Errorf("Expected fallback '(no git)', got title %q, body %q", title, body)
	}

	gr := GetWidget("git-repo")
	title, body, err = gr.Render(types.WidgetItem{Type: "git-repo"}, ctx, settings)
	if err != nil {
		t.Fatalf("GitRepo render error: %v", err)
	}
	if title != "" || body != "(no git)" {
		t.Errorf("Expected fallback '(no git)', got title %q, body %q", title, body)
	}
}

func TestGitWidgets_WithGitDataProvider(t *testing.T) {
	RegisterAll()
	settings := types.DefaultSettings()

	mockProvider := &testMockGitProvider{
		branch:      "main",
		dirty:       true,
		changes:     types.GitChanges{Insertions: 42, Deletions: 10},
		aheadBehind: types.GitAheadBehind{Ahead: 2, Behind: 1},
		repo:        "agystatusline",
	}

	ctx := types.RenderContext{
		GitProvider: mockProvider,
		IsPreview:   false, // Explicitly false: proves widgets rely on provider, NOT IsPreview!
	}

	// 1. Git Branch
	gb := GetWidget("git-branch")
	_, body, err := gb.Render(types.WidgetItem{Type: "git-branch"}, ctx, settings)
	if err != nil {
		t.Fatalf("git-branch error: %v", err)
	}
	if body != "⎇ main*" {
		t.Errorf("git-branch body = %q, want %q", body, "⎇ main*")
	}

	// 2. Git Changes
	gc := GetWidget("git-changes")
	_, body, err = gc.Render(types.WidgetItem{Type: "git-changes"}, ctx, settings)
	if err != nil {
		t.Fatalf("git-changes error: %v", err)
	}
	if body != "(+42,-10)" {
		t.Errorf("git-changes body = %q, want %q", body, "(+42,-10)")
	}

	// 3. Git Ahead/Behind
	gab := GetWidget("git-ahead-behind")
	_, body, err = gab.Render(types.WidgetItem{Type: "git-ahead-behind"}, ctx, settings)
	if err != nil {
		t.Fatalf("git-ahead-behind error: %v", err)
	}
	if body != "↑2 ↓1" {
		t.Errorf("git-ahead-behind body = %q, want %q", body, "↑2 ↓1")
	}

	// 4. Git Repo
	gr := GetWidget("git-repo")
	_, body, err = gr.Render(types.WidgetItem{Type: "git-repo"}, ctx, settings)
	if err != nil {
		t.Fatalf("git-repo error: %v", err)
	}
	if body != "agystatusline" {
		t.Errorf("git-repo body = %q, want %q", body, "agystatusline")
	}
}
