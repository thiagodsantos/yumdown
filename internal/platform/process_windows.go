//go:build windows

package platform

import (
	"os"
	"os/exec"
	"strconv"
)

func ConfigureProcessTree(*exec.Cmd) {}

func TerminateProcessTree(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return os.ErrProcessDone
	}
	kill := exec.Command("taskkill.exe", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	if err := kill.Run(); err != nil {
		return cmd.Process.Kill()
	}
	return nil
}
