package widgets

import (
	"errors"
	"sync"
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
		wantErrIs  error
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
			wantErrIs:  types.ErrNotGitRepository,
		},
		{
			name: "Detached HEAD with exact tag match (clean)",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				switch cmd {
				case "rev-parse --is-inside-work-tree":
					return "true", nil
				case "symbolic-ref --short HEAD":
					return "", errors.New("fatal: ref HEAD is not a symbolic ref")
				case "describe --tags --exact-match":
					return "v1.2.0\n", nil
				case "status --porcelain":
					return "", nil
				default:
					return "", nil
				}
			},
			wantBranch: "(v1.2.0)",
			wantDirty:  false,
			wantErr:    false,
		},
		{
			name: "Detached HEAD with exact tag match (dirty)",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				switch cmd {
				case "rev-parse --is-inside-work-tree":
					return "true", nil
				case "symbolic-ref --short HEAD":
					return "", errors.New("fatal: ref HEAD is not a symbolic ref")
				case "describe --tags --exact-match":
					return "v1.2.0\n", nil
				case "status --porcelain":
					return " M main.go\n", nil
				default:
					return "", nil
				}
			},
			wantBranch: "(v1.2.0)",
			wantDirty:  true,
			wantErr:    false,
		},
		{
			name: "Detached HEAD without tag match, fallback to short commit SHA (clean)",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				switch cmd {
				case "rev-parse --is-inside-work-tree":
					return "true", nil
				case "symbolic-ref --short HEAD":
					return "", errors.New("fatal: ref HEAD is not a symbolic ref")
				case "describe --tags --exact-match":
					return "", errors.New("fatal: No names found, cannot describe anything.")
				case "rev-parse --short HEAD":
					return "1a2b3c4\n", nil
				case "status --porcelain":
					return "", nil
				default:
					return "", nil
				}
			},
			wantBranch: "(1a2b3c4)",
			wantDirty:  false,
			wantErr:    false,
		},
		{
			name: "Detached HEAD without tag match, fallback to short commit SHA (dirty)",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				switch cmd {
				case "rev-parse --is-inside-work-tree":
					return "true", nil
				case "symbolic-ref --short HEAD":
					return "", errors.New("fatal: ref HEAD is not a symbolic ref")
				case "describe --tags --exact-match":
					return "", errors.New("fatal: No names found, cannot describe anything.")
				case "rev-parse --short HEAD":
					return "1a2b3c4\n", nil
				case "status --porcelain":
					return "?? newfile.txt\n", nil
				default:
					return "", nil
				}
			},
			wantBranch: "(1a2b3c4)",
			wantDirty:  true,
			wantErr:    false,
		},
		{
			name: "Detached HEAD where both tag and rev-parse fail",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				switch cmd {
				case "rev-parse --is-inside-work-tree":
					return "true", nil
				case "symbolic-ref --short HEAD":
					return "", errors.New("fatal: ref HEAD is not a symbolic ref")
				case "describe --tags --exact-match":
					return "", errors.New("fatal: No names found")
				case "rev-parse --short HEAD":
					return "", errors.New("fatal: Needed a single revision")
				default:
					return "", nil
				}
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

			p := NewGitProvider("/mock/dir", 5)

			branch, dirty, err := p.GetBranch()
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("GetBranch() error = %v, wantErrIs %v", err, tt.wantErrIs)
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
		wantErrIs      error
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
			wantErrIs:      types.ErrNotGitRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldRunner := runGitCommand
			defer func() { runGitCommand = oldRunner }()
			runGitCommand = tt.runner

			p := NewGitProvider("/mock/dir", 5)

			changes, err := p.GetChanges()
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetChanges() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("GetChanges() error = %v, wantErrIs %v", err, tt.wantErrIs)
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
		wantErrIs  error
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
			name: "Not a git repository",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "", errors.New("fatal: not a git repository")
			},
			wantAhead:  0,
			wantBehind: 0,
			wantErr:    true,
			wantErrIs:  types.ErrNotGitRepository,
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
			wantErrIs:  types.ErrNoUpstream,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldRunner := runGitCommand
			defer func() { runGitCommand = oldRunner }()
			runGitCommand = tt.runner

			p := NewGitProvider("/mock/dir", 5)

			ab, err := p.GetAheadBehind()
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAheadBehind() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("GetAheadBehind() error = %v, wantErrIs %v", err, tt.wantErrIs)
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
		name      string
		runner    func(cmd string, ctx CwdResolver, ttl int) (string, error)
		wantRepo  string
		wantErr   bool
		wantErrIs error
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
			wantRepo:  "",
			wantErr:   true,
			wantErrIs: types.ErrNotGitRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldRunner := runGitCommand
			defer func() { runGitCommand = oldRunner }()
			runGitCommand = tt.runner

			p := NewGitProvider("/mock/dir", 5)

			repo, err := p.GetRepoName()
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetRepoName() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("GetRepoName() error = %v, wantErrIs %v", err, tt.wantErrIs)
			}
			if repo != tt.wantRepo {
				t.Errorf("GetRepoName() repo = %q, want %q", repo, tt.wantRepo)
			}
		})
	}
}

