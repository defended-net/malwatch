// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSetStatus(t *testing.T) {
	tests := map[string]struct {
		start status
		input error
		want  status
	}{
		"nil": {
			start: StatusErr,
			input: nil,
			want:  StatusOK,
		},

		"err-cli": {
			start: StatusOK,
			input: errors.New("cli: malwatch -help"),
			want:  StatusErrArg,
		},

		"err-hit": {
			start: StatusOK,
			input: ErrHit,
			want:  StatusHit,
		},

		"err-other": {
			start: StatusOK,
			input: io.EOF,
			want:  StatusErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			state := &State{Exit: test.start}

			SetStatus(state, test.input)

			if state.Exit != test.want {
				t.Errorf("unexpected status: %v, want %v", state.Exit, test.want)
			}
		})
	}
}

func TestCancelAll(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	state := &State{
		Cancel: &Cancel{},
	}

	state.AddCancel(cancel)

	state.CancelAll()

	if err := ctx.Done(); err == nil {
		t.Errorf("unexpected ctx success")
	}
}

func TestGetCode(t *testing.T) {
	tests := map[string]struct {
		input error
		want  status
	}{
		"hit": {
			input: ErrHit,
			want:  StatusHit,
		},

		"cli": {
			input: errors.New("cli: malwatch -help"),
			want:  StatusErrArg,
		},

		"err": {
			input: io.EOF,
			want:  StatusErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if result := getStatus(test.input); result != test.want {
				t.Errorf("unexpected code: %v, want %v", result, test.want)
			}
		})
	}
}

func TestSetCode(t *testing.T) {
	tests := map[string]struct {
		input error
		want  status
	}{
		"cli": {
			input: errors.New("cli: malwatch -help"),
			want:  StatusErrArg,
		},

		"err": {
			input: io.EOF,
			want:  StatusErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := &State{}

			SetStatus(result, test.input)

			if result.Exit != test.want {
				t.Errorf("unexpected code: %v, want %v", result.Exit, test.want)
			}
		})
	}
}

func TestSetCodeHit(t *testing.T) {
	tests := map[string]struct {
		input error
		want  status
	}{
		"cli": {
			input: errors.New("cli: malwatch -help"),
			want:  StatusHit,
		},

		"err": {
			input: io.EOF,
			want:  StatusHit,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := &State{}

			SetStatus(result, ErrHit)

			SetStatus(result, test.input)

			if result.Exit != test.want {
				t.Errorf("unexpected code: %v, want %v", result.Exit, test.want)
			}
		})
	}
}

func TestLock(t *testing.T) {
	input := &State{
		Lockfile: filepath.Join(t.TempDir(), t.Name()),
	}

	if err := input.Lock(input.Lockfile); err != nil {
		t.Errorf("lockfile lock error: %v", err)
	}
}

func TestLockExist(t *testing.T) {
	if os.Getenv(t.Name()) == "1" {
		input := &State{
			Lockfile: filepath.Join(t.TempDir(), t.Name()),
		}

		file, err := os.Create(input.Lockfile)
		if err != nil {
			t.Fatalf("lockfile create error: %v", err)
		}
		defer file.Close()

		if err := input.Lock(input.Lockfile); err != nil {
			t.Errorf("lock error: %v", err)
		}

		return
	}

	// -cover arg breaks test.
	if len(os.Args) == 1 {
		t.Skip()
	}

	cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
	cmd.Env = append(os.Environ(), t.Name()+"=1")

	if e, ok := cmd.Run().(*exec.ExitError); ok && e.ExitCode() != 1 {
		t.Errorf("unexpected lockfile exit code: %v, want %v", e.ExitCode(), 1)
	}
}

func TestExit(t *testing.T) {
	tests := map[string]struct {
		input error
		want  int
	}{
		"TestStatusErr": {
			input: io.ErrUnexpectedEOF,
			want:  int(StatusErr),
		},

		"TestStatusErrArg": {
			input: errors.New("cli: malwatch -help"),
			want:  int(StatusErrArg),
		},

		"TestStatusHit": {
			input: ErrHit,
			want:  int(StatusHit),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run="+name)
			cmd.Env = append(os.Environ(), "SKIP=0")

			if result, ok := cmd.Run().(*exec.ExitError); ok && result.ExitCode() != test.want {
				t.Errorf("unexpected status: %v, want %v", result.ExitCode(), test.want)
			}
		})
	}
}

func TestStatusErr(t *testing.T) {
	if os.Getenv("SKIP") != "0" {
		t.Skip()
	}

	input := &State{
		Exit:   StatusErr,
		Cancel: &Cancel{},
	}

	Exit(input, nil)
}

func TestStatusArg(t *testing.T) {
	// Package test will otherwise cause exit status failure.
	if os.Getenv("SKIP") != "0" {
		t.Skip()
	}

	input := &State{
		Exit: StatusErrArg,
	}

	Exit(input, nil)
}

func TestStatusHit(t *testing.T) {
	// Package test will otherwise cause exit status failure.
	if os.Getenv("SKIP") != "0" {
		t.Skip()
	}

	input := &State{
		Exit:   StatusHit,
		Cancel: &Cancel{},
	}

	Exit(input, nil)
}
