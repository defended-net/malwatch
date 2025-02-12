// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cpanel

import "regexp"

var (
	// reTarget validates an account's home path.
	reTarget = regexp.MustCompile("^/home/(?P<target>[A-Za-z0-9]{1,16}).*/?$")

	// TODO
	// reUser is used to validate an account's username.
	// https://docs.cpanel.net/knowledge-base/accounts/reserved-invalid-and-misconfigured-username/
	// reUser = regexp.MustCompile("^[A-Za-z0-9]{1,16}$")
)