func TestGitProvider_IntraPassMemoization(t *testing.T) {
	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	callCounts := make(map[string]int)
	var mu sync.Mutex

	runGitCommand = func(cmd string, ctx CwdResolver, ttl int) (string, error) {
		mu.Lock()
		callCounts[cmd]++
		mu.Unlock()

		switch cmd {
		case "rev-parse --is-inside-work-tree":
			return "true", nil
		case "symbolic-ref --short HEAD":
			return "feature/memo\n", nil
		case "status --porcelain":
			return " M file.go\n", nil
		case "diff --shortstat":
			return " 1 file changed, 10 insertions(+), 2 deletions(-)", nil
		case "diff --cached --shortstat":
			return "", nil
		case "rev-list --left-right --count HEAD...@{upstream}":
			return "2\t1\n", nil
		case "rev-parse --show-toplevel":
			return "/path/to/my-repo\n", nil
		default:
			return "", nil
		}
	}

	p := NewGitProvider("/mock/dir", 5)

	// Call GetBranch twice
	b1, d1, err1 := p.GetBranch()
	b2, d2, err2 := p.GetBranch()
	if err1 != nil || err2 != nil || b1 != b2 || d1 != d2 {
		t.Fatalf("GetBranch mismatch or error: %v, %v", err1, err2)
	}

	// Call GetChanges twice
	c1, err1 := p.GetChanges()
	c2, err2 := p.GetChanges()
	if err1 != nil || err2 != nil || c1 != c2 {
		t.Fatalf("GetChanges mismatch or error: %v, %v", err1, err2)
	}

	// Call GetAheadBehind twice
	ab1, err1 := p.GetAheadBehind()
	ab2, err2 := p.GetAheadBehind()
	if err1 != nil || err2 != nil || ab1 != ab2 {
		t.Fatalf("GetAheadBehind mismatch or error: %v, %v", err1, err2)
	}

	// Call GetRepoName twice
	r1, err1 := p.GetRepoName()
	r2, err2 := p.GetRepoName()
	if err1 != nil || err2 != nil || r1 != r2 {
		t.Fatalf("GetRepoName mismatch or error: %v, %v", err1, err2)
	}

	// Verify each underlying command was executed exactly ONCE
	mu.Lock()
	defer mu.Unlock()
	expectedOnce := []string{
		"rev-parse --is-inside-work-tree",
		"symbolic-ref --short HEAD",
		"status --porcelain",
		"diff --shortstat",
		"diff --cached --shortstat",
		"rev-list --left-right --count HEAD...@{upstream}",
		"rev-parse --show-toplevel",
	}

	for _, cmd := range expectedOnce {
		if count := callCounts[cmd]; count != 1 {
			t.Errorf("command %q called %d times, want exactly 1", cmd, count)
		}
	}
}

