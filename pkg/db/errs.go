// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import "errors"

var (
	// ErrOpen means db open error.
	ErrOpen = errors.New("db: db open error")

	// ErrBktCreate means bucket create error.
	ErrBktCreate = errors.New("db: bucket create error")

	// ErrTxUpdate means update error.
	ErrTxUpdate = errors.New("db: update error")
)
