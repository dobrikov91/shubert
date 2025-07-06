package main

import (
	"dobrikov91/shubert/model"
	"fmt"
	"strconv"
)

func MidiMock(c *Controller) {
	// manual testing without midi keyboard
	var input string
	for {
		fmt.Scanln(&input)

		ival, err := strconv.Atoi(input)
		if err != nil {
			continue
		}

		evt := model.Event{
			Device:  "kb",
			Channel: 0,
			Key:     ival,
			Value:   ival * 10,
		}
		fmt.Println(evt)

		c.midiDevices.Event <- evt
	}
}
