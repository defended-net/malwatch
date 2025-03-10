// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package acter

import (
	"errors"
	"reflect"
	"testing"

	"github.com/defended-net/malwatch/pkg/scan/state"
)

func TestGet(t *testing.T) {
	var (
		input = t.Name()
		want  = Mock(t.Name())
	)

	got, err := Get([]Acter{want}, input)
	if err != nil {
		t.Fatalf("get error: %s", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected get result %v, want %v", got, want)
	}
}

func TestGetNoActer(t *testing.T) {
	if _, got := Get([]Acter{}, t.Name()); !errors.Is(got, ErrVerbUnknown) {
		t.Errorf("unexpected get no acter result %v, want %v", got, ErrVerbUnknown)
	}
}

func TestDo(t *testing.T) {
	input := []Acter{Mock(t.Name())}

	if err := Do(input, t.Name(), state.NewResult("fs", state.Paths{})); err != nil {
		t.Errorf("do error: %s", err)
	}
}

func TestDoNoActer(t *testing.T) {
	var (
		input = []Acter{Mock(t.Name())}
		want  = ErrVerbUnknown
	)

	if got := Do(input, "not-exist", state.NewResult("fs", state.Paths{})); !errors.Is(got, want) {
		t.Errorf("unexpected do err %v, want %v", got, want)
	}
}
