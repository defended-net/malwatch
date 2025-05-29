// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package exec

import "regexp"

var (
	// reMetaChars validates shell proc args against unsafe meta chars.
	reMetaChars = regexp.MustCompile(`[;&|$\\(\\)<>]`)
)
