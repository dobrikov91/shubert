package model

import (
	"os"
	"testing"
	"time"
)

func TestStopScan(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("skipping: opens a real ALSA MIDI connection, which hangs on CI runners with no MIDI hardware")
	}

	d := NewMidiDevices()
	go d.ScanInputDevices()

	time.Sleep(time.Second)
	d.Quit <- true

	time.Sleep(time.Second)
}
