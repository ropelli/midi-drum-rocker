package main

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/0xcafed00d/joystick"
	goinput "github.com/AndrewDonelson/go-input"
	"github.com/AndrewDonelson/go-input/keyboard"
	"github.com/holoplot/go-evdev"
)

type Keyboard struct {
	kb   *keyboard.Watcher
	devs []*evdev.InputDevice
}

// Keyboard implements joystick.Joystick
var _ joystick.Joystick = (*Keyboard)(nil)

func NewKeyboard() (*Keyboard, error) {
	var devs []*evdev.InputDevice
	for i := 0; i < 32; i++ {
		path := fmt.Sprintf("/dev/input/event%d", i)
		dev, err := evdev.Open(path)
		if err != nil {
			continue
		}
		if slices.Contains(dev.CapableEvents(evdev.EV_KEY), evdev.KEY_Q) {
			slog.Info("Found keyboard device:", "path", path)
			devs = append(devs, dev)
		} else {
			dev.Close()
		}
	}
	if len(devs) > 0 {
		kb := goinput.NewKeyboard()
		return &Keyboard{kb: kb, devs: devs}, nil
	}
	return nil, fmt.Errorf("no keyboard device found")
}

var codeToKey = map[evdev.EvCode]keyboard.Key{
	evdev.KEY_Q:          keyboard.Q,
	evdev.KEY_W:          keyboard.W,
	evdev.KEY_E:          keyboard.E,
	evdev.KEY_R:          keyboard.R,
	evdev.KEY_T:          keyboard.T,
	evdev.KEY_Y:          keyboard.Y,
	evdev.KEY_U:          keyboard.U,
	evdev.KEY_I:          keyboard.I,
	evdev.KEY_O:          keyboard.O,
	evdev.KEY_P:          keyboard.P,
	evdev.KEY_LEFTBRACE:  keyboard.LeftBracket,
	evdev.KEY_RIGHTBRACE: keyboard.RightBracket,
	evdev.KEY_A:          keyboard.A,
	evdev.KEY_S:          keyboard.S,
	evdev.KEY_D:          keyboard.D,
	evdev.KEY_F:          keyboard.F,
	evdev.KEY_G:          keyboard.G,
	evdev.KEY_H:          keyboard.H,
	evdev.KEY_J:          keyboard.J,
	evdev.KEY_K:          keyboard.K,
	evdev.KEY_L:          keyboard.L,
	evdev.KEY_SEMICOLON:  keyboard.Semicolon,
	evdev.KEY_APOSTROPHE: keyboard.Apostrophe,
	evdev.KEY_BACKSLASH:  keyboard.BackSlash,
	evdev.KEY_Z:          keyboard.Z,
	evdev.KEY_X:          keyboard.X,
	evdev.KEY_C:          keyboard.C,
	evdev.KEY_V:          keyboard.V,
	evdev.KEY_B:          keyboard.B,
	evdev.KEY_N:          keyboard.N,
	evdev.KEY_M:          keyboard.M,
	evdev.KEY_COMMA:      keyboard.Comma,
}

var keysToBit = map[keyboard.Key]int{
	keyboard.Q:            31,
	keyboard.W:            30,
	keyboard.E:            29,
	keyboard.R:            28,
	keyboard.T:            27,
	keyboard.Y:            26,
	keyboard.U:            25,
	keyboard.I:            24,
	keyboard.O:            23,
	keyboard.P:            22,
	keyboard.LeftBracket:  21,
	keyboard.RightBracket: 20,
	keyboard.A:            19,
	keyboard.S:            18,
	keyboard.D:            17,
	keyboard.F:            16,
	keyboard.G:            15,
	keyboard.H:            14,
	keyboard.J:            13,
	keyboard.K:            12,
	keyboard.L:            11,
	keyboard.Semicolon:    10,
	keyboard.Apostrophe:   9,
	keyboard.BackSlash:    8,
	keyboard.Z:            7,
	keyboard.X:            6,
	keyboard.C:            5,
	keyboard.V:            4,
	keyboard.B:            3,
	keyboard.N:            2,
	keyboard.M:            1,
	keyboard.Comma:        0,
}

func (k *Keyboard) AxisCount() int {
	return 0
}

func (k *Keyboard) ButtonCount() int {
	return 32
}

func (k *Keyboard) Name() string {
	return "Keyboard"
}

func (k *Keyboard) Handle() error {
	for _, dev := range k.devs {
		go k.HandleOne(dev)
	}
	return nil
}

func (k *Keyboard) HandleOne(dev *evdev.InputDevice) error {
	pressed := make(map[evdev.EvCode]bool)
	for {
		event, err := dev.ReadOne()
		if err != nil {
			panic(err)
		}
		if event.Type == evdev.EV_KEY {
			switch event.Value {
			case 0: // Released
				pressed[event.Code] = false
				k.kb.SetState(codeToKey[event.Code], keyboard.Up)
			case 1: // Pressed
				k.kb.SetState(codeToKey[event.Code], keyboard.Down)
				pressed[event.Code] = true
			}
		}
	}
}

func (k *Keyboard) Read() (joystick.State, error) {
	var state joystick.State
	for key, button := range keysToBit {
		if k.kb.Down(key) {
			state.Buttons |= 1 << button
		}
	}
	return state, nil
}

func (k *Keyboard) Close() {
	for _, dev := range k.devs {
		defer dev.Close()
	}
}
