// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import "errors"

var (
	// ErrSysdMissing means missing systemd path.
	ErrSysdMissing = errors.New("install: systemd path not found")
)
