// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package acter

import "errors"

var (
	// ErrDisabled means disabled.
	ErrDisabled = errors.New("acter: disabled")

	// ErrVerbUnknown means verb is unknown.
	ErrVerbUnknown = errors.New("act: unknown verb")
)
