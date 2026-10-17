package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/0xcafed00d/joystick"
)

const (
	GREEN_MASK             uint32 = 0b0000_0000_0000_0001
	RED_MASK               uint32 = 0b0000_0000_0000_0010
	BLUE_MASK              uint32 = 0b0000_0000_0000_0100
	YELLOW_MASK            uint32 = 0b0000_0000_0000_1000
	KICK_MASK              uint32 = 0b0000_0000_0001_0000
	CYMBAL_MASK            uint32 = 0b0000_0000_0010_0000
	KICK_CONNECTED_MASK    uint32 = 0b0000_0010_0000_0000
	TOM_MASK               uint32 = 0b0000_0100_0000_0000
	UNKNOWN_MASK           uint32 = 0b0000_1000_0000_0000
	UNKNOWN_MASK2          uint32 = 0b0001_0000_0000_0000
	YELLOW_CYMBAL_MOD_MASK uint32 = 0b0010_0000_0000_0000
	BLUE_CYMBAL_MOD_MASK   uint32 = 0b0100_0000_0000_0000
)

const (
	FULL_VOLUME = 100
)

const (
	SNARE_COMBO  = RED_MASK | TOM_MASK
	TOM1_COMBO   = YELLOW_MASK | TOM_MASK
	TOM2_COMBO   = BLUE_MASK | TOM_MASK
	TOM3_COMBO   = GREEN_MASK | TOM_MASK
	HIHAT_COMBO  = YELLOW_MASK | UNKNOWN_MASK | CYMBAL_MASK
	HIHAT2_COMBO = YELLOW_MASK | YELLOW_CYMBAL_MOD_MASK | CYMBAL_MASK
	RIDE_COMBO   = BLUE_MASK | UNKNOWN_MASK2 | CYMBAL_MASK
	RIDE2_COMBO  = BLUE_MASK | BLUE_CYMBAL_MOD_MASK | CYMBAL_MASK
	CRASH_COMBO  = GREEN_MASK | CYMBAL_MASK

	SNARE_TOM1_COMBO   = SNARE_COMBO | TOM1_COMBO
	SNARE_TOM2_COMBO   = SNARE_COMBO | TOM2_COMBO
	SNARE_TOM3_COMBO   = SNARE_COMBO | TOM3_COMBO
	SNARE_HIHAT_COMBO  = SNARE_COMBO | HIHAT_COMBO
	SNARE_HIHAT2_COMBO = SNARE_COMBO | HIHAT2_COMBO
	SNARE_RIDE_COMBO   = SNARE_COMBO | RIDE_COMBO
	SNARE_RIDE2_COMBO  = SNARE_COMBO | RIDE2_COMBO
	SNARE_CRASH_COMBO  = SNARE_COMBO | CRASH_COMBO

	TOM1_TOM2_COMBO   = TOM1_COMBO | TOM2_COMBO
	TOM1_TOM3_COMBO   = TOM1_COMBO | TOM3_COMBO
	TOM1_HIHAT_COMBO  = TOM1_COMBO | HIHAT_COMBO
	TOM1_HIHAT2_COMBO = TOM1_COMBO | HIHAT2_COMBO
	TOM1_RIDE_COMBO   = TOM1_COMBO | RIDE_COMBO
	TOM1_RIDE2_COMBO  = TOM1_COMBO | RIDE2_COMBO
	TOM1_CRASH_COMBO  = TOM1_COMBO | CRASH_COMBO

	TOM2_TOM3_COMBO   = TOM2_COMBO | TOM3_COMBO
	TOM2_HIHAT_COMBO  = TOM2_COMBO | HIHAT_COMBO
	TOM2_HIHAT2_COMBO = TOM2_COMBO | HIHAT2_COMBO
	TOM2_RIDE_COMBO   = TOM2_COMBO | RIDE_COMBO
	TOM2_RIDE2_COMBO  = TOM2_COMBO | RIDE2_COMBO
	TOM2_CRASH_COMBO  = TOM2_COMBO | CRASH_COMBO

	TOM3_HIHAT_COMBO  = TOM3_COMBO | HIHAT_COMBO
	TOM3_HIHAT2_COMBO = TOM3_COMBO | HIHAT2_COMBO
	TOM3_RIDE_COMBO   = TOM3_COMBO | RIDE_COMBO
	TOM3_RIDE2_COMBO  = TOM3_COMBO | RIDE2_COMBO
	TOM3_CRASH_COMBO  = TOM3_COMBO | CRASH_COMBO

	HIHAT_RIDE_COMBO   = (YELLOW_MASK | RIDE2_COMBO)
	HIHAT_RIDE_SIMPLE_COMBO = (YELLOW_MASK | BLUE_MASK | CYMBAL_MASK)
	HIHAT_CRASH_SIMPLE_COMBO = (YELLOW_MASK | GREEN_MASK | CYMBAL_MASK)
	HIHAT_CRASH_COMBO  = HIHAT_COMBO | CRASH_COMBO
	HIHAT2_CRASH_COMBO = HIHAT2_COMBO | CRASH_COMBO

	RIDE_CRASH_COMBO = RIDE_COMBO | CRASH_COMBO
	RIDE2_CRASH_COMBO = RIDE2_COMBO | CRASH_COMBO

	COMBO_FINISHED = KICK_CONNECTED_MASK | KICK_MASK
)

