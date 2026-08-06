package widgets

import (
	"fmt"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestGitBranchWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("git-branch")
	if w == nil {
		t.Fatalf("Git branch widget not found")
	}
	if w.GetDefaultColor() != "brightMagenta" {
		t.Errorf("Expected default color %q, got %q", "brightMagenta", w.GetDefaultColor())
	}

	settings := types.DefaultSettings()
	hideTrue := true
	hideFalse := false

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
			name: "Normal git repository branch",
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
		},
		{
			name: "Non-git directory with Hide=true",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "false", nil
			},
			item:          types.WidgetItem{Type: "git-branch", Hide: &hideTrue},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}},
			expectedTitle: "",
			expectedBody:  "",
		},
		{
			name: "Git command error with Hide=true",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "", fmt.Errorf("git not installed")
			},
			item:          types.WidgetItem{Type: "git-branch", Hide: &hideTrue},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/no-git-dir"}},
			expectedTitle: "",
			expectedBody:  "",
		},
		{
			name: "Git command error with Hide=false",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "", fmt.Errorf("git not installed")
			},
			item:          types.WidgetItem{Type: "git-branch", Hide: &hideFalse},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/no-git-dir"}},
			expectedTitle: "",
			expectedBody:  "⎇ no git",
		},
		{
			name: "Dirty git branch color check",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "true", nil
			},
			item:          types.WidgetItem{Type: "git-branch", Hide: &hideFalse},
			ctx:           types.RenderContext{Data: types.StatusJSON{VCS: &types.VCSInfo{Dirty: &hideTrue}}},
			expectedColor: "brightRed",
		},
	}

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runGitCommand = tt.gitRunner
			if tt.expectedColor != "" {
				if color := w.GetBodyColor(tt.item, tt.ctx); color != tt.expectedColor {
					t.Errorf("Expected body color %q, got %q", tt.expectedColor, color)
				}
				return
			}
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
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

	settings := types.DefaultSettings()
	hideTrue := true

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
			name: "Non-git directory",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "false", fmt.Errorf("not a git repository")
			},
			item:          types.WidgetItem{Type: "git-changes"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}},
			expectedTitle: "",
			expectedBody:  "(no git)",
		},
		{
			name: "Git missing with Hide=true",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "", fmt.Errorf("git not installed")
			},
			item:          types.WidgetItem{Type: "git-changes", Hide: &hideTrue},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/no-git-dir"}},
			expectedTitle: "",
			expectedBody:  "",
		},
	}

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runGitCommand = tt.gitRunner
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
		})
	}
}

func TestParseShortStat(t *testing.T) {
	tests := []struct {
		name               string
		statInput          string
		expectedInsertions int
		expectedDeletions  int
	}{
		{
			name:               "Empty stat input",
			statInput:          "",
			expectedInsertions: 0,
			expectedDeletions:  0,
		},
		{
			name:               "Insertions only",
			statInput:          "5 insertions(+)",
			expectedInsertions: 5,
			expectedDeletions:  0,
		},
		{
			name:               "Deletions only",
			statInput:          "10 deletions(-)",
			expectedInsertions: 0,
			expectedDeletions:  10,
		},
		{
			name:               "Insertions and deletions combined",
			statInput:          "2 files changed, 10 insertions(+), 5 deletions(-)",
			expectedInsertions: 10,
			expectedDeletions:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ins, del := parseShortStat(tt.statInput)
			if ins != tt.expectedInsertions || del != tt.expectedDeletions {
				t.Errorf("For %q expected (%d, %d), got (%d, %d)", tt.statInput, tt.expectedInsertions, tt.expectedDeletions, ins, del)
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
}
