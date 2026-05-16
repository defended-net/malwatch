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
		t.Errorf("env mock err %v", err)
	}

	if err := sig.Mock(env, true); err != nil {
		t.Errorf("sigs mock err %v", err)
	}

	env.Plat = preset.New(env)

	if env.Cfg.Secrets.S3.Endpoint == "" {
		t.Skip("s3 secret not stored")
	}

	if err := env.Plat.Load(env.Paths.Install.Root); err != nil {
		t.Error("preset load err ", err)
	}

	if got := Do(env, nil); got != nil {
		t.Error("do err ", got)
	}
}
