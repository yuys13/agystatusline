package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestSaveSettings_Success(t *testing.T) {
	settings := types.DefaultSettings()
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "config.toml")

	err := saveSettings(targetPath, settings)
	if err != nil {
		t.Fatalf("saveSettings() failed unexpectedly: %v", err)
	}

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Errorf("Expected settings file to exist at %q", targetPath)
	}

	// Verify temporary files are cleaned up
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read tempDir: %v", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".tmp") {
			t.Errorf("Expected no temporary files remaining in %q, but found %q", tempDir, f.Name())
		}
	}
}

func TestSaveSettings_ErrorPaths(t *testing.T) {
	settings := types.DefaultSettings()

	t.Run("Target path is directory error handling", func(t *testing.T) {
		tempDir := t.TempDir()
		invalidPath := filepath.Join(tempDir, "a_dir")
		if err := os.Mkdir(invalidPath, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		err := saveSettings(invalidPath, settings)
		if err == nil {
			t.Errorf("Expected saveSettings to fail when target path is a directory")
		}
	})

	t.Run("MkdirAll error simulation", func(t *testing.T) {
		tempDir := t.TempDir()
		parentFile := filepath.Join(tempDir, "file_blocking_mkdir")
		if err := os.WriteFile(parentFile, []byte("data"), 0644); err != nil {
			t.Fatalf("Failed to create file blocking mkdir: %v", err)
		}
		targetPath := filepath.Join(parentFile, "config.toml")

		err := saveSettings(targetPath, settings)
		if err == nil {
			t.Errorf("Expected saveSettings to fail when parent path is a regular file")
		}
	})

	t.Run("OpenFile error simulation and cleanup execution", func(t *testing.T) {
		tempDir := t.TempDir()
		readOnlyDir := filepath.Join(tempDir, "read_only_dir")
		if err := os.Mkdir(readOnlyDir, 0555); err != nil {
			t.Fatalf("Failed to create read only dir: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chmod(readOnlyDir, 0755)
		})

		targetPath := filepath.Join(readOnlyDir, "config.toml")
		err := saveSettings(targetPath, settings)
		if err == nil {
			t.Errorf("Expected saveSettings to fail when target directory is read-only")
		}
	})
}
