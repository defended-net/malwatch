// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package acter

import (
	"errors"
	"reflect"
	"testing"

	"github.com/defended-net/malwatch/pkg/scan/state"
)

func TestLoad(t *testing.T) {
	var (
		input = []Acter{
			Mock(t.Name(), true),
			Mock(t.Name()+"-disabled", false),
		}

		got, err = Load(nil, input)

		want = []Acter{
			input[0],
		}
	)

	if err != nil {
		t.Fatalf("load err %s", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected load result %v, want %v", got, want)
	}
}

func TestGet(t *testing.T) {
	var (
		input    = t.Name()
		want     = Mock(t.Name(), true)
		got, err = Get([]Acter{want}, input)
	)

	if err != nil {
		t.Fatalf("get err %s", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected get result %v, want %v", got, want)
	}
}

func TestGetNoActer(t *testing.T) {
	want := ErrVerbUnknown

	if _, got := Get([]Acter{}, t.Name()); !errors.Is(got, want) {
		t.Errorf("unexpected get result %v, want %v", got, want)
	}
}

func TestDo(t *testing.T) {
	var (
		input = []Acter{
			Mock(t.Name(), true),
		}

		result = state.NewResult("fs", state.Paths{})
	)

	if got := Do(input, t.Name(), result); got != nil {
		t.Errorf("do err %s", got)
	}
}

func TestDoNoActer(t *testing.T) {
	var (
		input  = []Acter{Mock(t.Name(), true)}
		result = state.NewResult("fs", state.Paths{})
		want   = ErrVerbUnknown
	)

	if got := Do(input, "not-exist", result); !errors.Is(got, want) {
		t.Errorf("unexpected do err %v, want %v", got, want)
	}
}
