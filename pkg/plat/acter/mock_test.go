// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package acter

import (
	"reflect"
	"testing"

	"github.com/defended-net/malwatch/pkg/scan/state"
)

func TestLoad(t *testing.T) {
	input := &mock{}

	if got := input.Load(); got != nil {
		t.Errorf("unexpected verb result %v, want %v", got, nil)
	}
}

func TestVerb(t *testing.T) {
	var (
		want = t.Name()

		input = &mock{
			verb: want,
		}

		got = input.Verb()
	)

	if got != want {
		t.Errorf("unexpected verb result %v, want %v", got, want)
	}
}

func TestActed(t *testing.T) {
	input := &mock{}

	if err := input.Act(&state.Result{}); err != nil || input.Acted != true {
		t.Errorf("unexpected acted result %v %v, want %v", err, input.Acted, true)
	}
}

func TestMock(t *testing.T) {
	var (
		input = t.Name()

		got = Mock(input)

		want = &mock{
			verb: input,
		}
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected mock result %v, want %v", got, want)
	}
}
