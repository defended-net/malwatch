// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package directadmin

import "errors"

var (
	// ErrAPIExec means an api bin exec error.
	ErrAPIExec = errors.New("directadmin: api bin exec error")

	// ErrAPIDomInfoUnmarshal means an domaininfo unmarshal error.
	ErrAPIDomInfoUnmarshal = errors.New("directadmin: domain info unmarshal error")
)
