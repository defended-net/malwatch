// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func mock(dir string, name string) *State {
	state := &State{
		Lockfile: filepath.Join(dir, name),
		Signal:   make(chan os.Signal, 1),
		Cancel:   &Cancel{},
	}

	_, cancel := context.WithCancel(context.Background())

	state.AddCancel(cancel)

	return state
}

func TestListenNoSig(t *testing.T) {
	state := mock(t.TempDir(), t.Name())

	if _, err := os.Create(state.Lockfile); err != nil {
		t.Fatalf("lockfile create error: %v", err)
	}

	if err := Listen(state); err != nil {
		t.Fatalf("listen error: %v", err)
	}

	select {
	case <-state.Signal:
		t.Errorf("unexpected signal received")

	case <-time.After(time.Second * 3):
		// Expected.
	}
}

func TestListen(t *testing.T) {
	state := mock(t.TempDir(), t.Name())

	_, cancel := context.WithCancel(context.Background())
	state.AddCancel(cancel)

	tests := map[string]struct {
		input os.Signal
	}{
		"sighup": {
			input: syscall.SIGHUP,
		},

		"sigint": {
			input: syscall.SIGINT,
		},

		"sigquit": {
			input: syscall.SIGQUIT,
		},

		"sigterm": {
			input: syscall.SIGTERM,
		},

		"interrupt": {
			input: os.Interrupt,
		},
	}

	for _, test := range tests {
		if _, err := os.Create(state.Lockfile); err != nil {
			t.Fatalf("lockfile create error: %v", err)
		}

		if err := Listen(state); err != nil {
			t.Fatalf("listen error: %v", err)
		}

		go func() {
			time.Sleep(time.Second * 3)
			state.Signal <- test.input
		}()
	}
}
