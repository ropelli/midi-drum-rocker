package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path"
	"reflect"
	"testing"
)

type testableNotePlayer struct {
	sequence []uint8
}

const SETUP = 99

func (p *testableNotePlayer) clear() {
	p.sequence = []uint8{}
}

func (p *testableNotePlayer) Setup() error {
	p.sequence = append(p.sequence, SETUP)
	return nil
}

func (p *testableNotePlayer) SendNote(note, vel uint8, on bool) {
	p.sequence = append(p.sequence, note)
}

func (p *testableNotePlayer) assertSequence(t *testing.T, expected []uint8, message ...string) {
	if !reflect.DeepEqual(p.sequence, expected) {
		info := ""
		if len(message) == 1 {
			info = info + message[0]
		} else if len(message) > 1 {
			t.Fatalf("assert message can have one message, got %s", message)
		}
		t.Errorf("sequence %v is not %v %s", p.sequence, expected, info)
	}
}

func TestReplayOnEmptyFileRunsFine(t *testing.T) {
	np := &testableNotePlayer{}
	settings := &Settings{
		"js",
		"midi",
		NO_RETRY,
	}
	dir := t.TempDir()
	filePath := path.Join(dir, "foobar")
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	replay(settings, filePath, true, np)
	np.assertSequence(t, []uint8{SETUP})

}

func TestReplayOnZeroDurationRunsFine(t *testing.T) {
	np := &testableNotePlayer{}
	settings := &Settings{
		"js",
		"midi",
		NO_RETRY,
	}
	dir := t.TempDir()
	filePath := path.Join(dir, "foobar")
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("duration: 0")
	f.Close()
	replay(settings, filePath, true, np)
	np.assertSequence(t, []uint8{SETUP})
}

func TestReplayWithSnareSendsCorrectNote(t *testing.T) {
	np := &testableNotePlayer{}
	settings := &Settings{
		"js",
		"midi",
		NO_RETRY,
	}
	// Kick not connected
	filePath, f, err := recreateFile(t)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("buttons: 00000000000000000000000000000000\n")
	f.WriteString("buttons: 00000000000000000000010000000000\n")
	f.WriteString("buttons: 00000000000000000000010000000010\n")
	f.WriteString("buttons: 00000000000000000000000000000010\n")
	f.WriteString("buttons: 00000000000000000000000000000000\n")
	f.Close()
	replay(settings, filePath, true, np)
	np.assertSequence(t, []uint8{SETUP, NOTE_SNARE}, "when kick not connected")
	np.clear()
	// Kick connected
	filePath, f, err = recreateFile(t)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("buttons: 00000000000000000000001000000000\n")
	f.WriteString("buttons: 00000000000000000000011000000000\n")
	f.WriteString("buttons: 00000000000000000000011000000010\n")
	f.WriteString("buttons: 00000000000000000000001000000010\n")
	f.WriteString("buttons: 00000000000000000000001000000000\n")
	f.Close()
	replay(settings, filePath, true, np)
	np.assertSequence(t, []uint8{SETUP, NOTE_SNARE}, "when kick connected")
	np.clear()
	// Kick held
	filePath, f, err = recreateFile(t)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("buttons: 00000000000000000000001000000000\n")
	f.WriteString("buttons: 00000000000000000000001000010000\n")
	f.WriteString("buttons: 00000000000000000000011000010000\n")
	f.WriteString("buttons: 00000000000000000000011000010010\n")
	f.WriteString("buttons: 00000000000000000000001000010010\n")
	f.WriteString("buttons: 00000000000000000000001000010000\n")
	f.WriteString("buttons: 00000000000000000000001000000000\n")
	f.Close()
	replay(settings, filePath, true, np)
	np.assertSequence(t, []uint8{SETUP, NOTE_KICK, NOTE_SNARE}, "when kick held")
	np.clear()
}

func TestReplayWithAllNotes(t *testing.T) {
	np := &testableNotePlayer{}
	settings := &Settings{
		"js",
		"midi",
		NO_RETRY,
	}
	filePath := "all.txt"
	replay(settings, filePath, true, np)
	np.assertSequence(t, []uint8{SETUP, NOTE_HIHAT, NOTE_RIDE, NOTE_CRASH, NOTE_SNARE, NOTE_TOM1, NOTE_TOM2, NOTE_TOM3, NOTE_KICK})
	np.clear()
	filePath = "all-pedal-down.txt"
	replay(settings, filePath, true, np)
	np.assertSequence(t, []uint8{SETUP, NOTE_KICK, NOTE_SNARE, NOTE_TOM1, NOTE_TOM2, NOTE_TOM3, NOTE_HIHAT, NOTE_RIDE, NOTE_CRASH, NOTE_KICK})
	np.clear()
	filePath = "pedal-to-the-metal.txt"
	replay(settings, filePath, true, np)
	np.assertSequence(t, []uint8{SETUP, NOTE_KICK, NOTE_SNARE, NOTE_TOM1, NOTE_TOM2, NOTE_TOM3, NOTE_HIHAT, NOTE_RIDE, NOTE_CRASH, NOTE_KICK})
}

func recreateFile(t *testing.T) (string, *os.File, error) {
	dir := t.TempDir()
	randomId := rand.IntN(10000)
	filePath := path.Join(dir, fmt.Sprintf("test-%v", randomId))
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	return filePath, f, err
}
