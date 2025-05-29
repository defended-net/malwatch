// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package exec

import "errors"

var (
	// ErrRun means run error.
	ErrRun = errors.New("exec: run error")

	// ErrMetaChars means meta chars detected.
	ErrMetaChars = errors.New("exec: detected meta chars")
)
