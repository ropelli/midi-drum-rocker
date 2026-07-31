package main

import (
	"context"
	"fmt"
	"log/slog"
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
			s := getSettings(cmd)
			mp := NewMidiPlayer(s)
			go bridge(ctx, getSettings(cmd), mp)
			execCommand(args)
			cancel()
		} else {
			cmd.Help()
		}
	},
}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record joystick events to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		record(getSettings(cmd))
	},
}

var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replay recorded joystick events from a file, sending MIDI messages",
	Run: func(cmd *cobra.Command, args []string) {
		f, err := cmd.Flags().GetString("input-file")
		if err != nil {
			panic(err)
		}
		ip, err := cmd.Flags().GetBool("ignore-pauses")
		if err != nil {
			panic(err)
		}
		s := getSettings(cmd)
		mp := NewMidiPlayer(s)
		replay(s, f, ip, mp)
	},
}

var listJoysCmd = &cobra.Command{
	Use:   "list",
	Short: "List available joysticks",
	Run: func(cmd *cobra.Command, args []string) {
		listJoysticks(getSettings(cmd))
	},
}

func getSettings(cmd *cobra.Command) *Settings {
	bool, err := cmd.Root().Flags().GetBool("debug")
	if err != nil {
		panic(err)
	}
	if bool {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}
	return &Settings{
		joyName:  cmd.Root().Flag("joy-name").Value.String(),
		midiName: cmd.Root().Flag("midi-device").Value.String(),
	}
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the midi bridge",
	Run: func(cmd *cobra.Command, args []string) {
		s := getSettings(cmd)
		mp := NewMidiPlayer(s)
		bridge(context.Background(), s, mp)
	},
}

var testMidiCmd = &cobra.Command{
	Use:   "test-midi",
	Short: "Test sending MIDI messages",
	Run: func(cmd *cobra.Command, args []string) {
		mp := NewMidiPlayer(getSettings(cmd))
		mp.TestMidi()
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringP("joy-name", "j", "Ion Drum Rocker", "Name of the joystick to use")
	rootCmd.PersistentFlags().StringP("midi-device", "d", "IonDrumBridge", "Name of the midi device to create")
	rootCmd.AddCommand(launchCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(testMidiCmd)
	rootCmd.AddCommand(recordCmd)
	rootCmd.AddCommand(replayCmd)
	replayCmd.Flags().StringP("input-file", "f", "-", "Input file to replay, use stdin with -")
	replayCmd.Flags().Bool("ignore-pauses", false, "Ignore pauses between button presses")
	rootCmd.AddCommand(listJoysCmd)
}
