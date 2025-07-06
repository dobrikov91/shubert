package model

import (
	"testing"
	"time"
)

func TestStopScan(t *testing.T) {
	d := NewMidiDevices()
	go d.ScanInputDevices()

	time.Sleep(time.Second)
	d.Quit <- true

	time.Sleep(time.Second)
}
