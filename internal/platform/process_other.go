//go:build !windows && !linux && !darwin

package platform

import (
	"os"
	"os/exec"
)

func ConfigureProcessTree(*exec.Cmd) {}

func TerminateProcessTree(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return os.ErrProcessDone
	}
	return cmd.Process.Kill()
}
