// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cfg

import "errors"

var (
	// ErrNoTargets means no scan paths were found.
	ErrNoTargets = errors.New("cfg: no scan paths found, please check cfg.toml")
)