func TestGitProvider_DetachedHEAD_IntraPassMemoization(t *testing.T) {
	tests := []struct {
		name                 string
		runner               func(cmd string, ctx CwdResolver, ttl int) (string, error)
		wantBranch           string
		wantDirty            bool
		expectedCommandCalls map[string]int
	}{
		{
			name: "Detached HEAD with tag memoization",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				switch cmd {
				case "rev-parse --is-inside-work-tree":
					return "true", nil
				case "symbolic-ref --short HEAD":
					return "", errors.New("fatal: ref HEAD is not a symbolic ref")
				case "describe --tags --exact-match":
					return "v1.2.0\n", nil
				case "status --porcelain":
					return " M file.go\n", nil
				default:
					return "", nil
				}
			},
			wantBranch: "(v1.2.0)",
			wantDirty:  true,
			expectedCommandCalls: map[string]int{
				"rev-parse --is-inside-work-tree": 1,
				"symbolic-ref --short HEAD":       1,
				"describe --tags --exact-match":   1,
				"status --porcelain":              1,
			},
		},
		{
			name: "Detached HEAD with short SHA memoization",
			runner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				switch cmd {
				case "rev-parse --is-inside-work-tree":
					return "true", nil
				case "symbolic-ref --short HEAD":
					return "", errors.New("fatal: ref HEAD is not a symbolic ref")
				case "describe --tags --exact-match":
					return "", errors.New("fatal: No names found")
				case "rev-parse --short HEAD":
					return "1a2b3c4\n", nil
				case "status --porcelain":
					return "", nil
				default:
					return "", nil
				}
			},
			wantBranch: "(1a2b3c4)",
			wantDirty:  false,
			expectedCommandCalls: map[string]int{
				"rev-parse --is-inside-work-tree": 1,
				"symbolic-ref --short HEAD":       1,
				"describe --tags --exact-match":   1,
				"rev-parse --short HEAD":          1,
				"status --porcelain":              1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldRunner := runGitCommand
			defer func() { runGitCommand = oldRunner }()

			callCounts := make(map[string]int)
			var mu sync.Mutex

			runGitCommand = func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				mu.Lock()
				callCounts[cmd]++
				mu.Unlock()
				return tt.runner(cmd, ctx, ttl)
			}

			p := NewGitProvider("/mock/dir", 5)

			// Call GetBranch twice to test memoization
			b1, d1, err1 := p.GetBranch()
			b2, d2, err2 := p.GetBranch()

			if err1 != nil || err2 != nil {
				t.Fatalf("unexpected error: %v, %v", err1, err2)
			}
			if b1 != tt.wantBranch || b2 != tt.wantBranch {
				t.Errorf("got branches %q and %q, want %q", b1, b2, tt.wantBranch)
			}
			if d1 != tt.wantDirty || d2 != tt.wantDirty {
				t.Errorf("got dirties %v and %v, want %v", d1, d2, tt.wantDirty)
			}

			mu.Lock()
			defer mu.Unlock()
			for cmd, expectedCount := range tt.expectedCommandCalls {
				if count := callCounts[cmd]; count != expectedCount {
					t.Errorf("command %q called %d times, want %d", cmd, count, expectedCount)
				}
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

func TestParseAheadBehind(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedAhead  int
		expectedBehind int
		expectErr      bool
	}{
		{
			name:           "Valid tab-separated count",
			input:          "2\t1\n",
			expectedAhead:  2,
			expectedBehind: 1,
			expectErr:      false,
		},
		{
			name:           "Valid space-separated count",
			input:          "0 4",
			expectedAhead:  0,
			expectedBehind: 4,
			expectErr:      false,
		},
		{
			name:           "Both zero",
			input:          "0\t0",
			expectedAhead:  0,
			expectedBehind: 0,
			expectErr:      false,
		},
		{
			name:           "Empty string",
			input:          "",
			expectedAhead:  0,
			expectedBehind: 0,
			expectErr:      true,
		},
		{
			name:           "Single number only",
			input:          "2",
			expectedAhead:  0,
			expectedBehind: 0,
			expectErr:      true,
		},
		{
			name:           "Non-numeric tokens",
			input:          "foo bar",
			expectedAhead:  0,
			expectedBehind: 0,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ahead, behind, err := parseAheadBehind(tt.input)
			if (err != nil) != tt.expectErr {
				t.Fatalf("For %q expected err=%v, got err=%v", tt.input, tt.expectErr, err)
			}
			if ahead != tt.expectedAhead || behind != tt.expectedBehind {
				t.Errorf("For %q expected (%d, %d), got (%d, %d)", tt.input, tt.expectedAhead, tt.expectedBehind, ahead, behind)
			}
		})
	}
}
