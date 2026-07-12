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

const (
	NOTE_KICK  = 36
	NOTE_SNARE = 38
	NOTE_HIHAT = 42
	NOTE_TOM1  = 47
	NOTE_TOM2  = 48
	NOTE_TOM3  = 41
	NOTE_CRASH = 49
	NOTE_RIDE  = 50
)

func initMIDI(deviceName string) error {
	var err error
	drv, err = rtmididrv.New()
	if err != nil {
		return errors.New(fmt.Sprintf("could not open ALSA MIDI driver: %v", err))
	}
	// Create a virtual MIDI output port
	out, err = drv.OpenVirtualOut(deviceName)
	if err != nil {
		return errors.New(fmt.Sprintf("could not open virtual MIDI out: %v", err))
	}
	slog.Info("Virtual MIDI port created", "name", out.String())
	return err
}

func sendNote(note, vel uint8, on bool) {
	var msg midi.Message
	if on {
		msg = midi.NoteOn(0, note, vel)
	} else {
		msg = midi.NoteOff(0, note)
	}
	err := out.Send(msg)
	if err != nil {
		log.Printf("send error: %v", err)
	}
}

func testMidi(settings *Settings) {
	err := initMIDI(settings.midiName)
	if err != nil {
		log.Fatalf("FATAL %v", err)
	}
	defer func() {
		out.Close()
		drv.Close()
	}()

	fmt.Println("Sending test snare hits every 1s...")
	for {
		sendNote(38, 100, true) // snare on
		time.Sleep(1000 * time.Millisecond)
		sendNote(38, 0, false) // snare off
		time.Sleep(1 * time.Second)
	}
}
