# Ion Drum Rocker Midi Bridge (midi-drum-rocker)

This project provides a MIDI bridge for the Ion Drum Rocker Rock Band drum set.
It allows you to connect the drum set to your computer and use it as a MIDI controller.
This can also be used to play games like YARG which support MIDI input and where the support for the drum set is not implemented yet.
The virtual MIDI device that is automatically created is named `IonDrumBridge`.

## Issues

Latency is pretty bad at the moment, I am unsure if it can be improved.
The project has been tested in Linux only and with XBox 360 version of the drum set.
The drum set has a wired USB connection and needs to be plugged in before starting the program.
The program finds the first joystick device whose name is Ion Drum Rocker, so it supports only one drum set at a time.
Only the cymbals, kick and pads are supported at the moment. The buttons haven't been mapped.
Also, velocity is not supported, so the drum set will always send maximum velocity.

## Usage

```sh
# Foreground mode
./midi-drum-rocker run

# Command line help
./midi-drum-rocker --help

# Background mode, use for Steam Launch options for e.g. YARG which supports MIDI input
/path/to/midi-drum-rocker launch -- %command%
```

## Building

You may need to install dependencies even on top of go dependencies:
```sh
# On fedora:
sudo dnf install alsa-lib-devel
```

To build the project, run the following command:
```sh
go build .
```
