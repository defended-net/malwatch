// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import "errors"

var (
	// ErrInvalidChars means invalid characters.
	ErrInvalidChars = errors.New("act: invalid char input")

	// ErrUnknownVerb means unknown verb.
	ErrUnknownVerb = errors.New("act: verb not configured")

	// ErrNoActs means no acts.
	ErrNoActs = errors.New("act: none configured for")

	// ErrStarSkipNotAllowed means skip not permitted for *.
	ErrStarSkipNotAllowed = errors.New("act: skip not permitted for rule wildcard")
)
