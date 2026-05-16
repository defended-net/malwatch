// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package boot

import (
	"os"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/sig"
)

func TestRun(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := sig.Mock(env, true); err != nil {
		t.Fatalf("sig mock err %v", err)
	}

	env.Cfg.Acts.Quarantine.Dir = ""

	if got := Run(env); got != nil {
		t.Errorf("run err %v", got)
	}
}

func TestRunErrs(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := os.WriteFile(env.Paths.Cfg.Base, []byte(""), 0000); err != nil {
		t.Fatalf("file create err %v", err)
	}

	if got := Run(env); got == nil {
		t.Errorf("unexpected run success")
	}
}
