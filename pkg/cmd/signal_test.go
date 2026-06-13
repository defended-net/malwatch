// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestListenNoSig(t *testing.T) {
	state := mock(t, t.TempDir(), t.Name())

	if _, err := os.Create(state.LockPath); err != nil {
		t.Fatalf("file create err %v", err)
	}

	if got := Listen(state); got != nil {
		t.Fatalf("listen err %v", got)
	}

	select {
	case <-state.Signal:
		t.Errorf("unexpected signal")

	case <-time.After(time.Second * 3):
		// ok
	}
}

// Cases as subproc so os.Exit doesn't term. Parent verifies
// exit code.
func TestListen(t *testing.T) {
	tests := map[string]struct {
		want int
	}{
		"TestListenSigHup": {
			want: int(StatusHup),
		},

		"TestListenSigInt": {
			want: int(StatusInt),
		},

		"TestListenSigQuit": {
			want: int(StatusQuit),
		},

		"TestListenSigTerm": {
			want: int(StatusTerm),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^"+name+"$")
			cmd.Env = append(os.Environ(), "SKIP=0")

			if got, ok := cmd.Run().(*exec.ExitError); ok && got.ExitCode() != test.want {
				t.Errorf("unexpected status %v, want %v", got.ExitCode(), test.want)
			}
		})
	}
}

func TestListenSigHup(t *testing.T) {
	if os.Getenv("SKIP") != "0" {
		t.Skip()
	}

	injSig(t, syscall.SIGHUP)
}

func TestListenSigInt(t *testing.T) {
	if os.Getenv("SKIP") != "0" {
		t.Skip()
	}

	injSig(t, syscall.SIGINT)
}

func TestListenSigQuit(t *testing.T) {
	if os.Getenv("SKIP") != "0" {
		t.Skip()
	}

	injSig(t, syscall.SIGQUIT)
}

func TestListenSigTerm(t *testing.T) {
	if os.Getenv("SKIP") != "0" {
		t.Skip()
	}

	injSig(t, syscall.SIGTERM)
}

func mock(t *testing.T, dir string, name string) *State {
	t.Helper()

	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	t.Cleanup(func() {
		//lint
		_ = root.Close()
	})

	var (
		state = &State{
			LockPath: filepath.Join(dir, name),
			LockName: name,
			LockRoot: root,
			Signal:   make(chan os.Signal, 1),
			Cancel:   &Cancel{},
		}

		_, cancel = context.WithCancel(context.Background())
	)

	state.AddCancel(cancel)

	return state
}

// injSig injects req sig to proc for goroutine.
func injSig(t *testing.T, sig syscall.Signal) {
	t.Helper()

	state := mock(t, t.TempDir(), t.Name())

	if _, err := os.Create(state.LockPath); err != nil {
		t.Fatalf("file create err %v", err)
	}

	if got := Listen(state); got != nil {
		t.Fatalf("listen err %v", got)
	}

	if err := syscall.Kill(os.Getpid(), sig); err != nil {
		t.Fatalf("kill err %v", err)
	}

	// block for os.Exit
	time.Sleep(time.Second * 5)
}
