// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package base

import "errors"

var (
	// ErrHostnameLookup means hostname lookup error.
	ErrHostnameLookup = errors.New("base: hostname lookup error")
)
