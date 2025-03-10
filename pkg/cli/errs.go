// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import "errors"

var (
	// ErrArgInvalid means invalid args.
	ErrArgInvalid = errors.New("cli: invalid arg specified")

	// ErrArgNone means no args.
	ErrArgNone = errors.New("cli: no args specified")
)
