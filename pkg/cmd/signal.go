// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var (
	errs = map[os.Signal]error{
		syscall.SIGHUP:  ErrHup,
		syscall.SIGINT:  ErrInt,
		syscall.SIGQUIT: ErrQuit,
		syscall.SIGTERM: ErrTerm,
	}

	signals = map[os.Signal]status{
		syscall.SIGHUP:  StatusHup,
		syscall.SIGINT:  StatusInt,
		syscall.SIGQUIT: StatusQuit,
		syscall.SIGTERM: StatusTerm,
	}
)

// Listen listens for proc ending signal events. A suitable exit code is then applied as final responsibility.
func Listen(state *State) error {
	signal.Notify(
		state.Signal,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	go func() {
		trapped := <-state.Signal
		signal.Stop(state.Signal)

		state.CancelAll()

		if err, ok := errs[trapped]; ok && err != nil {
			slog.Error(err.Error())
		}

		// Deferred Exit fn might have already removed the lockfile. Which is fine.
		if state.LockRoot != nil && state.LockName != "" {
			if err := state.LockRoot.Remove(state.LockName); err != nil {
				slog.Error(ErrLockDel.Error(), "path", state.LockPath, "msg", err)
			}
		}

		os.Exit(int(signals[trapped]))
	}()

	return nil
}