const (
	GREEN = iota
	RED
	BLUE
	YELLOW
	KICK
	CYMBAL
	TOM
	BLUE_CYMBAL
	YELLOW_CYMBAL
	GREEN_CYMBAL
)

var pendingMessages []int = []int{}
var knownMessages []int = []int{}

func getColors(messages []int) []int {
	var result []int
	for _, msg := range messages {
		switch msg {
		case GREEN, RED, BLUE, YELLOW:
			result = append(result, msg)
		}
	}
	return result
}

func containsSameColorCymbal(messages []int, color int) bool {
	switch color {
	case GREEN:
		return slices.Contains(messages, GREEN_CYMBAL)
	case BLUE:
		return slices.Contains(messages, BLUE_CYMBAL)
	case YELLOW:
		return slices.Contains(messages, YELLOW_CYMBAL)
	default:
		return false
	}
}

func findIonDrumJoystick(name string) (joystick.Joystick, error) {
	numJoysticks := 0
	joystickPath := "/dev/input"
	// get the number of joysticks by checking for existing device files by direct file access
	entries, err := os.ReadDir(joystickPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read joystick directory: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name()[:2] == "js" {
			numJoysticks++
		}
	}

	slog.Info("Found joysticks", "amount", numJoysticks)

	for id := range numJoysticks {
		js, err := joystick.Open(id)
		if err != nil || js == nil {
			slog.Error("Failed to open joystick", "id", strconv.Itoa(id), "error", err)
			continue
		}
		jsName := strings.Trim(js.Name(), "\x00")
		slog.Debug("Considered joystick", "id", id, "name", jsName)
		if jsName == name {
			return js, nil
		}
		js.Close()
	}
	return nil, fmt.Errorf("%s joystick not found", name)
}

type NotePlayer interface {
	Setup() error
	SendNote(note MidiNote, vel uint8, on bool)
	Sleep(duration time.Duration)
}

type StateHandler interface {
	Handle(state joystick.State) error
}

