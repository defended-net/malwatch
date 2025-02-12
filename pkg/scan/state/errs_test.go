// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

import (
	"io"
	"slices"
	"testing"
)

var (
	single = []error{
		io.EOF,
	}

	compound = []error{
		io.EOF,
		io.ErrUnexpectedEOF,
	}
)

func TestAddSingle(t *testing.T) {
	var (
		input = &Errs{}
		want  = single
	)

	input.Add(want[0])

	if !slices.Equal(input.Vals, want) {
		t.Errorf("unexpected single add err result %v, want %v", input.Vals, want)
	}
}

func TestAddCompound(t *testing.T) {
	input := &Errs{}

	want := compound

	for _, err := range want {
		input.Add(err)
	}

	got := input.Vals

	if !slices.Equal(got, want) {
		t.Errorf("unexpected compound add err result %v, want %v", got, want)
	}
}

func TestGet(t *testing.T) {
	input := &Errs{
		Vals: compound,
	}

	result := input.Get()

	if !slices.Equal(result, compound) {
		t.Errorf("unexpected get result %v, want %v", result, compound)
	}
}
