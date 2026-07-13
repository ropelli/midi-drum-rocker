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
	GREEN_MASK             uint32 = 0b1
	RED_MASK               uint32 = 0b10
	BLUE_MASK              uint32 = 0b100
	YELLOW_MASK            uint32 = 0b1000
	KICK_MASK              uint32 = 0b10000
	KICK_CONNECTED_MASK    uint32 = 0b1000000000
	CYMBAL_MASK            uint32 = 0b100000
	TOM_MASK               uint32 = 0b10000000000
	YELLOW_CYMBAL_MOD_MASK uint32 = 0b10_000_000_000_000
	BLUE_CYMBAL_MOD_MASK   uint32 = 0b100_000_000_000_000
)

const (
	FULL_VOLUME = 100
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

func containsOtherColors(messages []int, exclude int) bool {
	for _, msg := range pendingMessages {
		if msg != exclude {
			return true
		}
	}
	return false
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

func replay(settings *Settings, fileName string, ignorePauses bool) {
	initMIDI(settings.midiName)
	var file *os.File
	if fileName == "" || fileName == "-" {
		file = os.Stdin
	} else {
		file, err := os.Open(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	}
	var handler = &BridgeStateHandler{}
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
				slog.Debug("sleeping", "ms", value)
				time.Sleep(millis * time.Millisecond)
			} else if strings.HasPrefix(text, "buttons") {
				number, err := strconv.ParseUint(value, 2, 64)
				if err != nil {
					slog.Error("line parse error", "line", lineNumber, "string", value, "error", err)
				}
				num32 := uint32(number)
				state := joystick.State{
					Buttons: num32,
				}
				handler.Handle(state)

			} else {
				continue
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
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
	if h.stream == nil {
		h.stream = os.Stdout
	}
	_, err := fmt.Fprintf(h.stream, "buttons: %032b\n", buttons)
	if err != nil {
		log.Fatalln("FATAL ", err)
	}
	return nil
}

func bridge(ctx context.Context, settings *Settings) {
	err := initMIDI(settings.midiName)
	if err != nil {
		log.Fatalln("FATAL ", err)
	}
	handler := &BridgeStateHandler{}
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
	handler.Handle(prevState)
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
		handler.Handle(state)
		prevState = state
	}
}

type BridgeStateHandler struct {
	prevState joystick.State
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
		knownMessages = append(knownMessages, KICK)
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

	if state.Buttons == KICK_CONNECTED_MASK || state.Buttons == 0 || state.Buttons == KICK_MASK || state.Buttons == (KICK_MASK|KICK_CONNECTED_MASK) {
		if slices.Contains(pendingMessages, CYMBAL) {
			if slices.Contains(pendingMessages, YELLOW_CYMBAL) {
				knownMessages = append(knownMessages, YELLOW_CYMBAL)
			}
			if slices.Contains(pendingMessages, BLUE_CYMBAL) {
				knownMessages = append(knownMessages, BLUE_CYMBAL)
			}
			if slices.Contains(pendingMessages, CYMBAL) && slices.Contains(pendingMessages, GREEN) {
				if !slices.Contains(pendingMessages, TOM) {
					knownMessages = append(knownMessages, GREEN_CYMBAL)
				} else if slices.Contains(pendingMessages, TOM) {
					if !slices.Contains(knownMessages, YELLOW_CYMBAL) && !slices.Contains(knownMessages, BLUE_CYMBAL) {
						knownMessages = append(knownMessages, GREEN_CYMBAL)
					}
				}
			}
		}
		if slices.Contains(pendingMessages, TOM) {
			var pendingColors = getColors(pendingMessages)
			if len(pendingColors) == 1 {
				knownMessages = append(knownMessages, pendingColors[0])
			} else {
				for _, color := range pendingColors {
					if !containsSameColorCymbal(knownMessages, color) {
						knownMessages = append(knownMessages, color)
					}
				}

			}
		}

	}
	if state.Buttons == KICK_CONNECTED_MASK || state.Buttons == 0 || state.Buttons == KICK_MASK || state.Buttons == (KICK_MASK|KICK_CONNECTED_MASK) {
		slog.Debug("Known messages", "msgs", knownMessages)
		for _, msg := range knownMessages {
			switch msg {
			case RED:
				sendNote(NOTE_SNARE, FULL_VOLUME, true)
			case YELLOW:
				sendNote(NOTE_TOM1, FULL_VOLUME, true)
			case BLUE:
				sendNote(NOTE_TOM2, FULL_VOLUME, true)
			case GREEN:
				sendNote(NOTE_TOM3, FULL_VOLUME, true)
			case YELLOW_CYMBAL:
				sendNote(NOTE_HIHAT, FULL_VOLUME, true)
			case BLUE_CYMBAL:
				sendNote(NOTE_RIDE, FULL_VOLUME, true)
			case GREEN_CYMBAL:
				sendNote(NOTE_CRASH, FULL_VOLUME, true)
			case KICK:
				sendNote(NOTE_KICK, FULL_VOLUME, true)
			}
		}

		knownMessages = []int{}
		pendingMessages = []int{}
	}
	if slog.Default().Enabled(nil, slog.LevelDebug) && state.Buttons != h.prevState.Buttons {
		slog.Debug("Buttons", "state", fmt.Sprintf("%032b", state.Buttons))
		h.prevState = state
	}
	slog.Debug("-------------")
	return nil
}
