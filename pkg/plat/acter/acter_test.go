// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package acter

import (
	"errors"
	"reflect"
	"testing"
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
		t.Errorf("unexpected get acter result %v, want %v", got, want)
	}
}

func TestGetNoActer(t *testing.T) {
	if _, got := Get([]Acter{}, t.Name()); !errors.Is(got, ErrVerbUnknown) {
		t.Errorf("unexpected get no acter result %v, want %v", got, ErrVerbUnknown)
	}
}
