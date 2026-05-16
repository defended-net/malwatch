// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
)

func TestPreset(t *testing.T) {
	want := 4

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if got := Preset(env); len(got) != want {
		t.Errorf("unexpected preset count %v, want %v", len(got), want)
	}
}
