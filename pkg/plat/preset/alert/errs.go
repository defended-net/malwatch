// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package alert

import "errors"

var (
	// ErrSend means alert send error.
	ErrSend = errors.New("alert: send error")
)
