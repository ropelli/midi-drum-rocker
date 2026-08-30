package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path"
	"reflect"
	"testing"
	"time"
)

type testableNotePlayer struct {
	sequence []MidiNote
}

const SETUP = MidiNote(99)

func (p *testableNotePlayer) clear() {
	p.sequence = []MidiNote{}
}

func (p *testableNotePlayer) Setup() error {
	p.sequence = append(p.sequence, SETUP)
	return nil
}

func (p *testableNotePlayer) SendNote(note MidiNote, vel uint8, on bool) {
	p.sequence = append(p.sequence, note)
}

func (p *testableNotePlayer) Sleep(duration time.Duration) {
	lastIndex := len(p.sequence) - 1
	lastNote := p.sequence[lastIndex]
	if lastNote >= 1000 {
		p.sequence[lastIndex] = MidiNote(int(lastNote) + int(duration.Milliseconds()) + 1000)
		return
	}
	p.sequence = append(p.sequence, MidiNote(duration.Milliseconds()+1000)) // +200 to avoid confusion with actual notes
}

func (p *testableNotePlayer) assertSequence(t *testing.T, expected []MidiNote, message ...string) {
	if !reflect.DeepEqual(p.sequence, expected) {
		info := ""
		if len(message) == 1 {
			info = info + message[0]
		} else if len(message) > 1 {
			t.Fatalf("assert message can have one message, got %s", message)
		}
		t.Errorf("sequences differ\nactual:   %v\nexpected: %v\n%s", p.sequence, expected, info)
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
	np.assertSequence(t, []MidiNote{SETUP})

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
	np.assertSequence(t, []MidiNote{SETUP})
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
	np.assertSequence(t, []MidiNote{SETUP, NOTE_SNARE}, "when kick not connected")
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
	np.assertSequence(t, []MidiNote{SETUP, NOTE_SNARE}, "when kick connected")
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
	np.assertSequence(t, []MidiNote{SETUP, NOTE_KICK, NOTE_SNARE}, "when kick held")
	np.clear()
}

func TestReplayWithAllNotes(t *testing.T) {
	np := &testableNotePlayer{}
	settings := &Settings{
		"js",
		"midi",
		NO_RETRY,
	}
	filePath := "test_data/all.txt"
	replay(settings, filePath, true, np)
	np.assertSequence(t, []MidiNote{SETUP, NOTE_HIHAT, NOTE_RIDE, NOTE_CRASH, NOTE_SNARE, NOTE_TOM1, NOTE_TOM2, NOTE_TOM3, NOTE_KICK})
	np.clear()
	filePath = "test_data/all-pedal-down.txt"
	replay(settings, filePath, true, np)
	np.assertSequence(t, []MidiNote{SETUP, NOTE_KICK, NOTE_SNARE, NOTE_TOM1, NOTE_TOM2, NOTE_TOM3, NOTE_HIHAT, NOTE_RIDE, NOTE_CRASH, NOTE_KICK})
	np.clear()
	filePath = "test_data/pedal-to-the-metal.txt"
	replay(settings, filePath, true, np)
	np.assertSequence(t, []MidiNote{SETUP, NOTE_KICK, NOTE_SNARE, NOTE_TOM1, NOTE_TOM2, NOTE_TOM3, NOTE_HIHAT, NOTE_RIDE, NOTE_CRASH, NOTE_KICK})
}

func TestReplayWithHihatAndRideCombo(t *testing.T) {
	np := &testableNotePlayer{}
	settings := &Settings{
		"js",
		"midi",
		NO_RETRY,
	}
	filePath := "test_data/hihat-ride.txt"
	replay(settings, filePath, true, np)
	np.assertSequence(t, []MidiNote{SETUP, NOTE_HIHAT, NOTE_RIDE})
}

func TestReplayWithCombos(t *testing.T) {
	np := &testableNotePlayer{}
	settings := &Settings{
		"js",
		"midi",
		NO_RETRY,
	}
	filePath := "test_data/combos.txt"
	replay(settings, filePath, false, np)
	np.assertSequence(t, []MidiNote{SETUP, 3956, NOTE_SNARE, NOTE_TOM1, 4536, NOTE_SNARE, NOTE_TOM2, 3412, NOTE_TOM3, NOTE_SNARE, 4728, NOTE_TOM2, NOTE_TOM1, 4736, NOTE_TOM1, NOTE_TOM3, 6192, NOTE_TOM3, NOTE_TOM2, 5608, NOTE_HIHAT, NOTE_SNARE, 4272, NOTE_HIHAT, NOTE_TOM1, 4296, NOTE_HIHAT, NOTE_TOM2, 4248, NOTE_HIHAT, NOTE_TOM3, 9744, NOTE_RIDE, NOTE_SNARE, 4156, NOTE_RIDE, NOTE_TOM1, 4108, NOTE_RIDE, NOTE_TOM2, 4216, NOTE_RIDE, NOTE_TOM3, 5376, NOTE_CRASH, NOTE_SNARE, 3064, NOTE_CRASH, NOTE_TOM1, 3124, NOTE_CRASH, NOTE_TOM2, 4012, NOTE_CRASH, NOTE_TOM3, 8980, NOTE_RIDE, NOTE_HIHAT, 3396, NOTE_CRASH, NOTE_HIHAT, 4383, NOTE_HIHAT, NOTE_RIDE})
	np.clear()
	filePath = "test_data/combos-with-pedal.txt"
	replay(settings, filePath, false, np)
	np.assertSequence(t, []MidiNote{SETUP, 2859, NOTE_KICK, 4960, NOTE_SNARE, 4680, NOTE_SNARE, NOTE_TOM1, 4656, NOTE_SNARE, NOTE_TOM2, 4591, NOTE_TOM3, NOTE_SNARE, 2740, NOTE_TOM2, NOTE_TOM1, 2596, NOTE_TOM3, NOTE_TOM2, 6397, NOTE_TOM1, NOTE_TOM3, 4523, NOTE_HIHAT, NOTE_SNARE, 3456, NOTE_HIHAT, NOTE_TOM1, 3452, NOTE_HIHAT, NOTE_TOM2, 2456, NOTE_HIHAT, NOTE_TOM3, 4772, NOTE_RIDE, NOTE_SNARE, 3444, NOTE_RIDE, NOTE_TOM1, 2432, NOTE_RIDE, NOTE_TOM2, 3452, NOTE_RIDE, NOTE_TOM3, 3112, NOTE_CRASH, NOTE_SNARE, 3460, NOTE_CRASH, NOTE_TOM1, 2440, NOTE_CRASH, NOTE_TOM2, 2460, NOTE_CRASH, NOTE_TOM3, 6612, NOTE_HIHAT, NOTE_RIDE, 3600, NOTE_CRASH, NOTE_HIHAT, 4064, NOTE_RIDE, NOTE_CRASH, 4184})
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
