// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package restore

import "errors"

var (
	// ErrStatusUnknown means unknown status.
	ErrStatusUnknown = errors.New("restore: unknown status")
)
