package main

import (
	"context"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "midi-drum-rocker",
	Short: "A command-line tool create a virtual MIDI device from Ion Drum Rocker",
	Run: func(cmd *cobra.Command, args []string) {
		// Default action when no subcommand is provided
		cmd.Help()
	},
}

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Run midi bridge along with a command",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			ctx, cancel := context.WithCancel(context.Background())
			go run(ctx, getSettings(cmd))
			execCommand(args)
			cancel()
		} else {
			cmd.Help()
		}
	},
}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record joystick events to a file",
	Run: func(cmd *cobra.Command, args []string) {
		record(getSettings(cmd))
	},
}

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Play recorded joystick events from a file, sending MIDI messages",
	Run: func(cmd *cobra.Command, args []string) {
		play(getSettings(cmd))
	},
}

func getSettings(cmd *cobra.Command) *Settings {
	return &Settings{
		joyName: cmd.Root().Flag("joy-name").Value.String(),
	}
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the midi bridge",
	Run: func(cmd *cobra.Command, args []string) {
		run(context.Background(), getSettings(cmd))
	},
}

var testMidiCmd = &cobra.Command{
	Use:   "test-midi",
	Short: "Test sending MIDI messages",
	Run: func(cmd *cobra.Command, args []string) {
		testMidi()
	},
}

func execCommand(args []string) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func parseArgs() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.Flags().StringP("joy-name", "j", "Ion Drum Rocker", "Name of the joystick to use")
	rootCmd.AddCommand(launchCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(testMidiCmd)
	rootCmd.AddCommand(recordCmd)
}
