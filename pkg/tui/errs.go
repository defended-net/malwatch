// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package tui

import "errors"

var (
	// ErrYesNoInvalid means input does not match Y/y or N/n.
	ErrYesNoInvalid = errors.New("tui: invalid input, please y or n")
)
