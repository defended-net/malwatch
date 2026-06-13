// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"testing"
)

func TestAcquireNil(t *testing.T) {
	curr.Store(nil)

	if _, got := Acquire(); got == nil {
		t.Errorf("unexpected acquire success")
	}
}

func TestSetErr(t *testing.T) {
	if got := Set("/dev/null/not-exist", 0); got == nil {
		t.Errorf("unexpected set success")
	}
}
