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

func TestNewState(t *testing.T) {
	var (
		input = NewState()
		want  = StatusOK
	)

	if input == nil {
		t.Fatal("expected non nil state")
	}

	if input.Signal == nil {
		t.Error("expected non nil chan")
	}

	if input.Cancel == nil {
		t.Error("expected non nil cancel")
	}

	if input.Exit != want {
		t.Errorf("unexpected exit status %v, want %v", input.Exit, want)
	}
}

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
			got := &State{
				Exit: test.start,
			}

			SetStatus(got, test.input)

			if got.Exit != test.want {
				t.Errorf("unexpected status %v, want %v", got.Exit, test.want)
			}
		})
	}
}

func TestCancelAll(t *testing.T) {
	var (
		ctx, cancel = context.WithCancel(context.Background())

		input = &State{
			Cancel: &Cancel{},
		}
	)

	input.AddCancel(cancel)

	input.CancelAll()

	if got := ctx.Done(); got == nil {
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

		"nil": {
			input: nil,
			want:  StatusOK,
		},

		"err": {
			input: io.EOF,
			want:  StatusErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := getStatus(test.input); got != test.want {
				t.Errorf("unexpected code: %v, want %v", got, test.want)
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
			got := &State{}

			SetStatus(got, test.input)

			if got.Exit != test.want {
				t.Errorf("unexpected code: %v, want %v", got.Exit, test.want)
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
			got := &State{}

			SetStatus(got, ErrHit)

			SetStatus(got, test.input)

			if got.Exit != test.want {
				t.Errorf("unexpected code: %v, want %v", got.Exit, test.want)
			}
		})
	}
}

func TestLock(t *testing.T) {
	tmp := t.TempDir()

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Errorf("open root error %v", err)
	}

	defer func() {
		//lint
		_ = root.Close()
	}()

	input := &State{
		LockPath: filepath.Join(tmp, t.Name()),
	}

	if got := input.Lock(root, input.LockPath); got != nil {
		t.Errorf("lock error %v", got)
	}

	if input.LockName == "" || input.LockRoot != root {
		t.Errorf("unexpected lock root/name population")
	}
}

func TestLockExist(t *testing.T) {
	if os.Getenv(t.Name()) == "1" {
		root, err := os.OpenRoot(t.TempDir())
		if err != nil {
			t.Errorf("open root error %v", err)
		}

		input := &State{
			LockPath: filepath.Join(t.TempDir(), t.Name()),
		}

		file, err := os.Create(input.LockPath)
		if err != nil {
			t.Fatalf("lockfile create error %v", err)
		}

		defer func() {
			//lint
			_ = root.Close()

			//lint
			_ = file.Close()
		}()

		if got := input.Lock(root, input.LockPath); got != nil {
			t.Errorf("lock error %v", got)
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

func TestLockCreateErr(t *testing.T) {
	var (
		tmp   = t.TempDir()
		input = &State{}
		path  = filepath.Join(tmp, t.Name(), "lock")
	)

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Errorf("open root error %v", err)
	}

	defer func() {
		//lint
		_ = root.Close()
	}()

	got := input.Lock(root, path)

	if got == nil {
		t.Fatal("expected lock create error")
	}

	if !errors.Is(got, ErrLockCreate) {
		t.Errorf("unexpected error %v, want %v", got, ErrLockCreate)
	}
}

func TestGetCancels(t *testing.T) {
	var (
		state = &State{
			Cancel: &Cancel{},
		}

		_, cancel1 = context.WithCancel(context.Background())
		_, cancel2 = context.WithCancel(context.Background())
	)

	state.AddCancel(cancel1)
	state.AddCancel(cancel2)

	got := state.GetCancels()

	if len(got) != 2 {
		t.Errorf("unexpected cancel count: %v, want 2", len(got))
	}

	got = state.GetCancels()

	if len(got) != 0 {
		t.Errorf("expected empty cancels, got %v", len(got))
	}
}

func TestExit(t *testing.T) {
	tests := map[string]struct {
		input error
		want  int
	}{
		"TestExitNilState": {
			input: io.EOF,
			want:  int(StatusErr),
		},

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

		"TestExitWithLockfile": {
			input: io.EOF,
			want:  int(StatusErr),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run="+name)
			cmd.Env = append(os.Environ(), "SKIP=0")

			if got, ok := cmd.Run().(*exec.ExitError); ok && got.ExitCode() != test.want {
				t.Errorf("unexpected status: %v, want %v", got.ExitCode(), test.want)
			}
		})
	}
}

func TestExitNilState(t *testing.T) {
	if os.Getenv("SKIP") != "0" {
		t.Skip()
	}

	Exit(nil, io.EOF)
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
