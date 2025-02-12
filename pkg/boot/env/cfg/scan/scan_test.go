// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package scan

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	got := New()

	want := &Cfg{
		Targets: []string{
			`^/var/www/(?P<target>[^/]+)`,
		},

		Paths: []string{
			"/var/www",
		},

		Timeout: 60,
		MaxAge:  0,

		BlkSz:   65536,
		BatchSz: 500,

		Monitor: &Monitor{
			Timeout: 5,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected new cfg result %v, want %v", got, true)
	}
}
