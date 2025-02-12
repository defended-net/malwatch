// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cpanel

import "errors"

var (
	// ErrAPIToolExec means an apitool exec error.
	ErrAPIToolExec = errors.New("cpanel: apitool exec error")

	// ErrAPIDomInfoUnmarshal means an domaininfo unmarshal error.
	ErrAPIDomInfoUnmarshal = errors.New("cpanel: domaininfo unmarshal error")
)
