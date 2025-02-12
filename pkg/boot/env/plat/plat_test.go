// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package plat

import (
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
)

func TestLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err = Load(env); err != nil {
		t.Errorf("plat load error: %v", err)
	}
}
