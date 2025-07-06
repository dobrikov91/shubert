package model

import (
	"errors"
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	cmd := "Here is $VALUE"
	res := InsertValue(cmd, 10)

	if res != "Here is 10" {
		t.Errorf("Replace wrong, got %s", res)
	}
}

func TestRun(t *testing.T) {
	_, err := RunCommand(Command{Command: "", TimeoutMs: 0})
	if err == nil {
		t.Errorf("Empty command works, but shouldn't")
	}

	_, err = RunCommand(Command{Command: "ls", TimeoutMs: 0})
	if err != nil {
		t.Errorf("Command without args failed %v", err)
	}

	res, err := RunCommand(Command{Command: "echo \"Hi\"", TimeoutMs: 0})
	if err != nil {
		t.Errorf("Echo failed with %v", err)
	}

	if res != "Hi\n" {
		t.Errorf("Wrong response from Hi: %s", res)
	}

	res, err = RunCommand(Command{Command: "blabla", TimeoutMs: 0})
	if err == nil {
		t.Errorf("Can run unknown command, result: %s", res)
	}
}

func TestTimeout(t *testing.T) {
	target := &CmdTimeoutError{}

	start := time.Now()
	_, err := RunCommand(Command{Command: "sleep 10", TimeoutMs: 1000})
	if !errors.As(err, &target) {
		t.Errorf("Timeout doesn't work, %v", err)
	}

	elapsed := time.Since(start)
	if elapsed.Milliseconds() > 1100 {
		t.Errorf("Timeout should be 1000ms, real %d", elapsed.Milliseconds())
	}

	_, err = RunCommand(Command{Command: "echo no timeout", TimeoutMs: 1000})
	if errors.As(err, &target) {
		t.Errorf("Echo got timeout, %v", err)
	}
}
