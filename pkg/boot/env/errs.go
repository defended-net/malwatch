// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package env

import "errors"

var (
	// ErrSelfBin means executable path is unknown.
	ErrSelfBin = errors.New("env: executable path unknown")
)
