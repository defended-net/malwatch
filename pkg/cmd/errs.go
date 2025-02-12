// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import "errors"

// SIGNALS

var (
	// ErrHup means hangup signal.
	ErrHup = errors.New("cmd: received hangup signal")

	// ErrInt means interrupt signal.
	ErrInt = errors.New("cmd: received interrupt signal")

	// ErrQuit means quit signal.
	ErrQuit = errors.New("cmd: received quit signal")

	// ErrTerm means terminate signal.
	ErrTerm = errors.New("cmd: received terminate signal")
)

// LOCKFILE

var (
	// ErrLockExists means lockfile exists.
	ErrLockExists = errors.New("cmd: lockfile exists")

	// ErrLockCreate means lock err.
	ErrLockCreate = errors.New("cmd: lockfile create error")

	// ErrLockDel means lockfile del error.
	ErrLockDel = errors.New("cmd: lockfile delete error")
)
