// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package plat

import (
	"os"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
)

func TestLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	env.Cfg.Acts.Quarantine.Dir = ""

	if got := Load(env); got != nil {
		t.Errorf("plat load err %v", got)
	}
}

func TestLoadCpanel(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	var (
		src = env.Paths.Plat.Dir + "/cpanel.disabled"
		dst = env.Paths.Plat.Dir + "/cpanel.toml"
	)

	env.Cfg.Acts.Quarantine.Dir = ""

	if err := Load(env); err != nil {
		t.Errorf("plat load err %v", err)
	}

	if err := os.Rename(src, dst); err != nil {
		t.Fatalf("rename err %v", err)
	}

	if got := Load(env); got == nil {
		t.Log("cpanel load success")
	}
}

func TestLoadInstallErr(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	env.Paths.Plat.Dir = "/dev/null/not-exist"
	env.Cfg.Acts.Quarantine.Dir = ""

	if err := Load(env); err == nil {
		t.Errorf("unexpected plat load success")
	}
}
