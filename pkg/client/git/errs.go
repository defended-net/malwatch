// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package git

import "errors"

var (
	// ErrClone means a clone error.
	ErrClone = errors.New("git: clone error")

	// ErrRefTag means a tag lookup error.
	ErrRefTag = errors.New("git: tag lookup error")

	// ErrRefIter means a ref iteration error.
	ErrRefIter = errors.New("git: ref iteration error")
)
