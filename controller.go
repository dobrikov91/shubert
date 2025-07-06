package main

import (
	"dobrikov91/shubert/model"
	"fmt"
	"log"
)

type Controller struct {
	EditMode    bool
	Config      *model.Config
	midiDevices *model.MidiDevices
	web         *WebServer

	Port    string
	Version string
}

func NewController(path string, port string, version string) (*Controller, error) {
	c, err := model.NewConfig(path)
	if err != nil {
		return nil, err
	}

	d := model.NewMidiDevices()
	return &Controller{
		EditMode:    false,
		Config:      c,
		midiDevices: d,
		web:         nil, // add later
		Port:        port,
		Version:     version,
	}, nil
}

func (c *Controller) MainLoop(debug bool) {
	if !debug {
		// trick to remove error logs
		go c.midiDevices.ScanInputDevices()
	} else {
		fmt.Println("DEBUG MODE")
	}

	for {
		e := <-c.midiDevices.Event
		fmt.Println(e)

		if c.EditMode {
			err := c.handleEdit(e)
			if err != nil && debug {
				log.Print(err)
			}
		} else {
			output, err := c.handleCommand(e)
			if err != nil {
				log.Print(err)
			} else {
				log.Print(output)
			}
		}
	}
}

func (c *Controller) handleEdit(e model.Event) error {
	if id, err := c.Config.GetEventId(e); err == nil {
		if c.web != nil {
			c.web.broadcast <- model.Commands{HighlightId: id + 1, Commands: []model.Command{}}
		}
		return fmt.Errorf("event already present")
	}

	e.Value = 0
	c.Config.AddCommand(model.Command{
		Event:     e,
		Alias:     "",
		Trigger:   "OnPress",
		Command:   "",
		TimeoutMs: 0,
	})

	if c.web != nil {
		c.web.broadcast <- c.Config.Data
		c.web.broadcast <- model.Commands{HighlightId: len(c.Config.Data.Commands), Commands: []model.Command{}}
	}

	return nil
}

func (c *Controller) handleCommand(e model.Event) (string, error) {
	cmd, id, err := c.Config.GetCommandWithId(e)
	if err != nil {
		return "", fmt.Errorf("command not found %v", err)
	}
	cmd.Event.Value = e.Value

	// +1 is un ugly trick to avoid omitting field in json
	if c.web != nil {
		c.web.broadcast <- model.Commands{HighlightId: id + 1, Commands: []model.Command{}}
	}

	return model.RunCommand(cmd)
}
