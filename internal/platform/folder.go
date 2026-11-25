package platform

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func OpenFolder(path string) error {
	if path == "" {
		return errors.New("folder path is empty")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve folder: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("open folder: %w", err)
	}
	if !info.IsDir() {
		return errors.New("path is not a folder")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer.exe", abs)
	case "darwin":
		cmd = exec.Command("open", abs)
	case "linux":
		cmd = exec.Command("xdg-open", abs)
	default:
		return fmt.Errorf("opening folders is unsupported on %s", runtime.GOOS)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start file manager: %w", err)
	}
	return cmd.Process.Release()
}
