package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/0xcafed00d/joystick"
)

var greenMask uint32 = 0b1
var redMask uint32 = 0b10
var blueMask uint32 = 0b100
var yellowMask uint32 = 0b1000
var kickMask uint32 = 0b10000
var kickConnectedMask uint32 = 0b1000000000
var cymbalMask uint32 = 0b100000
var tomMask uint32 = 0b10000000000
var yellowCymbalModMask uint32 = 0b10_000_000_000_000
var blueCymbalModMask uint32 = 0b100_000_000_000_000
var greenTomMask uint32 = tomMask | greenMask
var redTomMask uint32 = tomMask | redMask
var blueTomMask uint32 = tomMask | blueMask
var yellowTomMask uint32 = tomMask | yellowMask
var greenCymbalMask uint32 = cymbalMask | greenMask
var blueCymbalMask uint32 = cymbalMask | blueMask | blueCymbalModMask
var yellowCymbalMask uint32 = cymbalMask | yellowMask | yellowCymbalModMask

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
var colors = []int{GREEN, RED, BLUE, YELLOW}

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

	log.Println("Found joystick entries:", numJoysticks)

	for id := range numJoysticks {
		js, err := joystick.Open(id)
		if err != nil || js == nil {
			log.Printf("Failed to open joystick %d: %v\n", id, err)
			continue
		}
		log.Printf("Joystick %d: %s\n", id, js.Name())
		if strings.Contains(js.Name(), "Ion Drum Rocker") {
			return js, nil
		}
		js.Close()
	}
	return nil, fmt.Errorf("%s joystick not found", name)
}

func main() {
	parseArgs()
}

func run(ctx context.Context, settings *Settings) {
	js, err := findIonDrumJoystick(settings.joyName)
	if err != nil {
		log.Fatal(err)
	}
	defer js.Close()

	log.Printf("Joystick Name: %s", js.Name())
	log.Printf("   Axis Count: %d", js.AxisCount())
	log.Printf(" Button Count: %d\n", js.ButtonCount())
	prevState, err := js.Read()
	if err != nil {
		log.Fatal(err)
	}

	initMIDI()
	defer func() {
		out.Close()
		drv.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Context canceled, exiting...")
			return
		default:
		}
		state, err := js.Read()
		if err != nil {
			log.Fatal(err)
		}

		xor := state.Buttons ^ prevState.Buttons
		if xor == 0 {
			continue
		}
		var toPrint []string = []string{}
		if xor&cymbalMask != 0 && state.Buttons&cymbalMask == cymbalMask {
			toPrint = append(toPrint, "CYMBAL ")
			pendingMessages = append(pendingMessages, CYMBAL)
		}
		if xor&tomMask != 0 && state.Buttons&tomMask == tomMask {
			toPrint = append(toPrint, "TOM ")
			pendingMessages = append(pendingMessages, TOM)
		}
		if xor&greenMask != 0 && state.Buttons&greenMask == greenMask {
			toPrint = append(toPrint, "GREEN ")
			pendingMessages = append(pendingMessages, GREEN)
		}
		if xor&redMask != 0 && state.Buttons&redMask == redMask {
			toPrint = append(toPrint, "RED ")
			pendingMessages = append(pendingMessages, RED)
		}
		if xor&blueMask != 0 && state.Buttons&blueMask == blueMask {
			toPrint = append(toPrint, "BLUE ")
			pendingMessages = append(pendingMessages, BLUE)
		}
		if xor&yellowMask != 0 && state.Buttons&yellowMask == yellowMask {
			toPrint = append(toPrint, "YELLOW ")
			pendingMessages = append(pendingMessages, YELLOW)
		}
		if xor&kickMask != 0 && state.Buttons&kickMask == kickMask {
			toPrint = append(toPrint, "KICK ")
			pendingMessages = append(pendingMessages, KICK)
			knownMessages = append(knownMessages, KICK)
		}
		if xor&yellowCymbalModMask != 0 && state.Buttons&yellowCymbalModMask == yellowCymbalModMask {
			toPrint = append(toPrint, "YCYMBAL MOD ")
			pendingMessages = append(pendingMessages, YELLOW_CYMBAL)
		}
		if xor&blueCymbalModMask != 0 && state.Buttons&blueCymbalModMask == blueCymbalModMask {
			toPrint = append(toPrint, "BCYMBAL MOD ")
			pendingMessages = append(pendingMessages, BLUE_CYMBAL)
		}

		if len(toPrint) > 0 {
			log.Printf("Changed: ")
			for _, str := range toPrint {
				log.Printf("%s", str)
			}
			log.Printf("\n")
		}

		if state.Buttons == kickConnectedMask || state.Buttons == 0 || state.Buttons == kickMask || state.Buttons == (kickMask|kickConnectedMask) {
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

			log.Println("Known messages:", knownMessages)
			for _, msg := range knownMessages {
				switch msg {
				case RED:
					sendNote(NOTE_SNARE, 100, true)
				case YELLOW:
					sendNote(NOTE_TOM1, 100, true)
				case BLUE:
					sendNote(NOTE_TOM2, 100, true)
				case GREEN:
					sendNote(NOTE_TOM3, 100, true)
				case YELLOW_CYMBAL:
					sendNote(NOTE_HIHAT, 100, true)
				case BLUE_CYMBAL:
					sendNote(NOTE_RIDE, 100, true)
				case GREEN_CYMBAL:
					sendNote(NOTE_CRASH, 100, true)
				case KICK:
					sendNote(NOTE_KICK, 100, true)
				}
			}

			log.Println("-------------")
			knownMessages = []int{}
			pendingMessages = []int{}
		}

		if state.Buttons != prevState.Buttons {
			log.Printf("Buttons: %032b\n", state.Buttons)
			prevState = state
		}
	}
}
