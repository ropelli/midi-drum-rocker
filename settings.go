package main

import "time"

const NO_RETRY = false
const DO_RETRY = true

const NO_KEYBOARD = false
const USE_KEYBOARD = true

type Settings struct {
	joyName     string
	kickDelay   time.Duration
	midiName    string
	retry       bool
	useKeyboard bool
}

func GetDefaultSettings() *Settings {
	return &Settings{
		joyName:     "Ion Drum Rocker",
		kickDelay:   0,
		midiName:    "IonDrumBridge",
		retry:       NO_RETRY,
		useKeyboard: NO_KEYBOARD,
	}
}
