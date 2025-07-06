//go:build !windows

package model

import (
	"os/exec"
	"syscall"
	"time"
)

func RunCommandArgs(args []string, timeoutMs int) (string, error) {
	shellCommand := exec.Command(args[0], args[1:]...)
	shellCommand.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmdDone := make(chan cmdResult, 1)
	go func() {
		outb, err := shellCommand.CombinedOutput()
		cmdDone <- cmdResult{outb, err}
	}()

	if timeoutMs == 0 {
		res := <-cmdDone
		return string(res.outb), res.err
	}

	select {
	case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
		syscall.Kill(-shellCommand.Process.Pid, syscall.SIGKILL)
		return "", &CmdTimeoutError{}
	case res := <-cmdDone:
		return string(res.outb), res.err
	}
}