func listJoysticks(_ *Settings) {
	numJoysticks := 0
	joystickPath := "/dev/input"
	// get the number of joysticks by checking for existing device files by direct file access
	entries, err := os.ReadDir(joystickPath)
	if err != nil {
		log.Fatalln("FATAL failed to read joystick directory: ", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name()[:2] == "js" {
			numJoysticks++
		}
	}

	slog.Info("Found joysticks", "amount", numJoysticks)

	for id := range numJoysticks {
		js, err := joystick.Open(id)
		if err != nil || js == nil {
			slog.Error("Failed to open joystick", "id", id, "error", err)
			continue
		}
		jsName := strings.Trim(js.Name(), "\x00")
		fmt.Printf("Joystick %d: %s\n", id, jsName)
		js.Close()
	}
}

func replay(settings *Settings, fileName string, ignorePauses bool, np NotePlayer) {
	err := np.Setup()
	if err != nil {
		log.Fatalln("FATAL", err)
	}
	var file *os.File
	if fileName == "" || fileName == "-" {
		file = os.Stdin
	} else {
		file, err = os.Open(fileName)
		if err != nil {
			log.Fatalln("FATAL", err)
		}
	}
	defer file.Close()
	var handler = &BridgeStateHandler{}
	handler.player = &np
	scanner := bufio.NewScanner(file)
	var lineNumber = 0
	for scanner.Scan() {
		lineNumber++
		text := scanner.Text()
		text = strings.TrimSpace(text)
		if strings.Contains(text, ": ") {
			value := strings.Split(text, ": ")[1]
			if !ignorePauses && strings.HasPrefix(text, "duration") {
				number, err := strconv.Atoi(value)
				if err != nil {
					slog.Error("line parse error", "line", lineNumber, "string", value, "error", err)
				}
				millis := time.Duration(number)
				np.Sleep(millis * time.Millisecond)
			} else if strings.HasPrefix(text, "buttons") {
				number, err := strconv.ParseUint(value, 2, 64)
				if err != nil {
					slog.Error("line parse error", "line", lineNumber, "string", value, "error", err)
				}
				num32 := uint32(number)
				state := joystick.State{
					Buttons: num32,
				}
				err = handler.Handle(state)
				if err != nil {
					log.Fatalln("FATAL ", err)
				}

			} else {
				continue
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("FATAL ", "file scanner error: ", err)
	}
}

func record(settings *Settings) {
	handler := &RecordStateHandler{}
	loop(context.Background(), settings, handler)
}

type RecordStateHandler struct {
	stream   *os.File
	duration int64
	prevTime int64
}

// just print out the state of the buttons into the stream
func (h *RecordStateHandler) Handle(state joystick.State) error {
	duration := time.Now().UnixMilli() - h.prevTime
	if h.prevTime == 0 {
		duration = 0
	}
	h.prevTime = time.Now().UnixMilli()
	if duration > 0 {
		_, err := fmt.Fprintf(h.stream, "duration: %d\n", duration)
		if err != nil {
			log.Fatalln("FATAL ", err)
		}
	}
	buttons := state.Buttons
	axisData := state.AxisData
	if h.stream == nil {
		h.stream = os.Stdout
	}
	_, err := fmt.Fprintf(h.stream, "buttons: %032b\n", buttons)
	for i, axis := range axisData {
		_, err := fmt.Fprintf(h.stream, "axis%d: %d\n", i, axis)
		if err != nil {
			log.Fatalln("FATAL ", err)
		}
	}
	if err != nil {
		log.Fatalln("FATAL ", err)
	}
	return nil
}

func bridge(ctx context.Context, settings *Settings, np NotePlayer) {
	mp := np
	err := mp.Setup()
	if err != nil {
		log.Fatalln("FATAL ", err)
	}
	handler := &BridgeStateHandler{}
	handler.player = &np
	loop(ctx, settings, handler)
}

func loop(ctx context.Context, settings *Settings, handler StateHandler) {
	js, err := findIonDrumJoystick(settings.joyName)
	if err != nil {
		log.Fatalln("FATAL ", err)
	}
	defer js.Close()

	jsName := strings.Trim(js.Name(), "\x00")
	slog.Info("Got joystick", "name", jsName)
	slog.Debug("Joystick", "axis count", js.AxisCount(), "button count", js.ButtonCount())
	prevState, err := js.Read()
	if err != nil {
		log.Fatalln("FATAL ", err)
	}
	err = handler.Handle(prevState)
	if err != nil {
		log.Fatalln("FATAL ", err)
	}

	defer func() {
		out.Close()
		drv.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context canceled, exiting...")
			return
		default:
		}
		state, err := js.Read()
		if err != nil {
			log.Fatalln("FATAL ", err)
		}

		xor := state.Buttons ^ prevState.Buttons
		if xor == 0 {
			continue
		}
		err = handler.Handle(state)
		if err != nil {
			log.Fatalln("FATAL ", err)
		}
		prevState = state
	}
}

type BridgeStateHandler struct {
	prevState  joystick.State
	comboState joystick.State
	player     *NotePlayer
}

func (h *BridgeStateHandler) Handle(state joystick.State) error {
	xor := state.Buttons ^ h.prevState.Buttons
	var toPrint string
	if xor&CYMBAL_MASK != 0 && state.Buttons&CYMBAL_MASK == CYMBAL_MASK {
		toPrint = toPrint + "CYMBAL "
		pendingMessages = append(pendingMessages, CYMBAL)
	}
	if xor&TOM_MASK != 0 && state.Buttons&TOM_MASK == TOM_MASK {
		toPrint = toPrint + "TOM "
		pendingMessages = append(pendingMessages, TOM)
	}
	if xor&GREEN_MASK != 0 && state.Buttons&GREEN_MASK == GREEN_MASK {
		toPrint = toPrint + "GREEN "
		pendingMessages = append(pendingMessages, GREEN)
	}
	if xor&RED_MASK != 0 && state.Buttons&RED_MASK == RED_MASK {
		toPrint = toPrint + "RED "
		pendingMessages = append(pendingMessages, RED)
	}
	if xor&BLUE_MASK != 0 && state.Buttons&BLUE_MASK == BLUE_MASK {
		toPrint = toPrint + "BLUE "
		pendingMessages = append(pendingMessages, BLUE)
	}
	if xor&YELLOW_MASK != 0 && state.Buttons&YELLOW_MASK == YELLOW_MASK {
		toPrint = toPrint + "YELLOW "
		pendingMessages = append(pendingMessages, YELLOW)
	}
	if xor&KICK_MASK != 0 && state.Buttons&KICK_MASK == KICK_MASK {
		toPrint = toPrint + "KICK "
		pendingMessages = append(pendingMessages, KICK)
	}
	if xor&YELLOW_CYMBAL_MOD_MASK != 0 && state.Buttons&YELLOW_CYMBAL_MOD_MASK == YELLOW_CYMBAL_MOD_MASK {
		toPrint = toPrint + "YCYMBAL MOD "
		pendingMessages = append(pendingMessages, YELLOW_CYMBAL)
	}
	if xor&BLUE_CYMBAL_MOD_MASK != 0 && state.Buttons&BLUE_CYMBAL_MOD_MASK == BLUE_CYMBAL_MOD_MASK {
		toPrint = toPrint + "BCYMBAL MOD "
		pendingMessages = append(pendingMessages, BLUE_CYMBAL)
	}

	slog.Debug("Changed", "buttons", toPrint)
	if xor&KICK_MASK != 0 && state.Buttons&KICK_MASK == KICK_MASK {
		p := h.player
		player := *p
		player.SendNote(NOTE_KICK, FULL_VOLUME, true)
	}
	if (h.prevState.Buttons &^ COMBO_FINISHED) == 0 {
		h.comboState.Buttons = state.Buttons
	} else if (state.Buttons &^ COMBO_FINISHED) != 0 {
		h.comboState.Buttons = h.comboState.Buttons | state.Buttons
	} else {
		p := h.player
		player := *p
		noKicks := h.comboState.Buttons &^ COMBO_FINISHED
		switch noKicks {
		case SNARE_COMBO:
			player.SendNote(NOTE_SNARE, FULL_VOLUME, true)
		case TOM1_COMBO:
			player.SendNote(NOTE_TOM1, FULL_VOLUME, true)
		case TOM2_COMBO:
			player.SendNote(NOTE_TOM2, FULL_VOLUME, true)
		case TOM3_COMBO:
			player.SendNote(NOTE_TOM3, FULL_VOLUME, true)
		case HIHAT_COMBO, HIHAT2_COMBO:
			player.SendNote(NOTE_HIHAT, FULL_VOLUME, true)
		case RIDE_COMBO, RIDE2_COMBO:
			player.SendNote(NOTE_RIDE, FULL_VOLUME, true)
		case CRASH_COMBO:
			player.SendNote(NOTE_CRASH, FULL_VOLUME, true)
		case SNARE_TOM1_COMBO:
			player.SendNote(NOTE_SNARE, FULL_VOLUME, true)
			player.SendNote(NOTE_TOM1, FULL_VOLUME, true)
		case SNARE_TOM2_COMBO:
			player.SendNote(NOTE_SNARE, FULL_VOLUME, true)
			player.SendNote(NOTE_TOM2, FULL_VOLUME, true)
		case SNARE_TOM3_COMBO:
			player.SendNote(NOTE_SNARE, FULL_VOLUME, true)
			player.SendNote(NOTE_TOM3, FULL_VOLUME, true)
		case SNARE_HIHAT_COMBO, SNARE_HIHAT2_COMBO:
			player.SendNote(NOTE_SNARE, FULL_VOLUME, true)
			player.SendNote(NOTE_HIHAT, FULL_VOLUME, true)
		case SNARE_RIDE_COMBO, SNARE_RIDE2_COMBO:
			player.SendNote(NOTE_SNARE, FULL_VOLUME, true)
			player.SendNote(NOTE_RIDE, FULL_VOLUME, true)
		case SNARE_CRASH_COMBO:
			player.SendNote(NOTE_SNARE, FULL_VOLUME, true)
			player.SendNote(NOTE_CRASH, FULL_VOLUME, true)
		case TOM1_TOM2_COMBO:
			player.SendNote(NOTE_TOM1, FULL_VOLUME, true)
			player.SendNote(NOTE_TOM2, FULL_VOLUME, true)
		case TOM1_TOM3_COMBO:
			player.SendNote(NOTE_TOM1, FULL_VOLUME, true)
			player.SendNote(NOTE_TOM3, FULL_VOLUME, true)
		case TOM1_HIHAT_COMBO, TOM1_HIHAT2_COMBO:
			player.SendNote(NOTE_TOM1, FULL_VOLUME, true)
			player.SendNote(NOTE_HIHAT, FULL_VOLUME, true)
		case TOM1_RIDE_COMBO, TOM1_RIDE2_COMBO:
			player.SendNote(NOTE_TOM1, FULL_VOLUME, true)
			player.SendNote(NOTE_RIDE, FULL_VOLUME, true)
		case TOM1_CRASH_COMBO:
			player.SendNote(NOTE_TOM1, FULL_VOLUME, true)
			player.SendNote(NOTE_CRASH, FULL_VOLUME, true)
		case TOM2_TOM3_COMBO:
			player.SendNote(NOTE_TOM2, FULL_VOLUME, true)
			player.SendNote(NOTE_TOM3, FULL_VOLUME, true)
		case TOM2_HIHAT_COMBO, TOM2_HIHAT2_COMBO:
			player.SendNote(NOTE_TOM2, FULL_VOLUME, true)
			player.SendNote(NOTE_HIHAT, FULL_VOLUME, true)
		case TOM2_RIDE_COMBO, TOM2_RIDE2_COMBO:
			player.SendNote(NOTE_TOM2, FULL_VOLUME, true)
			player.SendNote(NOTE_RIDE, FULL_VOLUME, true)
		case TOM2_CRASH_COMBO:
			player.SendNote(NOTE_TOM2, FULL_VOLUME, true)
			player.SendNote(NOTE_CRASH, FULL_VOLUME, true)
		case TOM3_HIHAT_COMBO, TOM3_HIHAT2_COMBO:
			player.SendNote(NOTE_TOM3, FULL_VOLUME, true)
			player.SendNote(NOTE_HIHAT, FULL_VOLUME, true)
		case TOM3_RIDE_COMBO, TOM3_RIDE2_COMBO:
			player.SendNote(NOTE_TOM3, FULL_VOLUME, true)
			player.SendNote(NOTE_RIDE, FULL_VOLUME, true)
		case TOM3_CRASH_COMBO:
			player.SendNote(NOTE_TOM3, FULL_VOLUME, true)
			player.SendNote(NOTE_CRASH, FULL_VOLUME, true)
		case HIHAT_RIDE_SIMPLE_COMBO, HIHAT_RIDE_COMBO:
			player.SendNote(NOTE_HIHAT, FULL_VOLUME, true)
			player.SendNote(NOTE_RIDE, FULL_VOLUME, true)
		case HIHAT_CRASH_COMBO, HIHAT2_CRASH_COMBO, HIHAT_CRASH_SIMPLE_COMBO:
			player.SendNote(NOTE_HIHAT, FULL_VOLUME, true)
			player.SendNote(NOTE_CRASH, FULL_VOLUME, true)
		case RIDE_CRASH_COMBO, RIDE2_CRASH_COMBO:
			player.SendNote(NOTE_RIDE, FULL_VOLUME, true)
			player.SendNote(NOTE_CRASH, FULL_VOLUME, true)
		default:
			binary := fmt.Sprintf("%016b", noKicks)
			binaryFours := ""
			for i := range len(binary) {
				if (len(binary)-i)%4 == 0 {
					binaryFours = binaryFours + " "
				}
				binaryFours = binaryFours + string(binary[i])
			}
			binaryFours = strings.Trim(binaryFours, " ")
			fmt.Println("Unknown combo: " + binaryFours)
		}
	}

	if state.Buttons != h.prevState.Buttons {
		slog.Debug("Buttons", "state", fmt.Sprintf("%032b", state.Buttons))
		h.prevState = state
	}
	slog.Debug("-------------")
	return nil
}
