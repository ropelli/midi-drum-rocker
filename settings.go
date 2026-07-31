package main

const NO_RETRY = false
const DO_RETRY = true

type Settings struct {
	joyName  string
	midiName string
	retry    bool
}
