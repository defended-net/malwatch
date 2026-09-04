// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package fsys

import (
	"testing"
)

var (
	bases = []string{
		"/srv/www/site-a",
		"/srv/www/site-b",
		"/var/lib",
		"/var/log",
	}

	inputs = map[string]string{
		"hit":     "/srv/www/site-a/index.php",
		"miss":    "/srv/www/site-c/wp-content/plugins/name/class.name.php",
		"sibling": "/srv/www/site-b/wp-content/uploads/index.php",
	}
)

func BenchmarkIsRel(b *testing.B) {
	for name, input := range inputs {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_ = IsRel(input, bases...)
			}
		})
	}
}

func BenchmarkIsUnder(b *testing.B) {
	for name, input := range inputs {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_ = IsUnder(input, bases...)
			}
		})
	}
}
