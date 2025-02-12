// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package re

import (
	"regexp"
)

var (
	targets []*regexp.Regexp
	yrName  = regexp.MustCompile(`(^[\w*]+$)`)
)

// SetTargets sets the re for targets.
func SetTargets(re ...*regexp.Regexp) {
	targets = append(targets, re...)
}

// IsValidYrName validates a yr rule name.
func IsValidYrName(name string) bool {
	return yrName.MatchString(name)
}

// FindYrName finds a yr rule name.
func FindYrName(name string) string {
	return yrName.FindString(name)
}

// Target returns a path's target.
func Target(path string) string {
	for _, target := range targets {
		matches := target.FindStringSubmatch(path)

		if len(matches) < 1 {
			continue
		}

		return matches[1]
	}

	return "fs"
}
