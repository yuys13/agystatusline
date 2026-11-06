package widgets

import (
	"errors"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestGitProvider_GetBranch(t *testing.T) {
	tests := []struct {
		name       string
		runner     func(cmd string, ctx CwdResolver, ttl int) (string, error)
		wantBranch string
		wantDirty  bool
		wantErr    bool
	}{
		{
			name: "Clean branch",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				switch cmd {
				case "rev-parse --is-inside-work-tree":
					return "true", nil
				case "symbolic-ref --short HEAD":
					return "main\n", nil
				case "status --porcelain":
					return "", nil
				default:
					return "", nil
				}
			},
			wantBranch: "main",
			wantDirty:  false,
			wantErr:    false,
		},
		{
			name: "Dirty branch",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				switch cmd {
				case "rev-parse --is-inside-work-tree":
					return "true", nil
				case "symbolic-ref --short HEAD":
					return "feature/test\n", nil
				case "status --porcelain":
					return " M file.go\n", nil
				default:
					return "", nil
				}
			},
			wantBranch: "feature/test",
			wantDirty:  true,
			wantErr:    false,
		},
		{
			name: "Not a git repository",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "", errors.New("fatal: not a git repository")
			},
			wantBranch: "",
			wantDirty:  false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldRunner := runGitCommand
			defer func() { runGitCommand = oldRunner }()
			runGitCommand = tt.runner

			p := &GitProvider{}
			ctx := types.RenderContext{
				Data:               types.StatusJSON{CWD: "/mock/dir"},
				GitCacheTTLSeconds: 5,
			}

			branch, dirty, err := p.GetBranch(ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
			if branch != tt.wantBranch {
				t.Errorf("GetBranch() branch = %q, want %q", branch, tt.wantBranch)
			}
			if dirty != tt.wantDirty {
				t.Errorf("GetBranch() dirty = %v, want %v", dirty, tt.wantDirty)
			}
		})
	}
}

func TestGitProvider_GetChanges(t *testing.T) {
	tests := []struct {
		name           string
		runner         func(cmd string, ctx CwdResolver, ttl int) (string, error)
		wantInsertions int
		wantDeletions  int
		wantErr        bool
	}{
		{
			name: "With unstaged and staged changes",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				switch cmd {
				case "rev-parse --is-inside-work-tree":
					return "true", nil
				case "diff --shortstat":
					return " 1 file changed, 10 insertions(+), 2 deletions(-)", nil
				case "diff --cached --shortstat":
					return " 2 files changed, 5 insertions(+), 3 deletions(-)", nil
				default:
					return "", nil
				}
			},
			wantInsertions: 15,
			wantDeletions:  5,
			wantErr:        false,
		},
		{
			name: "Not a git repository",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "", errors.New("not git")
			},
			wantInsertions: 0,
			wantDeletions:  0,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldRunner := runGitCommand
			defer func() { runGitCommand = oldRunner }()
			runGitCommand = tt.runner

			p := &GitProvider{}
			ctx := types.RenderContext{
				Data:               types.StatusJSON{CWD: "/mock/dir"},
				GitCacheTTLSeconds: 5,
			}

			changes, err := p.GetChanges(ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetChanges() error = %v, wantErr %v", err, tt.wantErr)
			}
			if changes.Insertions != tt.wantInsertions {
				t.Errorf("GetChanges() insertions = %d, want %d", changes.Insertions, tt.wantInsertions)
			}
			if changes.Deletions != tt.wantDeletions {
				t.Errorf("GetChanges() deletions = %d, want %d", changes.Deletions, tt.wantDeletions)
			}
		})
	}
}

func TestGitProvider_GetAheadBehind(t *testing.T) {
	tests := []struct {
		name       string
		runner     func(cmd string, ctx CwdResolver, ttl int) (string, error)
		wantAhead  int
		wantBehind int
		wantErr    bool
	}{
		{
			name: "Ahead and behind",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				switch cmd {
				case "rev-parse --is-inside-work-tree":
					return "true", nil
				case "rev-list --left-right --count HEAD...@{upstream}":
					return "3\t2\n", nil
				default:
					return "", nil
				}
			},
			wantAhead:  3,
			wantBehind: 2,
			wantErr:    false,
		},
		{
			name: "No upstream configured",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				switch cmd {
				case "rev-parse --is-inside-work-tree":
					return "true", nil
				case "rev-list --left-right --count HEAD...@{upstream}":
					return "", errors.New("fatal: no upstream configured")
				default:
					return "", nil
				}
			},
			wantAhead:  0,
			wantBehind: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldRunner := runGitCommand
			defer func() { runGitCommand = oldRunner }()
			runGitCommand = tt.runner

			p := &GitProvider{}
			ctx := types.RenderContext{
				Data:               types.StatusJSON{CWD: "/mock/dir"},
				GitCacheTTLSeconds: 5,
			}

			ab, err := p.GetAheadBehind(ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAheadBehind() error = %v, wantErr %v", err, tt.wantErr)
			}
			if ab.Ahead != tt.wantAhead {
				t.Errorf("GetAheadBehind() ahead = %d, want %d", ab.Ahead, tt.wantAhead)
			}
			if ab.Behind != tt.wantBehind {
				t.Errorf("GetAheadBehind() behind = %d, want %d", ab.Behind, tt.wantBehind)
			}
		})
	}
}

func TestGitProvider_GetRepoName(t *testing.T) {
	tests := []struct {
		name     string
		runner   func(cmd string, ctx CwdResolver, ttl int) (string, error)
		wantRepo string
		wantErr  bool
	}{
		{
			name: "Valid repo name",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				switch cmd {
				case "rev-parse --is-inside-work-tree":
					return "true", nil
				case "rev-parse --show-toplevel":
					return "/path/to/my-repo\n", nil
				default:
					return "", nil
				}
			},
			wantRepo: "my-repo",
			wantErr:  false,
		},
		{
			name: "Not a git repo",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "", errors.New("not git")
			},
			wantRepo: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldRunner := runGitCommand
			defer func() { runGitCommand = oldRunner }()
			runGitCommand = tt.runner

			p := &GitProvider{}
			ctx := types.RenderContext{
				Data:               types.StatusJSON{CWD: "/mock/dir"},
				GitCacheTTLSeconds: 5,
			}

			repo, err := p.GetRepoName(ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetRepoName() error = %v, wantErr %v", err, tt.wantErr)
			}
			if repo != tt.wantRepo {
				t.Errorf("GetRepoName() repo = %q, want %q", repo, tt.wantRepo)
			}
		})
	}
}
