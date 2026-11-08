package main_test

import(
	"github.com/0xcafed00d/joystick"
	"ropelli/midi-drum-rocker"
	"testing"
)

func TestKeyboard_Read(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		want    joystick.State
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, err := main.NewKeyboard()
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got, gotErr := k.Read()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Read() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Read() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

