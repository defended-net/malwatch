// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package info

import (
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/plat/preset"
	"github.com/defended-net/malwatch/pkg/sig"
)

func TestDo(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock error: %v", err)
	}

	if err := sig.Mock(env); err != nil {
		t.Errorf("sigs mock error: %v", err)
	}

	env.Plat = preset.New(env)

	if err := env.Plat.Load(); err != nil {
		t.Error("preset load error: ", err)
	}

	if err := Do(env, nil); err != nil {
		t.Error("print error: ", err)
	}
}
