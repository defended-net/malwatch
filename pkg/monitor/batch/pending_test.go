// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package batch

import (
	"slices"
	"testing"
)

func TestPendingGet(t *testing.T) {
	var (
		input = NewPending()
		want  = []string{t.TempDir()}
	)

	input.Add(want[0])

	got := input.Get()

	if !slices.Equal(got, want) {
		t.Errorf("unexpected batch get result %v, want %v", got, want)
	}
}

func TestPendingAdd(t *testing.T) {
	var (
		input = NewPending()
		want  = []string{t.TempDir()}
	)

	input.Add(want[0])

	got := input.Get()

	if !slices.Equal(got, want) {
		t.Errorf("unexpected batch add result %v, want %v", got, want)
	}
}
