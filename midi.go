package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

var drv *rtmididrv.Driver
var out drivers.Out

type MidiNote int

const (
	NOTE_KICK  = MidiNote(36)
	NOTE_SNARE = MidiNote(38)
	NOTE_HIHAT = MidiNote(42)
	NOTE_TOM1  = MidiNote(47)
	NOTE_TOM2  = MidiNote(48)
	NOTE_TOM3  = MidiNote(41)
	NOTE_CRASH = MidiNote(49)
	NOTE_RIDE  = MidiNote(50)
)

func (n MidiNote) String() string {
	switch n {
	case NOTE_KICK:
		return "Kick"
	case NOTE_SNARE:
		return "Snare"
	case NOTE_HIHAT:
		return "HiHat"
	case NOTE_TOM1:
		return "Tom1"
	case NOTE_TOM2:
		return "Tom2"
	case NOTE_TOM3:
		return "Tom3"
	case NOTE_CRASH:
		return "Crash"
	case NOTE_RIDE:
		return "Ride"
	default:
		return fmt.Sprintf("%d", n)
	}
}

type MidiPlayer struct {
	DeviceName string
}

var _ NotePlayer = (*MidiPlayer)(nil)

func NewMidiPlayer(s *Settings) *MidiPlayer {
	mp := MidiPlayer{}
	mp.DeviceName = s.midiName
	return &mp
}

func (p *MidiPlayer) Setup() error {
	var err error
	drv, err = rtmididrv.New()
	if err != nil {
		return errors.New(fmt.Sprintf("could not open ALSA MIDI driver: %v", err))
	}
	// Create a virtual MIDI output port
	out, err = drv.OpenVirtualOut(p.DeviceName)
	if err != nil {
		return errors.New(fmt.Sprintf("could not open virtual MIDI out: %v", err))
	}
	slog.Info("Virtual MIDI port created", "name", out.String())
	return err
}

func (p *MidiPlayer) Sleep(duration time.Duration) {
	slog.Debug("sleeping", "ms", duration)
	time.Sleep(duration)
}

func (p *MidiPlayer) SendNote(note MidiNote, vel uint8, on bool) {
	var msg midi.Message
	if on {
		msg = midi.NoteOn(0, uint8(note), vel)
	} else {
		msg = midi.NoteOff(0, uint8(note))
	}
	err := out.Send(msg)
	if err != nil {
		slog.Error("when seding midi message, got", "error", err)
	}
}

func (p *MidiPlayer) TestMidi() {
	err := p.Setup()
	if err != nil {
		log.Fatalln("FATAL ", err)
	}
	defer func() {
		out.Close()
		drv.Close()
	}()

	fmt.Println("Sending test snare hits every 1s...")
	for {
		p.SendNote(38, 100, true) // snare on
		time.Sleep(1000 * time.Millisecond)
		p.SendNote(38, 0, false) // snare off
		time.Sleep(1 * time.Second)
	}
}
