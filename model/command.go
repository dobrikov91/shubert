package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mattn/go-shellwords"
)

type cmdResult struct {
	outb []byte
	err  error
}

type CmdTimeoutError struct{}

func (m *CmdTimeoutError) Error() string {
	return "command timeout"
}

func InsertValue(cmd string, value int) string {
	return strings.ReplaceAll(cmd, "$VALUE", strconv.Itoa(value))
}

// https://jarv.org/posts/command-with-timeout/
func RunCommand(cmd Command) (string, error) {
	if cmd.Command == "" {
		return "", fmt.Errorf("empty command")
	}

	filledCmd := InsertValue(cmd.Command, cmd.Event.Value)

	args, err := shellwords.Parse(filledCmd)
	if err != nil {
		return "", fmt.Errorf("command parse error %v", err)
	}

	return RunCommandArgs(args, cmd.TimeoutMs)
}
