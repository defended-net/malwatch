// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package s3

import "errors"

var (
	// ErrClientPrep means client prepare error.
	ErrClientPrep = errors.New("s3: client prepare error")

	// ErrBktLookup means bucket lookup error.
	ErrBktLookup = errors.New("s3: bucket lookup error")

	// ErrBktAdd means bucket add error.
	ErrBktAdd = errors.New("s3: bucket add error")

	// ErrObjGet means get error.
	ErrObjGet = errors.New("s3: get error")
)
