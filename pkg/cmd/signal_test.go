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

	state := &State{
		LockPath: filepath.Join(dir, name),
		LockName: name,
		LockRoot: root,
		Signal:   make(chan os.Signal, 1),
		Cancel:   &Cancel{},
	}

	_, cancel := context.WithCancel(context.Background())

	state.AddCancel(cancel)

	return state
}

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
		t.Errorf("unexpected signal received")

	case <-time.After(time.Second * 3):
		// ok
	}
}

func TestListen(t *testing.T) {
	state := mock(t, t.TempDir(), t.Name())

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
		if _, err := os.Create(state.LockPath); err != nil {
			t.Fatalf("file create err %v", err)
		}

		if got := Listen(state); got != nil {
			t.Fatalf("listen err %v", got)
		}

		go func() {
			time.Sleep(time.Second * 3)
			state.Signal <- test.input
		}()
	}
}
