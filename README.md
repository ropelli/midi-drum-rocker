# Ion Drum Rocker Midi Bridge

This project provides a MIDI bridge for the Ion Drum Rocker Rock Band drum set.
It allows you to connect the drum set to your computer and use it as a MIDI controller.
Latency is pretty bad at the moment, I am unsure if it can be improved.
The project has been tested in Linux only and with XBox 360 version of the drum set.
The drum set has a wired USB connection and needs to be plugged in before starting the program.
The program presumes that the joystick device is located at `/dev/input/js1`
and the virtual MIDI device is named `IonDrumBridge`.
Only the cymbals, kick and pads are supported at the moment. The buttons haven't been mapped to anything.
