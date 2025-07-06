package main

import (
	"dobrikov91/shubert/model"
	"os"
	"testing"
)

func TestController(t *testing.T) {
	os.Remove("test.json")

	config, err := model.NewConfig("test.json")
	if err != nil {
		t.Errorf("Cant create config %v", err)
	}

	c := Controller{
		EditMode:    false,
		Config:      config,
		midiDevices: nil,
		web:         nil,
		Port:        "0",
	}

	// empty config
	evt := model.Event{"kb", 1, 2, 0}
	if _, err := c.handleCommand(evt); err == nil {
		t.Errorf("Found non-existent command")
	}

	// add command
	if err := c.handleEdit(evt); err != nil {
		t.Errorf("Cant add command %v", err)
	}

	// add twice
	if err := c.handleEdit(evt); err == nil {
		t.Errorf("Double-add fine")
	}

	// add with another value
	evtPress := model.Event{"kb", 1, 2, 10}
	if err := c.handleEdit(evtPress); err == nil {
		t.Errorf("Double-add with another value fine")
	}

	// run command
	c.Config.Data.Commands[0].Command = "echo test"
	res, err := c.handleCommand(evtPress)
	if err != nil {
		t.Errorf("Cant run existent command %v", err)
	}
	if res != "test\n" {
		t.Errorf("Cant run command")
	}

	// add new command
	evt2 := model.Event{"kb", 11, 12, 0}
	if err := c.handleEdit(evt2); err != nil {
		t.Errorf("Cant add command")
	}

	// delete command
	c.Config.DeleteCommand(0)
	if _, err := c.handleCommand(evtPress); err == nil {
		t.Errorf("Found non-existent command")
	}

	// add command with $VALUE
	evt3 := model.Event{"kb", 21, 22, 20}
	if err := c.handleEdit(evt3); err != nil {
		t.Errorf("Cant add command")
	}
	c.Config.Data.Commands[1].Command = "echo $VALUE"
	res, err = c.handleCommand(evt3)
	if err != nil {
		t.Errorf("Cant add command with $VALUE")
	}
	if res != "20\n" {
		t.Errorf("Cant insert $VALUE")
	}

	// test $VALUE insert with another value
	evt4 := model.Event{"kb", 21, 22, 30}
	c.Config.Data.Commands[1].Command = "echo $VALUE"
	res, err = c.handleCommand(evt4)
	if err != nil {
		t.Errorf("Cant add command with $VALUE")
	}
	if res != "30\n" {
		t.Errorf("Cant insert another $VALUE")
	}
}
