// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"github.com/defended-net/malwatch/third_party/yr"
)

// HasMatch verifies if matches contains given rule.
func HasMatch(matches *yr.MatchRules, rule string) bool {
	for _, match := range *matches {
		if match.Rule == rule {
			return true
		}
	}

	return false
}
