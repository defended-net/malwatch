// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// https://tldp.org/LDP/abs/html/exitcodes.html
// https://refspecs.linuxbase.org/LSB_5.0.0/LSB-Core-generic/LSB-Core-generic/iniscrptact.html
// https://en.wikipedia.org/wiki/Exit_status#POSIX

// State represents proc state. Cancel stores goroutine cancel fns.
type State struct {
	Exit     status
	Lockfile string
	Signal   chan os.Signal `json:"-"`
	Cancel   *Cancel        `json:"-"`
}

// status represents a status code.
type status int

// Cancel represents cancel fns.
type Cancel struct {
	mtx sync.Mutex
	fns []context.CancelFunc
}

var (
	// codes stores custom error status codes.
	codes = map[error]status{
		ErrHit: StatusHit,
	}
)

// STANDARDS
const (
	StatusOK     status = 0   // StatusOK stores ok status code.
	StatusErr    status = 1   // StatusErr stores catchall error status code.
	StatusErrArg status = 2   // StatusErrArg stores arg error status code.
	StatusHup    status = 129 // StatusHup stores hangup status code.
	StatusInt    status = 130 // StatusInt stores interrupt status code.
	StatusQuit   status = 131 // StatusQuit stores quit status code.
	StatusTerm   status = 143 // StatusTerm stores terminate status code.
)

// CUSTOMS
// 166 - 199
var (
	StatusHit status = 166                    // StatusHit stores hit status code.
	ErrHit           = errors.New("cmd: hit") // ErrHit stores hit error.
)

// NewState returns a new app state.
func NewState() *State {
	return &State{
		Signal: make(chan os.Signal, 1),
		Cancel: &Cancel{},
	}
}

// Exit exits. The status code is set. All goroutine cancel fns are run for clean exit.
func Exit(state *State, err error) {
	defer func(err error) {
		var code status

		if state != nil {
			SetStatus(state, err)

			for _, cancel := range state.GetCancels() {
				cancel()
			}

			// Listen fn might have already removed the lockfile. Which is fine.
			if state.Lockfile != "" {
				if err := os.Remove(state.Lockfile); err != nil && !errors.Is(err, fs.ErrNotExist) {
					slog.Error(ErrLockDel.Error(), "path", state.Lockfile)
				}
			}

			code = state.Exit
		} else {
			code = getStatus(err)
		}

		os.Exit(int(code))
	}(err)

	if err != nil {
		slog.Error(err.Error())
	}
}

// Lock checks exclusivity to create the lockfile if available.
func (state *State) Lock(path string) error {
	state.Lockfile = path

	if _, err := os.Stat(path); err == nil {
		slog.Error(ErrLockExists.Error(), "path", path)
		os.Exit(1)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0660)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrLockCreate, err, path)
	}
	defer file.Close()

	return nil
}

// getStatus returns an error's status code.
func getStatus(err error) status {
	switch {
	case err == nil:
		return StatusOK

	case strings.HasPrefix(err.Error(), "cli:"):
		return StatusErrArg
	}

	if status, ok := codes[err]; ok {
		return status
	}

	return StatusErr
}

// SetStatus sets the status code.
func SetStatus(state *State, err error) {
	// Leave hit status code as highest priority.
	if state.Exit == StatusHit {
		return
	}

	state.Exit = getStatus(err)
}

// CancelAll runs stored cancel fns.
func (state *State) CancelAll() {
	for _, cancel := range state.GetCancels() {
		cancel()
	}
}

// Add adds a given cancel fn.
func (cancel *Cancel) Add(fn context.CancelFunc) {
	cancel.mtx.Lock()
	defer cancel.mtx.Unlock()

	cancel.fns = append(cancel.fns, fn)
}

// AddCancel adds a given cancel fn.
func (state *State) AddCancel(fn context.CancelFunc) {
	state.Cancel.Add(fn)
}

// GetCancels returns cancel fns, then clears them.
func (state *State) GetCancels() []context.CancelFunc {
	state.Cancel.mtx.Lock()
	defer state.Cancel.mtx.Unlock()

	tmp := state.Cancel.fns
	state.Cancel.fns = nil

	return tmp
}
