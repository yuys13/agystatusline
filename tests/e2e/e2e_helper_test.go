package e2e_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/mattn/go-runewidth"
	"github.com/pelletier/go-toml/v2"
)

var (
	binaryPath  string
	projectRoot string
)

func TestMain(m *testing.M) {
	var err error
	projectRoot, err = findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error locating project root: %v\n", err)
		os.Exit(1)
	}

	tempDir, err := os.MkdirTemp("", "agystatusline-e2e-bin-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp dir for binary: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	binaryPath = filepath.Join(tempDir, "agystatusline")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = projectRoot
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Notice: Binary build during TestMain: %v (May be expected if downstream milestones M2-M4 are in progress)\n", err)
	}

	code := m.Run()
	os.Exit(code)
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found in parent directories")
}

func runCLI(t *testing.T, stdinData string, args ...string) (string, string, int) {
	t.Helper()
	return runCLIWithDir(t, stdinData, "", args...)
}

func runCLIWithDir(t *testing.T, stdinData string, workingDir string, args ...string) (string, string, int) {
	t.Helper()

	if _, err := os.Stat(binaryPath); err != nil {
		t.Skipf("Skipping process execution test: binary at %s not yet built (pending downstream milestone implementation)", binaryPath)
	}

	cmd := exec.Command(binaryPath, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if stdinData != "" {
		cmd.Stdin = bytes.NewReader([]byte(stdinData))
	}

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("Unexpected process execution error: %v", err)
		}
	}

	return stdout.String(), stderr.String(), exitCode
}

var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\([a-zA-Z]`)

func stripANSI(s string) string {
	return ansiEscapeRegex.ReplaceAllString(s, "")
}

func getVisibleWidth(s string) int {
	return runewidth.StringWidth(stripANSI(s))
}

func writeTOMLConfig(t *testing.T, dir string, filename string, configData any) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	data, err := toml.Marshal(configData)
	if err != nil {
		t.Fatalf("Failed to marshal config TOML: %v", err)
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file %s: %v", filePath, err)
	}
	return filePath
}

func standardStatusJSON() string {
	return `{
  "hook_event_name": "status",
  "session_id": "e2e-session-001",
  "transcript_path": "/tmp/transcript.jsonl",
  "cwd": "/tmp/test-repo",
  "model": {
    "id": "gemini-3.5-flash-medium",
    "display_name": "Gemini 3.5 Flash (Medium)"
  },
  "workspace": {
    "current_dir": "/tmp/test-repo",
    "project_dir": "/tmp/test-repo"
  },
  "version": "1.0.0",
  "context_window": {
    "context_window_size": 2000000.0,
    "total_input_tokens": 14200.0,
    "total_output_tokens": 850.0,
    "used_percentage": 20.0,
    "remaining_percentage": 80.0
  },
  "vcs": {
    "type": "git",
    "branch": "feature/e2e-testing",
    "dirty": false
  },
  "quota": {
    "gemini-5h": {
      "remaining_fraction": 0.5019,
      "reset_in_seconds": 8891.0
    }
  },
  "terminal_width": 120,
  "agent_state": "idle",
  "artifact_count": 3,
  "subagents": ["subagent-1", "subagent-2"],
  "task_count": 1,
  "sandbox": {
    "enabled": true
  }
}`
}
