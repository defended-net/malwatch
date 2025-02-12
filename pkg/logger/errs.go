// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package logger

import "errors"

var (
	// ErrOpen means log file open error.
	ErrOpen = errors.New("logger: log file open error")
)
