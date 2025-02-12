// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package plat

import "errors"

var (
	// ErrCfgInstall means a plat cfg install error.
	ErrCfgInstall = errors.New("plat: cfg install error")

	// ErrCfgLoad means a plat cfg load error.
	ErrCfgLoad = errors.New("plat: cfg load error")
)
