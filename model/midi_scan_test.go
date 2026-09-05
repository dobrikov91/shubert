package model

import (
	"testing"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/testdrv"
)

// readyDriver/readyIn wrap a drivers.Driver to close a channel right after
// Listen() is set up. ScanInputDevices runs in its own goroutine, and
// listenToMidiDevice's callback later runs on whichever goroutine calls
// Send() on the paired virtual output port — a plain time.Sleep in the
// test would "work" but leaves no real happens-before edge for the race
// detector, so tests wait on this channel instead.
type readyDriver struct {
	drivers.Driver
	ready chan struct{}
}

func (r *readyDriver) Ins() ([]drivers.In, error) {
	ins, err := r.Driver.Ins()
	if err != nil {
		return ins, err
	}
	wrapped := make([]drivers.In, len(ins))
	for i, in := range ins {
		wrapped[i] = &readyIn{In: in, ready: r.ready}
	}
	return wrapped, nil
}

type readyIn struct {
	drivers.In
	ready chan struct{}
}

func (r *readyIn) Listen(onMsg func(msg []byte, ms int32), conf drivers.ListenConfig) (func(), error) {
	stop, err := r.In.Listen(onMsg, conf)
	if err == nil {
		select {
		case <-r.ready:
		default:
			close(r.ready)
		}
	}
	return stop, err
}

// newTestMidiDevices wires a MidiDevices up to a virtual (no hardware,
// no CGO) MIDI driver so message decoding can be tested deterministically
// in any environment, including CI runners with no MIDI hardware.
func newTestMidiDevices(t *testing.T) (*MidiDevices, drivers.Out) {
	t.Helper()

	drv := testdrv.New("virtual")
	ready := make(chan struct{})

	d := NewMidiDevices()
	d.newDriver = func() (drivers.Driver, error) {
		return &readyDriver{Driver: drv, ready: ready}, nil
	}

	go d.ScanInputDevices()
	t.Cleanup(func() {
		d.Quit <- true
	})

	select {
	case <-ready:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for ScanInputDevices to start listening")
	}

	outs, err := drv.Outs()
	if err != nil || len(outs) == 0 {
		t.Fatalf("no virtual output ports: %v", err)
	}

	out := outs[0]
	if err := out.Open(); err != nil {
		t.Fatalf("can't open virtual out port: %v", err)
	}

	return d, out
}

// sendMessage sends via a goroutine: Send() blocks until listenToMidiDevice's
// callback finishes, which for a recognized message means until someone
// reads from d.Event.
func sendMessage(t *testing.T, out drivers.Out, msg midi.Message) {
	t.Helper()
	go func() {
		if err := out.Send(msg.Bytes()); err != nil {
			t.Errorf("send failed: %v", err)
		}
	}()
}

func recvEvent(t *testing.T, d *MidiDevices) Event {
	t.Helper()
	select {
	case evt := <-d.Event:
		return evt
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for MIDI event")
		return Event{}
	}
}

func TestListenToMidiDeviceNoteOn(t *testing.T) {
	d, out := newTestMidiDevices(t)

	sendMessage(t, out, midi.NoteOn(1, 60, 100))

	evt := recvEvent(t, d)
	if evt.Channel != 1 || evt.Key != 60 || evt.Value != 100 {
		t.Errorf("unexpected event: %+v", evt)
	}
}

func TestListenToMidiDeviceNoteOff(t *testing.T) {
	d, out := newTestMidiDevices(t)

	sendMessage(t, out, midi.NoteOff(2, 40))

	evt := recvEvent(t, d)
	if evt.Channel != 2 || evt.Key != 40 || evt.Value != 0 {
		t.Errorf("unexpected event: %+v", evt)
	}
}

func TestListenToMidiDeviceControlChange(t *testing.T) {
	d, out := newTestMidiDevices(t)

	sendMessage(t, out, midi.ControlChange(0, 7, 127))

	evt := recvEvent(t, d)
	if evt.Channel != 0 || evt.Key != 7 || evt.Value != 127 {
		t.Errorf("unexpected event: %+v", evt)
	}
}

func TestListenToMidiDeviceIgnoresUnrecognizedMessages(t *testing.T) {
	d, out := newTestMidiDevices(t)

	// SysEx isn't NoteOn/NoteOff/ControlChange, so it must be silently
	// dropped by the callback rather than surfacing as an Event or
	// blocking the pipeline for the NoteOn that follows it.
	if err := out.Send(midi.SysEx([]byte{0x01, 0x02}).Bytes()); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	sendMessage(t, out, midi.NoteOn(3, 10, 20))

	evt := recvEvent(t, d)
	if evt.Channel != 3 || evt.Key != 10 || evt.Value != 20 {
		t.Errorf("expected the NoteOn to come through, got: %+v", evt)
	}
}
