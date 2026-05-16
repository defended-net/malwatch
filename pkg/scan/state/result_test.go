// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

import (
	"io"
	"slices"
	"testing"
)

func TestNewResult(t *testing.T) {
	var (
		input = Paths{}
		want  = "/target"
		got   = NewResult(want, input)
	)

	switch {
	case got.Target != want:
		t.Errorf("unexpected target %v, want %v", got.Target, want)

	case got.errs == nil:
		t.Errorf("unexpected nil errs")
	}
}

func TestResultAddErr(t *testing.T) {
	var (
		input = NewResult("/target", Paths{})
		want  = io.EOF
	)

	input.AddErr(want)

	if got := input.Errs(); !slices.Equal(got, []error{want}) {
		t.Errorf("unexpected errs result %v, want %v", got, []error{want})
	}
}

func TestResultErrs(t *testing.T) {
	input := NewResult("/target", Paths{})

	if got := input.Errs(); len(got) != 0 {
		t.Errorf("unexpected errs result %v, want empty", got)
	}
}
