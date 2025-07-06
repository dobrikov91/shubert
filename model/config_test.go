package model

import (
	"testing"
)

func TestConfig(t *testing.T) {
	c := &Config{}

	c.AddCommand(Command{Event{"midi1", 1, 2, 3}, "Release", "OnRelease", "Cmd1", 0})
	if len(c.Data.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(c.Data.Commands))
	}

	c.AddCommand(Command{Event{"midi1", 11, 12, 13}, "Change", "OnChange", "Cmd2", 0})
	if len(c.Data.Commands) != 2 {
		t.Errorf("Expected 2 command, got %d", len(c.Data.Commands))
	}

	c.AddCommand(Command{Event{"midi1", 21, 22, 23}, "Press", "OnPress", "Cmd3", 0})
	if len(c.Data.Commands) != 3 {
		t.Errorf("Expected 3 command, got %d", len(c.Data.Commands))
	}

	// test id search
	id, err := c.GetEventId(Event{"midi1", 11, 12, 0})
	if err != nil || id != 1 {
		t.Errorf("Can't find cmd %v", err)
	}

	_, err = c.GetEventId(Event{"midi1", 0, 12, 0})
	if err == nil {
		t.Errorf("Found non-existent command")
	}

	// test cmds

	// 1. OnRelease
	// release => ok
	command, _, err := c.GetCommandWithId(Event{"midi1", 1, 2, 0})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if command.Command != "Cmd1" {
		t.Errorf("Expected Cmd1, got %v", command)
	}
	// press => not found
	_, _, err = c.GetCommandWithId(Event{"midi1", 1, 2, 10})
	if err == nil {
		t.Errorf("Expected error, press signal for OnRelease command")
	}

	// 2. OnChange
	// release => ok
	command, _, err = c.GetCommandWithId(Event{"midi1", 11, 12, 0})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if command.Command != "Cmd2" {
		t.Errorf("Expected Cmd2, got %v", command)
	}

	// press => ok
	command, _, err = c.GetCommandWithId(Event{"midi1", 11, 12, 10})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if command.Command != "Cmd2" {
		t.Errorf("Expected Cmd2, got %v", command)
	}

	// 3. OnPress
	// release => not found
	_, _, err = c.GetCommandWithId(Event{"midi1", 21, 22, 0})
	if err == nil {
		t.Errorf("Expected error, release signal for OnPress command")
	}

	// press => ok
	command, _, err = c.GetCommandWithId(Event{"midi1", 21, 22, 10})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if command.Command != "Cmd3" {
		t.Errorf("Expected Cmd3, got %v", command)
	}

	// non-existent event
	_, _, err = c.GetCommandWithId(Event{"midi1", 99, 0, 0})
	if err == nil {
		t.Errorf("Expected error, event not exist")
	}

	c.DeleteCommand(0)
	if len(c.Data.Commands) != 2 {
		t.Errorf("Expected 2 command, got %d", len(c.Data.Commands))
	}
}

func TestConfigSaveLoad(t *testing.T) {
	c := &Config{}
	c.FilePath = "test.json"

	c.AddCommand(Command{Event{"midi1", 1, 2, 3}, "Release", "OnRelease", "Cmd1", 0})
	if len(c.Data.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(c.Data.Commands))
	}

	c.AddCommand(Command{Event{"midi1", 11, 12, 13}, "Change", "OnChange", "Cmd2", 0})
	if len(c.Data.Commands) != 2 {
		t.Errorf("Expected 2 command, got %d", len(c.Data.Commands))
	}

	err := c.Save()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	c2, err := NewConfig("test.json")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(c2.Data.Commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(c2.Data.Commands))
	}
}
