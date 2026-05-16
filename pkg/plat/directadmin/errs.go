// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package directadmin

import "errors"

var (
	// ErrAPIExec means an api bin exec error.
	ErrAPIExec = errors.New("directadmin: api bin exec error")

	// ErrAPIDomInfoUnmarshal means an domaininfo unmarshal error.
	ErrAPIDomInfoUnmarshal = errors.New("directadmin: domain info unmarshal error")

	// ErrAPIAuthHost means invalid auth host.
	ErrAPIAuthHost = errors.New("directadmin: invalid auth host")

	// ErrAPIAuthScheme means invalid auth scheme.
	ErrAPIAuthScheme = errors.New("directadmin: auth scheme not https")

	// ErrAPIAuthURL means invalid auth url.
	ErrAPIAuthURL = errors.New("directadmin: invalid auth url")

	// ErrAPIAuthUser means invalid auth user.
	ErrAPIAuthUser = errors.New("directadmin: invalid auth user")

	// ErrAPIRespCode means invalid resp code.
	ErrAPIRespCode = errors.New("directadmin: invalid response code")
)
