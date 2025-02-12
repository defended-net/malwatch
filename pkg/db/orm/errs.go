// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package orm

import "errors"

// DB
var (
	// ErrDbNotLoaded means db not loaded.
	ErrDbNotLoaded = errors.New("db: not loaded")
)

// TX
var (
	// ErrTxBktNotFound means bucket not found.
	ErrTxBktNotFound = errors.New("orm: bucket not found")
)

// BUCKET
var (
	// ErrBktKeyNotFound means key not found.
	ErrBktKeyNotFound = errors.New("orm: key not found")

	// ErrBktGet means get error.
	ErrBktGet = errors.New("orm: get error")

	// ErrBktPut means put error.
	ErrBktPut = errors.New("orm: put error")

	// ErrBktIter means iteration error.
	ErrBktIter = errors.New("orm: iteration error")
)

// JSON
var (
	// ErrMarshal means marshal error.
	ErrMarshal = errors.New("orm: marshal error")
)
