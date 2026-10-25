package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath trims surrounding whitespace and expands leading ~ or ~/ to the user's home directory.
func ExpandPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return home, nil
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

// DumpPayload atomically writes data to the target path with 0600 file permissions
// and creates missing parent directories with 0700 permissions.
// If path is empty or whitespace-only, it returns nil immediately without performing any action.
func DumpPayload(path string, data []byte) error {
	resolvedPath, err := ExpandPath(path)
	if err != nil {
		return err
	}
	if resolvedPath == "" {
		return nil
	}

	dir := filepath.Dir(resolvedPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	base := filepath.Base(resolvedPath)
	tempFile, err := os.CreateTemp(dir, "."+base+".tmp-*")
	if err != nil {
		return err
	}

	tempPath := tempFile.Name()
	var success bool
	defer func() {
		_ = tempFile.Close()
		if !success {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.Write(data); err != nil {
		return err
	}
	if err := tempFile.Sync(); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tempPath, resolvedPath); err != nil {
		return err
	}

	success = true
	return nil
}
