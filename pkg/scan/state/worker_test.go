// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

import (
	"testing"
)

func TestGc(t *testing.T) {
	var (
		called = false

		input = &Scanner{
			GcFn: func() { called = true },
		}
	)

	input.Gc()

	switch {
	case !called:
		t.Errorf("gc fn not called")

	case input.GcFn != nil:
		t.Errorf("gc fn not cleared")
	}

	// Second call must be noop.
	input.Gc()
}

func TestGcEmpty(_ *testing.T) {
	input := &Scanner{}

	// Safe as nil val with nil gcfn.
	input.Gc()
}
