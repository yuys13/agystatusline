package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	tests := []struct {
		name      string
		homeEnv   *string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:     "empty string remains empty",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only string becomes empty",
			input:    "   \t\n  ",
			expected: "",
		},
		{
			name:     "absolute path remains unchanged",
			input:    "/var/log/status.json",
			expected: "/var/log/status.json",
		},
		{
			name:     "surrounding whitespace on path is trimmed",
			input:    "  ~/.local/share/agystatusline/payload.json  ",
			expected: filepath.Join(home, ".local", "share", "agystatusline", "payload.json"),
		},
		{
			name:     "tilde slash expands to home dir",
			input:    "~/.local/share/agystatusline/payload.json",
			expected: filepath.Join(home, ".local", "share", "agystatusline", "payload.json"),
		},
		{
			name:     "tilde only expands to home dir",
			input:    "~",
			expected: home,
		},
		{
			name:     "relative path without tilde remains unchanged",
			input:    "data/payload.json",
			expected: "data/payload.json",
		},
		{
			name:      "tilde only returns error when HOME is unset",
			homeEnv:   func() *string { s := ""; return &s }(),
			input:     "~",
			expectErr: true,
		},
		{
			name:      "tilde slash returns error when HOME is unset",
			homeEnv:   func() *string { s := ""; return &s }(),
			input:     "~/.local/share/agystatusline/payload.json",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.homeEnv != nil {
				t.Setenv("HOME", *tt.homeEnv)
			}
			got, err := ExpandPath(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("ExpandPath(%q) expected error, got nil (result: %q)", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ExpandPath(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDumpPayload(t *testing.T) {
	t.Run("empty path does nothing and returns nil", func(t *testing.T) {
		err := DumpPayload("", []byte(`{"status":"ok"}`))
		if err != nil {
			t.Errorf("Expected nil error for empty path, got %v", err)
		}
	})

	t.Run("whitespace only path does nothing and returns nil", func(t *testing.T) {
		err := DumpPayload("   \t\n  ", []byte(`{"status":"ok"}`))
		if err != nil {
			t.Errorf("Expected nil error for whitespace path, got %v", err)
		}
	})

	t.Run("returns error and does not create literal tilde directory when HOME is unset", func(t *testing.T) {
		tempDir := t.TempDir()
		origWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get wd: %v", err)
		}
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to chdir: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chdir(origWd)
		})
		t.Setenv("HOME", "")

		err = DumpPayload("~/.local/share/agystatusline/payload.json", []byte(`{"status":"ok"}`))
		if err == nil {
			t.Errorf("Expected error when HOME is unset for tilde path, got nil")
		}
		if _, statErr := os.Stat(filepath.Join(tempDir, "~")); !os.IsNotExist(statErr) {
			t.Errorf("Expected literal '~' directory not to be created in CWD, but stat returned: %v", statErr)
		}
	})

	t.Run("successful verbatim write with 0600 file mode and 0700 parent dir creation", func(t *testing.T) {
		tempDir := t.TempDir()
		parentDir := filepath.Join(tempDir, "nested", "sub")
		targetPath := filepath.Join(parentDir, "payload.json")
		payload := []byte(`{"version":"1.0.0","quota":{"gemini-5h":{"remaining_fraction":0.85}}}`)

		err := DumpPayload(targetPath, payload)
		if err != nil {
			t.Fatalf("DumpPayload failed: %v", err)
		}

		dirInfo, err := os.Stat(parentDir)
		if err != nil {
			t.Fatalf("Failed to stat parent dir: %v", err)
		}
		if dirPerm := dirInfo.Mode().Perm(); dirPerm != 0700 {
			t.Errorf("Expected parent directory mode 0700, got %o", dirPerm)
		}

		info, err := os.Stat(targetPath)
		if err != nil {
			t.Fatalf("Failed to stat target file: %v", err)
		}

		// Check file permissions (0600)
		perm := info.Mode().Perm()
		if perm != 0600 {
			t.Errorf("Expected file mode 0600, got %o", perm)
		}

		// Check verbatim content
		content, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("Failed to read dumped file: %v", err)
		}
		if !bytes.Equal(content, payload) {
			t.Errorf("Content mismatch: expected %q, got %q", string(payload), string(content))
		}
	})

	t.Run("concurrent writes to the same target do not corrupt", func(t *testing.T) {
		tempDir := t.TempDir()
		targetPath := filepath.Join(tempDir, "concurrent_payload.json")

		const workers = 10
		var wg sync.WaitGroup
		errs := make(chan error, workers)

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				data := []byte(fmt.Sprintf(`{"worker":%d,"data":"verbatim-content-%d"}`, workerID, workerID))
				if err := DumpPayload(targetPath, data); err != nil {
					errs <- err
				}
			}(i)
		}

		wg.Wait()
		close(errs)

		for err := range errs {
			t.Errorf("Worker encountered error: %v", err)
		}

		// File should exist and be valid non-empty
		content, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("Failed to read file after concurrent writes: %v", err)
		}
		if len(content) == 0 {
			t.Errorf("File is empty after concurrent writes")
		}

		// Ensure no temp files remained in tempDir
		entries, err := os.ReadDir(tempDir)
		if err != nil {
			t.Fatalf("Failed to read tempDir: %v", err)
		}
		if len(entries) != 1 || entries[0].Name() != "concurrent_payload.json" {
			t.Errorf("Leftover temp files detected: %+v", entries)
		}
	})

	t.Run("returns error when target directory is not writable", func(t *testing.T) {
		tempDir := t.TempDir()
		readOnlyDir := filepath.Join(tempDir, "readonly")
		if err := os.Mkdir(readOnlyDir, 0500); err != nil {
			t.Fatalf("Failed to create readonly dir: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chmod(readOnlyDir, 0700)
		})

		targetPath := filepath.Join(readOnlyDir, "payload.json")
		err := DumpPayload(targetPath, []byte(`{"test":"error"}`))
		if err == nil {
			t.Errorf("Expected error when writing to readonly dir, got nil")
		}
	})
}
