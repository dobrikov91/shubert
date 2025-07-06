//go:build windows

package model

import (
	"os/exec"
)

// no timeout support yet
func RunCommandArgs(args []string, timeoutMs int) (string, error) {
	shellCommand := exec.Command(args[0], args[1:]...)

	cmdDone := make(chan cmdResult, 1)
	go func() {
		outb, err := shellCommand.CombinedOutput()
		cmdDone <- cmdResult{outb, err}
	}()

	res := <-cmdDone
	return string(res.outb), res.err
}
