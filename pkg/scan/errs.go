// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package scan

import "errors"

var (
	// ErrNoScanPaths means no scan paths were found.
	ErrNoScanPaths = errors.New("scan: no scan paths in cfg.toml")

	// ErrPathGlob means glob error.
	ErrPathGlob = errors.New("fsys: path glob error")
)
