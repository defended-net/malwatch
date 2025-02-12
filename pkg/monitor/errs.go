// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitor

import "errors"

var (
	// ErrUnknownMnt means an unknown mountpoint.
	ErrUnknownMnt = errors.New("monitor: unknown mount")

	// ErrUnknownEvent means an unknown event.
	ErrUnknownEvent = errors.New("monitor: unknown event")

	// ErrNotifierMark means a notifier mark error.
	ErrNotifierMark = errors.New("monitor: notifier mark error")

	// ErrMetaClose means a meta close error.
	ErrMetaClose = errors.New("monitor: meta close error")
)
