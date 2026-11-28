package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/yuys13/agystatusline/types"
)

func saveSettings(path string, settings types.Settings) error {
	settings.ApplyDefaults()
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	tempPath := fmt.Sprintf("%s.%d.%d.tmp", path, os.Getpid(), time.Now().UnixNano())
	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	var success bool
	defer func() {
		_ = file.Close()
		if !success {
			_ = os.Remove(tempPath)
		}
	}()

	bytes, err := toml.Marshal(settings)
	if err != nil {
		return err
	}

	_, err = file.Write(bytes)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	success = true
	return nil
}
