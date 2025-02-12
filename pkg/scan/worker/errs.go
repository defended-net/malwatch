// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package worker

import "errors"

var (
	// ErrFileRead means file read error.
	ErrFileRead = errors.New("worker: file read error")

	// ErrYrScan means yr scan error.
	ErrYrScan = errors.New("worker: yr scan error")
)
