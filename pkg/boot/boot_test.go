// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package boot

import (
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
)

func TestRun(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	env.Cfg.Acts.Quarantine.Dir = ""

	if err = Run(env); err != nil {
		t.Errorf("boot error: %v", err)
	}
}
