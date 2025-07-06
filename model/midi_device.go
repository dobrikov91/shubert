package model

import (
	"log"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

type MidiDevices struct {
	Event              chan Event
	Quit               chan bool
	DeviceStopFuncList []func()
}

func NewMidiDevices() *MidiDevices {
	return &MidiDevices{
		make(chan Event),
		make(chan bool),
		nil,
	}
}

func (m *MidiDevices) ScanInputDevices() {
	var driver *rtmididrv.Driver

	driver, err := rtmididrv.New()
	if err != nil {
		log.Fatalf("Failed to create MIDI driver: %v", err)
	}
	defer driver.Close()

	var prevInputs []string

	for {
		// none of provided drivers can't update device list on-the-fly
		err := driver.Close()
		if err != nil {
			log.Printf("Failed to close midi driver: %v", err)
		}

		driver, err := rtmididrv.New()
		if err != nil {
			log.Fatalf("Failed to create MIDI driver: %v", err)
		}

		inputs, err := driver.Ins()
		if err != nil {
			log.Printf("Failed to get input devices: %v", err)
			continue
		}

		currentInputs := deviceNames(inputs)

		if len(currentInputs) != len(prevInputs) || !slicesEqual(currentInputs, prevInputs) {
			log.Println("Input devices changed:")
			log.Printf("Previous: %v\n", prevInputs)
			log.Printf("Current: %v\n", currentInputs)
			prevInputs = currentInputs

			m.stopAllListeners()

			for _, input := range inputs {
				stopFunc, err := m.listenToMidiDevice(input)
				if err != nil {
					log.Printf("Can't start listening %s", input.String())
					continue
				}
				m.DeviceStopFuncList = append(m.DeviceStopFuncList, stopFunc)
			}
		}

		select {
		case msg := <-m.Quit:
			if msg {
				m.stopAllListeners()
				return
			}
		default:
			time.Sleep(300 * time.Millisecond)
		}
	}
}

func (m *MidiDevices) stopAllListeners() {
	for i, stop := range m.DeviceStopFuncList {
		log.Printf("Stop listeners %d", i)
		stop()
	}

	m.DeviceStopFuncList = nil
}

func (m *MidiDevices) listenToMidiDevice(input drivers.In) (func(), error) {
	return midi.ListenTo(input, func(msg midi.Message, timestampms int32) {
		var channel, key, value uint8
		switch {
		case msg.GetNoteOn(&channel, &key, &value) ||
			msg.GetNoteOff(&channel, &key, &value) ||
			msg.GetControlChange(&channel, &key, &value):
			break
		default:
			log.Printf("Can't parse midi event[%s]: %v\n", input.String(), msg)
			return
		}

		event := Event{
			Device:  input.String(),
			Channel: int(channel),
			Key:     int(key),
			Value:   int(value),
		}
		m.Event <- event
	}, midi.UseSysEx())
}

func deviceNames(devices []drivers.In) []string {
	names := make([]string, len(devices))
	for i, device := range devices {
		names[i] = device.String()
	}
	return names
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
